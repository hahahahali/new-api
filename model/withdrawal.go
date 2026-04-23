package model

import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

type WithdrawalRequest struct {
	Id     int     `json:"id"`
	UserId int     `json:"user_id" gorm:"index"`
	Amount float64 `json:"amount" gorm:"type:decimal(10,6);not null;default:0"`
	Status string  `json:"status" gorm:"type:varchar(16);default:'pending';index"`
	// StripeTransferId string `json:"stripe_transfer_id" gorm:"type:varchar(128);default:''"` // [Stripe Connect - disabled]
	PaypalEmail         string `json:"paypal_email" gorm:"type:varchar(128);default:''"`
	PaypalName          string `json:"paypal_name" gorm:"type:varchar(128);default:''"`
	PaypalTransactionId string `json:"paypal_transaction_id" gorm:"type:varchar(128);default:''"`
	RejectReason        string `json:"reject_reason" gorm:"type:text"`
	CreatedAt           int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt           int64  `json:"updated_at" gorm:"bigint"`
}

// WithdrawalAdminView extends WithdrawalRequest with joined user info for admin display.
type WithdrawalAdminView struct {
	WithdrawalRequest
	UserEmail string `json:"user_email" gorm:"column:user_email"`
	Username  string `json:"username" gorm:"column:username"`
}

const (
	WithdrawalStatusPending  = "pending"
	WithdrawalStatusApproved = "approved"
	WithdrawalStatusRejected = "rejected"
	WithdrawalStatusPaid     = "paid"
)

func (w *WithdrawalRequest) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	w.CreatedAt = now
	w.UpdatedAt = now
	return nil
}

func (w *WithdrawalRequest) BeforeUpdate(tx *gorm.DB) error {
	w.UpdatedAt = common.GetTimestamp()
	return nil
}

func InsertWithdrawalRequest(req *WithdrawalRequest) error {
	return DB.Create(req).Error
}

func GetWithdrawalRequestById(id int) (*WithdrawalRequest, error) {
	var req WithdrawalRequest
	if err := DB.Where("id = ?", id).First(&req).Error; err != nil {
		return nil, err
	}
	return &req, nil
}

func GetWithdrawalRequestsByUserId(userId int, page, pageSize int) ([]WithdrawalRequest, int64, error) {
	var requests []WithdrawalRequest
	var total int64

	tx := DB.Model(&WithdrawalRequest{}).Where("user_id = ?", userId)
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

func GetAllWithdrawalRequests(status string, page, pageSize int) ([]WithdrawalRequest, int64, error) {
	var requests []WithdrawalRequest
	var total int64

	tx := DB.Model(&WithdrawalRequest{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&requests).Error; err != nil {
		return nil, 0, err
	}
	return requests, total, nil
}

// GetAllWithdrawalRequestsAdmin returns withdrawal requests joined with user info for admin management.
func GetAllWithdrawalRequestsAdmin(status string, page, pageSize int) ([]WithdrawalAdminView, int64, error) {
	var items []WithdrawalAdminView
	var total int64

	base := DB.Model(&WithdrawalRequest{}).
		Joins("LEFT JOIN users ON users.id = withdrawal_requests.user_id")

	if status != "" {
		base = base.Where("withdrawal_requests.status = ?", status)
	}

	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := base.
		Select("withdrawal_requests.*, users.email as user_email, users.username as username").
		Order("withdrawal_requests.created_at desc").
		Offset((page-1)*pageSize).
		Limit(pageSize).
		Find(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, total, nil
}

// CreateWithdrawalWithDeduction atomically deducts KOL balance and creates a pending withdrawal.
func CreateWithdrawalWithDeduction(userId int, amount float64, paypalEmail, paypalName string) (*WithdrawalRequest, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be > 0")
	}

	var req *WithdrawalRequest

	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ?", userId).First(&user).Error; err != nil {
			return err
		}
		if user.KolBalance < amount {
			return errors.New("余额不足")
		}

		user.KolBalance -= amount
		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		req = &WithdrawalRequest{
			UserId:      userId,
			Amount:      amount,
			Status:      WithdrawalStatusPending,
			PaypalEmail: paypalEmail,
			PaypalName:  paypalName,
		}
		return tx.Create(req).Error
	})

	return req, err
}

func ApproveWithdrawal(id int) (bool, error) {
	result := DB.Model(&WithdrawalRequest{}).Where("id = ? AND status = ?", id, WithdrawalStatusPending).
		Updates(map[string]interface{}{
			"status":     WithdrawalStatusApproved,
			"updated_at": common.GetTimestamp(),
		})
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, nil
	}
	return true, nil
}

func RejectWithdrawal(id int, reason string) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var req WithdrawalRequest
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Where("id = ? AND status = ?", id, WithdrawalStatusPending).First(&req).Error; err != nil {
			return err
		}

		req.Status = WithdrawalStatusRejected
		req.RejectReason = reason
		req.UpdatedAt = common.GetTimestamp()
		if err := tx.Save(&req).Error; err != nil {
			return err
		}

		// Refund balance
		return tx.Model(&User{}).Where("id = ?", req.UserId).
			Updates(map[string]interface{}{
				"kol_balance": gorm.Expr("kol_balance + ?", req.Amount),
			}).Error
	})
}

// RevertWithdrawalToPending reverts an approved withdrawal back to pending.
// [Stripe Connect - disabled] kept for reference
func RevertWithdrawalToPending(id int) error {
	return DB.Model(&WithdrawalRequest{}).Where("id = ? AND status = ?", id, WithdrawalStatusApproved).
		Updates(map[string]interface{}{
			"status":     WithdrawalStatusPending,
			"updated_at": common.GetTimestamp(),
		}).Error
}

// MarkWithdrawalPaid marks a pending withdrawal as paid with the given PayPal transaction ID.
func MarkWithdrawalPaid(id int, paypalTransactionId string) error {
	result := DB.Model(&WithdrawalRequest{}).
		Where("id = ? AND status IN ?", id, []string{WithdrawalStatusPending, WithdrawalStatusApproved}).
		Updates(map[string]interface{}{
			"status":                WithdrawalStatusPaid,
			"paypal_transaction_id": paypalTransactionId,
			"updated_at":            common.GetTimestamp(),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var req WithdrawalRequest
		if err := DB.Select("id", "status").Where("id = ?", id).First(&req).Error; err != nil {
			return err
		}
		if req.Status == WithdrawalStatusPaid {
			return nil // already paid
		}
		return fmt.Errorf("withdrawal %d has status %s, cannot mark as paid", id, req.Status)
	}
	return nil
}
