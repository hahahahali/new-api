package controller

import (
	"fmt"
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/account"
	"github.com/stripe/stripe-go/v81/accountlink"
)

// GetKolDashboard returns the KOL's overview data
func GetKolDashboard(c *gin.Context) {
	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	totalCommission, totalRecharge, commissionCount, err := model.GetCommissionSummaryByInviterId(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	pendingCommission, err := model.GetPendingCommissionAmountByInviterId(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var inviteeCount int64
	model.DB.Model(&model.User{}).Where("inviter_id = ?", userId).Count(&inviteeCount)

	common.ApiSuccess(c, gin.H{
		"kol_balance":              user.KolBalance,
		"kol_pending_balance":      pendingCommission,
		"kol_history_balance":      user.KolHistoryBalance,
		"commission_rate":          setting.KolCommissionRate,
		"invitee_count":            inviteeCount,
		"total_commission":         totalCommission,
		"total_recharge":           totalRecharge,
		"commission_count":         commissionCount,
		"aff_code":                 user.AffCode,
		"stripe_connect_onboarded": user.StripeConnectOnboarded,
		"min_withdrawal_amount":    setting.MinWithdrawalAmount,
	})
}

// GetKolInvitees returns the list of users invited by the KOL
func GetKolInvitees(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)

	invitees, total, err := model.GetKolInvitees(userId, pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": invitees,
			"total": total,
		},
	})
}

// GetKolCommissions returns the commission records for the KOL
func GetKolCommissions(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)

	records, total, err := model.GetCommissionRecordsByInviterId(userId, pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": records,
			"total": total,
		},
	})
}

// GetKolWithdrawals returns the withdrawal requests for the KOL
func GetKolWithdrawals(c *gin.Context) {
	userId := c.GetInt("id")
	pageInfo := common.GetPageQuery(c)

	requests, total, err := model.GetWithdrawalRequestsByUserId(userId, pageInfo.Page, pageInfo.PageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"items": requests,
			"total": total,
		},
	})
}

type KolWithdrawRequest struct {
	Amount float64 `json:"amount" binding:"required"`
}

// RequestKolWithdraw creates a withdrawal request and automatically transfers
// via Stripe Connect. No admin approval is needed.
func RequestKolWithdraw(c *gin.Context) {
	userId := c.GetInt("id")

	var req KolWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if req.Amount < setting.MinWithdrawalAmount {
		common.ApiErrorMsg(c, "提现金额不能低于最低限额")
		return
	}

	if !setting.StripeConnectEnabled || setting.StripeApiSecret == "" {
		common.ApiErrorMsg(c, "Stripe Connect 未启用，无法提现")
		return
	}

	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if !user.StripeConnectOnboarded {
		// DB flag may be stale — do a real-time Stripe check when account exists.
		if user.StripeConnectAccountId != "" {
			stripe.Key = setting.StripeApiSecret
			acct, err := account.GetByID(user.StripeConnectAccountId, nil)
			if err == nil && acct.ChargesEnabled && acct.PayoutsEnabled {
				model.DB.Model(&model.User{}).Where("id = ?", userId).
					Update("stripe_connect_onboarded", true)
			} else {
				common.ApiErrorMsg(c, "请先完成 Stripe Connect 账户认证")
				return
			}
		} else {
			common.ApiErrorMsg(c, "请先完成 Stripe Connect 账户认证")
			return
		}
	}

	// 1. Create withdrawal (pending) — balance deducted immediately
	withdrawal, err := model.CreateWithdrawalWithDeduction(userId, req.Amount)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// 2. Auto-approve (pending → approved)
	changed, err := model.ApproveWithdrawal(withdrawal.Id)
	if err != nil || !changed {
		// Shouldn't happen since we just created it, but handle gracefully
		_ = model.RejectWithdrawal(withdrawal.Id, "系统错误：自动审核失败")
		common.ApiErrorMsg(c, "提现失败，余额已退还")
		return
	}
	withdrawal.Status = model.WithdrawalStatusApproved

	// 3. Stripe Connect transfer
	transfer, err := service.CreateWithdrawalTransfer(withdrawal, user)
	if err != nil {
		// A definitive Stripe API failure is safe to roll back and refund.
		if !service.IsStripeTransferDefinitiveFailure(err) {
			common.SysError(fmt.Sprintf(
				"withdrawal %d: Stripe transfer result is uncertain (%v). Left in approved for reconciliation.",
				withdrawal.Id, err,
			))
			c.JSON(http.StatusOK, gin.H{
				"success": true,
				"message": "提现已提交，Stripe 打款状态确认中，请稍后查看提现记录",
				"data": gin.H{
					"withdrawal":                    withdrawal,
					"transfer_pending_confirmation": true,
				},
			})
			return
		}

		// Stripe transfer failed definitively — revert to pending, then reject to refund balance
		if revertErr := model.RevertWithdrawalToPending(withdrawal.Id); revertErr != nil {
			common.SysError(fmt.Sprintf(
				"withdrawal %d: Stripe transfer failed (%v) AND revert failed (%v). Manual fix required.",
				withdrawal.Id, err, revertErr,
			))
			common.ApiErrorMsg(c, "Stripe 打款失败且状态回退失败，请联系管理员处理")
			return
		}
		_ = model.RejectWithdrawal(withdrawal.Id, "Stripe 打款失败: "+err.Error())
		common.ApiErrorMsg(c, "Stripe 打款失败，余额已退还: "+err.Error())
		return
	}

	// 4. Mark as paid
	if dbErr := model.MarkWithdrawalPaid(withdrawal.Id, transfer.ID); dbErr != nil {
		common.SysError(fmt.Sprintf(
			"CRITICAL: Stripe transfer %s succeeded for withdrawal %d but DB update failed: %v",
			transfer.ID, withdrawal.Id, dbErr,
		))
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Stripe 已打款成功，但状态同步仍在处理中，请稍后查看提现记录",
			"data": gin.H{
				"withdrawal":        withdrawal,
				"transfer_id":       transfer.ID,
				"db_sync_pending":   true,
				"withdrawal_status": model.WithdrawalStatusApproved,
			},
		})
		return
	}
	withdrawal.Status = model.WithdrawalStatusPaid
	withdrawal.StripeTransferId = transfer.ID

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "提现成功",
		"data": gin.H{
			"withdrawal":  withdrawal,
			"transfer_id": transfer.ID,
		},
	})
}

