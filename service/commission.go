package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"

	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
)

const (
	// commissionSettleTickInterval is how often the settle task runs.
	// 10 minutes is sufficient — the freeze window is 10 days, so
	// exact-to-the-minute precision is not required.
	commissionSettleTickInterval = 10 * time.Minute

	// commissionFreezeSeconds is the duration (in seconds) that a commission
	// record is frozen before it becomes withdrawable: 10 days.
	commissionFreezeSeconds = int64(10 * 24 * 3600)

	// kolCommissionEligibleOrderLimit is the maximum number of commissionable
	// recharge orders per invitee.
	kolCommissionEligibleOrderLimit = int64(3)
)

var (
	commissionSettleOnce    sync.Once
	commissionSettleRunning atomic.Bool
)

// ProcessCommission creates a pending commission record for the KOL who invited
// inviteeUserId. tradeNo must be the payment order's trade_no (unique per payment).
// paidUSD is the actual amount paid by the invitee (after any rebate discount).
//
// The pre-discount "original" price is looked up from the TopUp record by tradeNo.
// Keeping the signature at 3 params minimises the merge surface on upstream topup
// controllers — each caller needs only a plain 3-arg call, regardless of whether
// their payment channel supports KOL rebate.
//
// Commission = originalUSD × commissionRate − (originalUSD − paidUSD)
// i.e. the KOL earns their percentage on the full price, minus the discount they gave away.
//
// Only the first 3 orders from each invitee trigger a commission. Subsequent orders
// are ignored silently.
//
// Idempotent: duplicate trade_no values are silently ignored via unique index.
func ProcessCommission(inviteeUserId int, tradeNo string, paidUSD float64) error {
	rate := setting.KolCommissionRate
	if rate <= 0 || paidUSD <= 0 {
		return nil
	}

	// originalUSD defaults to paidUSD (no discount channel).
	// Stripe writes OriginalMoney onto the TopUp record before this hook fires,
	// so we pick it up automatically without requiring each caller to pass it.
	originalUSD := paidUSD
	var topUp model.TopUp
	if err := model.DB.Select("original_money").
		Where("trade_no = ?", tradeNo).First(&topUp).Error; err == nil {
		if topUp.OriginalMoney > 0 {
			originalUSD = topUp.OriginalMoney
		}
	}

	now := common.GetTimestamp()
	// commission = originalUSD × rate − discount_given_to_invitee
	discount := originalUSD - paidUSD
	if discount < 0 {
		discount = 0
	}
	commissionAmount := originalUSD*rate - discount
	if commissionAmount <= 0 {
		return nil // KOL gave all commission as rebate; no record needed
	}

	groupCol := "`group`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
	}

	return model.DB.Transaction(func(tx *gorm.DB) error {
		// Lock the invitee row so concurrent payments from the same invitee cannot
		// race past the "first 3 orders" limit.
		var invitee model.User
		if err := tx.Set("gorm:query_option", "FOR UPDATE").
			Select("id, inviter_id").Where("id = ?", inviteeUserId).First(&invitee).Error; err != nil {
			return nil // user not found — no commission
		}
		if invitee.InviterId <= 0 {
			return nil
		}

		var inviter model.User
		if err := tx.Select("id, aff_code, "+groupCol).
			Where("id = ?", invitee.InviterId).First(&inviter).Error; err != nil {
			return nil
		}
		if inviter.Group != "kol" {
			return nil
		}

		var existingCount int64
		if err := tx.Model(&model.CommissionRecord{}).
			Where("inviter_id = ? AND invitee_id = ?", inviter.Id, inviteeUserId).
			Count(&existingCount).Error; err != nil {
			return fmt.Errorf("ProcessCommission: count existing records failed: %w", err)
		}
		if existingCount >= 3 {
			return nil
		}

		record := &model.CommissionRecord{
			InviterId:        inviter.Id,
			InviteeId:        inviteeUserId,
			InvitationCode:   inviter.AffCode,
			TradeNo:          tradeNo,
			RechargeAmount:   originalUSD,
			CommissionRate:   rate,
			CommissionAmount: commissionAmount,
			Status:           model.CommissionStatusPending,
			AvailableAt:      now + commissionFreezeSeconds,
		}
		if err := tx.Create(record).Error; err != nil {
			// Duplicate trade_no means this payment has already been processed — not an error.
			if isDuplicateKeyError(err) {
				return nil
			}
			return fmt.Errorf("ProcessCommission: insert record failed for trade_no=%s: %w", tradeNo, err)
		}
		return nil
	})
}

// StartCommissionSettleTask starts the periodic task that promotes pending commission
// records to settled and credits the KOL's kol_balance once the 10-day freeze passes.
func StartCommissionSettleTask() {
	commissionSettleOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(),
				"commission settle task started: tick=10m, mode=recharge-based, freeze=10days",
			)
			ticker := time.NewTicker(commissionSettleTickInterval)
			defer ticker.Stop()
			for range ticker.C {
				runCommissionSettleOnce()
			}
		})
	})
}

