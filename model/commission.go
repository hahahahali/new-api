package model

import (
	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	CommissionStatusPending = "pending"
	CommissionStatusSettled = "settled"
)

type CommissionRecord struct {
	Id               int     `json:"id"`
	InviterId        int     `json:"inviter_id" gorm:"index"`
	InviteeId        int     `json:"invitee_id" gorm:"index"`
	InvitationCode   string  `json:"invitation_code" gorm:"type:varchar(32)"`
	TradeNo          string  `json:"trade_no" gorm:"type:varchar(255);uniqueIndex"`
	RechargeAmount   float64 `json:"recharge_amount" gorm:"type:decimal(10,6);not null;default:0"`
	CommissionRate   float64 `json:"commission_rate" gorm:"type:decimal(5,4);not null;default:0"`
	CommissionAmount float64 `json:"commission_amount" gorm:"type:decimal(10,6);not null;default:0"`
	Status           string  `json:"status" gorm:"type:varchar(16);default:'pending';index"`
	AvailableAt      int64   `json:"available_at" gorm:"bigint;default:0"`
	CreatedAt        int64   `json:"created_at" gorm:"bigint"`
	UpdatedAt        int64   `json:"updated_at" gorm:"bigint"`
}

func (r *CommissionRecord) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	r.CreatedAt = now
	r.UpdatedAt = now
	return nil
}

func (r *CommissionRecord) BeforeUpdate(tx *gorm.DB) error {
	r.UpdatedAt = common.GetTimestamp()
	return nil
}

func InsertCommissionRecord(record *CommissionRecord) error {
	return DB.Create(record).Error
}

func GetCommissionRecordsByInviterId(inviterId int, page, pageSize int) ([]CommissionRecord, int64, error) {
	var records []CommissionRecord
	var total int64

	tx := DB.Model(&CommissionRecord{}).Where("inviter_id = ?", inviterId)
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := tx.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&records).Error; err != nil {
		return nil, 0, err
	}
	return records, total, nil
}

func GetCommissionSummaryByInviterId(inviterId int) (totalCommission float64, totalRecharge float64, count int64, err error) {
	var result struct {
		TotalCommission float64
		TotalRecharge   float64
		Count           int64
	}
	err = DB.Model(&CommissionRecord{}).
		Select("COALESCE(SUM(commission_amount), 0) as total_commission, COALESCE(SUM(recharge_amount), 0) as total_recharge, COUNT(*) as count").
		Where("inviter_id = ?", inviterId).
		Scan(&result).Error
	return result.TotalCommission, result.TotalRecharge, result.Count, err
}

// KolInvitee represents an invited user with their total recharge amount
type KolInvitee struct {
	Id              int     `json:"id"`
	Username        string  `json:"username"`
	Email           string  `json:"email"`
	CreatedAt       int64   `json:"created_at"`
	TotalRecharge   float64 `json:"total_recharge"`
	TotalCommission float64 `json:"total_commission"`
}

// GetPendingCommissionAmountByInviterId returns the total frozen (pending) commission
// amount for a KOL that has not yet become withdrawable.
func GetPendingCommissionAmountByInviterId(inviterId int) (float64, error) {
	var result struct{ Total float64 }
	err := DB.Model(&CommissionRecord{}).
		Select("COALESCE(SUM(commission_amount), 0) as total").
		Where("inviter_id = ? AND status = ?", inviterId, CommissionStatusPending).
		Scan(&result).Error
	return result.Total, err
}

// GetKolInvitees returns the list of users invited by a KOL with their recharge totals
func GetKolInvitees(inviterId int, page, pageSize int) ([]KolInvitee, int64, error) {
	var total int64
	if err := DB.Model(&User{}).Where("inviter_id = ?", inviterId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if total == 0 {
		return []KolInvitee{}, 0, nil
	}

	var users []User
	if err := DB.Where("inviter_id = ?", inviterId).
		Select("id, username, email").
		Order("id desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&users).Error; err != nil {
		return nil, 0, err
	}

	result := make([]KolInvitee, len(users))
	for i, u := range users {
		result[i] = KolInvitee{
			Id:       u.Id,
			Username: u.Username,
			Email:    u.Email,
		}
		// Get recharge and commission totals for this invitee
		var summary struct {
			TotalRecharge   float64
			TotalCommission float64
		}
		DB.Model(&CommissionRecord{}).
			Select("COALESCE(SUM(recharge_amount), 0) as total_recharge, COALESCE(SUM(commission_amount), 0) as total_commission").
			Where("inviter_id = ? AND invitee_id = ?", inviterId, u.Id).
			Scan(&summary)
		result[i].TotalRecharge = summary.TotalRecharge
		result[i].TotalCommission = summary.TotalCommission
	}

	return result, total, nil
}