// KolStripeConnectOnboard generates a Stripe Connect onboarding link
func KolStripeConnectOnboard(c *gin.Context) {
	if !setting.StripeConnectEnabled {
		common.ApiErrorMsg(c, "Stripe Connect 未启用")
		return
	}

	if setting.StripeApiSecret == "" {
		common.ApiErrorMsg(c, "Stripe 未配置")
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	stripe.Key = setting.StripeApiSecret

	var accountId string
	if user.StripeConnectAccountId != "" {
		accountId = user.StripeConnectAccountId
	} else {
		// Create a new Express account
		params := &stripe.AccountParams{
			Type: stripe.String(string(stripe.AccountTypeExpress)),
		}
		if user.Email != "" {
			params.Email = stripe.String(user.Email)
		}
		acct, err := account.New(params)
		if err != nil {
			common.ApiErrorMsg(c, "创建 Stripe Connect 账户失败: "+err.Error())
			return
		}
		accountId = acct.ID

		// Save account ID to user
		model.DB.Model(&model.User{}).Where("id = ?", userId).
			Update("stripe_connect_account_id", accountId)
	}

	// Generate onboarding link
	returnURL := system_setting.ServerAddress + "/console/kol"
	refreshURL := system_setting.ServerAddress + "/console/kol"

	linkParams := &stripe.AccountLinkParams{
		Account:    stripe.String(accountId),
		RefreshURL: stripe.String(refreshURL),
		ReturnURL:  stripe.String(returnURL),
		Type:       stripe.String("account_onboarding"),
	}
	link, err := accountlink.New(linkParams)
	if err != nil {
		common.ApiErrorMsg(c, "生成 Onboarding 链接失败: "+err.Error())
		return
	}

	common.ApiSuccess(c, gin.H{
		"onboarding_url": link.URL,
	})
}

// KolStripeConnectStatus checks if the Stripe Connect account onboarding is complete
// and updates the user's stripe_connect_onboarded flag accordingly.
func KolStripeConnectStatus(c *gin.Context) {
	if !setting.StripeConnectEnabled {
		common.ApiErrorMsg(c, "Stripe Connect 未启用")
		return
	}

	userId := c.GetInt("id")
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if user.StripeConnectAccountId == "" {
		common.ApiSuccess(c, gin.H{
			"onboarded":  false,
			"account_id": "",
		})
		return
	}

	stripe.Key = setting.StripeApiSecret

	acct, err := account.GetByID(user.StripeConnectAccountId, nil)
	if err != nil {
		common.ApiErrorMsg(c, "查询 Stripe Connect 账户失败: "+err.Error())
		return
	}

	onboarded := acct.ChargesEnabled && acct.PayoutsEnabled
	if onboarded != user.StripeConnectOnboarded {
		model.DB.Model(&model.User{}).Where("id = ?", userId).
			Update("stripe_connect_onboarded", onboarded)
	}

	common.ApiSuccess(c, gin.H{
		"onboarded":       onboarded,
		"account_id":      user.StripeConnectAccountId,
		"charges_enabled": acct.ChargesEnabled,
		"payouts_enabled": acct.PayoutsEnabled,
	})
}