func runCommissionSettleOnce() {
	if !commissionSettleRunning.CompareAndSwap(false, true) {
		return
	}
	defer commissionSettleRunning.Store(false)

	now := common.GetTimestamp()
	ctx := context.Background()

	var records []model.CommissionRecord
	if err := model.DB.
		Where("status = ? AND available_at <= ?", model.CommissionStatusPending, now).
		Find(&records).Error; err != nil {
		logger.LogWarn(ctx, fmt.Sprintf("commission settle: query pending records failed: %v", err))
		return
	}
	if len(records) == 0 {
		return
	}

	settled := 0
	for _, rec := range records {
		err := model.DB.Transaction(func(tx *gorm.DB) error {
			// Re-fetch with status check inside the transaction to prevent double-credit.
			var current model.CommissionRecord
			if err := tx.Where("id = ? AND status = ?", rec.Id, model.CommissionStatusPending).
				First(&current).Error; err != nil {
				// Already settled by a concurrent tick — skip silently.
				return nil
			}

			if err := tx.Model(&model.CommissionRecord{}).Where("id = ?", rec.Id).
				Updates(map[string]interface{}{
					"status":     model.CommissionStatusSettled,
					"updated_at": common.GetTimestamp(),
				}).Error; err != nil {
				return err
			}

			return tx.Model(&model.User{}).Where("id = ?", rec.InviterId).
				Updates(map[string]interface{}{
					"kol_balance":         gorm.Expr("kol_balance + ?", rec.CommissionAmount),
					"kol_history_balance": gorm.Expr("kol_history_balance + ?", rec.CommissionAmount),
				}).Error
		})
		if err != nil {
			logger.LogError(ctx, fmt.Sprintf(
				"commission settle: failed to settle record id=%d inviter=%d: %v",
				rec.Id, rec.InviterId, err,
			))
			// Continue — other records are independent; this one retries next tick.
			continue
		}
		settled++
	}

	if settled > 0 {
		logger.LogInfo(ctx, fmt.Sprintf("commission settle done: settled=%d", settled))
	}
}

// isDuplicateKeyError reports whether err is a unique-constraint violation
// across MySQL, PostgreSQL, and SQLite.
func isDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "1062") ||
		strings.Contains(msg, "Duplicate entry") ||
		strings.Contains(msg, "duplicate key") ||
		strings.Contains(msg, "UNIQUE constraint failed")
}

func getKolRebateForUserTx(tx *gorm.DB, userId int, lockUser bool) (float64, error) {
	commissionRate := setting.KolCommissionRate
	if commissionRate <= 0 {
		return 0, nil
	}

	var user model.User
	query := tx.Select("id, inviter_id")
	if lockUser {
		query = query.Set("gorm:query_option", "FOR UPDATE")
	}
	if err := query.Where("id = ?", userId).First(&user).Error; err != nil {
		return 0, err
	}
	if user.InviterId <= 0 {
		return 0, nil
	}

	groupCol := "`group`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
	}
	var inviter model.User
	if err := tx.Select("id, "+groupCol+", kol_rebate_rate").
		Where("id = ?", user.InviterId).First(&inviter).Error; err != nil {
		return 0, err
	}
	if inviter.Group != "kol" || inviter.KolRebateRate <= 0 {
		return 0, nil
	}

	var completedCount int64
	if err := tx.Model(&model.CommissionRecord{}).
		Where("inviter_id = ? AND invitee_id = ?", inviter.Id, userId).
		Count(&completedCount).Error; err != nil {
		return 0, err
	}

	var pendingCount int64
	if err := tx.Model(&model.TopUp{}).
		Where("user_id = ? AND status = ?", userId, common.TopUpStatusPending).
		Count(&pendingCount).Error; err != nil {
		return 0, err
	}

	if completedCount+pendingCount >= kolCommissionEligibleOrderLimit {
		return 0, nil
	}

	// Clamp rebate to current commission rate (guards against rate decreases)
	rebate := inviter.KolRebateRate
	if rebate > commissionRate {
		rebate = commissionRate
	}
	return rebate, nil
}

// GetKolRebateForUser returns the rebate rate that applies to the next purchase by userId.
// Returns 0 if: the user has no KOL inviter, the inviter's rebate rate is 0,
// or the invitee has already exhausted the first 3 commissionable orders.
func GetKolRebateForUser(userId int) float64 {
	rebate, err := getKolRebateForUserTx(model.DB, userId, false)
	if err != nil {
		return 0
	}
	return rebate
}

// CreatePendingTopUpWithKolRebate atomically reserves one commissionable order slot
// (via a pending top-up row) and returns the discounted payment amount.
func CreatePendingTopUpWithKolRebate(userId int, amount int64, originalMoney float64, tradeNo, paymentMethod string) (*model.TopUp, float64, error) {
	if originalMoney <= 0 {
		return nil, 0, fmt.Errorf("invalid original money %.4f", originalMoney)
	}

	var (
		topUp      *model.TopUp
		rebateRate float64
	)

	err := model.DB.Transaction(func(tx *gorm.DB) error {
		rate, err := getKolRebateForUserTx(tx, userId, true)
		if err != nil {
			return err
		}

		payMoney := originalMoney * (1 - rate)
		if payMoney < 0.01 {
			return fmt.Errorf("discounted payment amount too low: %.4f", payMoney)
		}

		topUp = &model.TopUp{
			UserId:        userId,
			Amount:        amount,
			Money:         payMoney,
			OriginalMoney: originalMoney,
			TradeNo:       tradeNo,
			PaymentMethod: paymentMethod,
			CreateTime:    time.Now().Unix(),
			Status:        common.TopUpStatusPending,
		}
		if err := tx.Create(topUp).Error; err != nil {
			return err
		}

		rebateRate = rate
		return nil
	})
	if err != nil {
		return nil, 0, err
	}
	return topUp, rebateRate, nil
}
