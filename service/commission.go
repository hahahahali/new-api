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
	// 10 minutes is sufficient — the freeze window is 3 days, so
	// exact-to-the-minute precision is not required.
	commissionSettleTickInterval = 10 * time.Minute

	// commissionFreezeSeconds is the duration (in seconds) that a commission
	// record is frozen before it becomes withdrawable: 3 days.
	commissionFreezeSeconds = int64(3 * 24 * 3600)
)

var (
	commissionSettleOnce    sync.Once
	commissionSettleRunning atomic.Bool
)

// ProcessCommission creates a pending commission record for the KOL who invited
// inviteeUserId. tradeNo must be the payment order's trade_no (unique per payment).
// rechargeUSD is the USD face value of credits the invitee just purchased
// (topUp.Money, which equals the recharge amount when no discounts are applied).
// The commission becomes available for withdrawal after commissionFreezeSeconds (3 days).
//
// Note: if top-up discounts are introduced in the future, rechargeUSD should be
// reconsidered — pass the actual USD charged rather than the face value if
// commission should be based on real payment amount.
//
// Idempotent: duplicate trade_no values are silently ignored via unique index.
func ProcessCommission(inviteeUserId int, tradeNo string, rechargeUSD float64) error {
	rate := setting.KolCommissionRate
	if rate <= 0 || rechargeUSD <= 0 {
		return nil
	}

	// Resolve inviter
	var invitee model.User
	if err := model.DB.Select("id, inviter_id").Where("id = ?", inviteeUserId).First(&invitee).Error; err != nil {
		return nil // user not found — no commission
	}
	if invitee.InviterId <= 0 {
		return nil
	}

	// Verify inviter is KOL
	groupCol := "`group`"
	if common.UsingPostgreSQL {
		groupCol = `"group"`
	}
	var inviter model.User
	if err := model.DB.Select("id, aff_code, "+groupCol).
		Where("id = ?", invitee.InviterId).First(&inviter).Error; err != nil {
		return nil
	}
	if inviter.Group != "kol" {
		return nil
	}

	now := common.GetTimestamp()
	commissionAmount := rechargeUSD * rate

	record := &model.CommissionRecord{
		InviterId:        inviter.Id,
		InviteeId:        inviteeUserId,
		InvitationCode:   inviter.AffCode,
		TradeNo:          tradeNo,
		RechargeAmount:   rechargeUSD,
		CommissionRate:   rate,
		CommissionAmount: commissionAmount,
		Status:           model.CommissionStatusPending,
		AvailableAt:      now + commissionFreezeSeconds,
	}

	if err := model.DB.Create(record).Error; err != nil {
		// Duplicate trade_no means this payment has already been processed — not an error.
		if isDuplicateKeyError(err) {
			return nil
		}
		return fmt.Errorf("ProcessCommission: insert record failed for trade_no=%s: %w", tradeNo, err)
	}
	return nil
}

// StartCommissionSettleTask starts the periodic task that promotes pending commission
// records to settled and credits the KOL's kol_balance once the 3-day freeze passes.
func StartCommissionSettleTask() {
	commissionSettleOnce.Do(func() {
		if !common.IsMasterNode {
			return
		}
		gopool.Go(func() {
			logger.LogInfo(context.Background(),
				"commission settle task started: tick=10m, mode=recharge-based, freeze=3days",
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
