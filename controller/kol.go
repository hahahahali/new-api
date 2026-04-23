package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/gin-gonic/gin"

	// [Stripe Connect - disabled]
	// "github.com/QuantumNous/new-api/service"
	// "github.com/QuantumNous/new-api/setting/system_setting"
	// "github.com/stripe/stripe-go/v81"
	// "github.com/stripe/stripe-go/v81/account"
	// "github.com/stripe/stripe-go/v81/accountlink"
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
		"kol_balance":           user.KolBalance,
		"kol_pending_balance":   pendingCommission,
		"kol_history_balance":   user.KolHistoryBalance,
		"commission_rate":       setting.KolCommissionRate,
		"kol_rebate_rate":       user.KolRebateRate,
		"invitee_count":         inviteeCount,
		"total_commission":      totalCommission,
		"total_recharge":        totalRecharge,
		"commission_count":      commissionCount,
		"aff_code":              user.AffCode,
		"min_withdrawal_amount": setting.MinWithdrawalAmount,
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
	Amount      float64 `json:"amount" binding:"required"`
	PaypalEmail string  `json:"paypal_email" binding:"required"`
	PaypalName  string  `json:"paypal_name" binding:"required"`
}

// RequestKolWithdraw creates a pending withdrawal with PayPal payout info.
// Payouts are processed manually by admin on the 1st and 15th of each month.
func RequestKolWithdraw(c *gin.Context) {
	userId := c.GetInt("id")

	var req KolWithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误：请填写提现金额、PayPal 邮箱和收款人姓名")
		return
	}

	if req.Amount < setting.MinWithdrawalAmount {
		common.ApiErrorMsg(c, fmt.Sprintf("提现金额不能低于最低限额 $%.2f", setting.MinWithdrawalAmount))
		return
	}

	withdrawal, err := model.CreateWithdrawalWithDeduction(userId, req.Amount, req.PaypalEmail, req.PaypalName)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "提现申请已提交",
		"data":    withdrawal,
	})
}

type KolSetRebateRequest struct {
	RebateRate float64 `json:"rebate_rate"`
}

// SetKolRebateRate allows a KOL to set how much of their commission they share back
// with invited users as a purchase discount (0 to KolCommissionRate).
// KolUpdateAffCode lets a KOL user set their own custom invite code (no admin approval needed).
func KolUpdateAffCode(c *gin.Context) {
	userId := c.GetInt("id")
	var req struct {
		AffCode string `json:"aff_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	code := strings.TrimSpace(req.AffCode)
	if len(code) < 4 || len(code) > 20 {
		common.ApiErrorMsg(c, "邀请码长度必须在 4～20 个字符之间")
		return
	}
	for _, ch := range code {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '_') {
			common.ApiErrorMsg(c, "邀请码只能包含字母、数字和下划线")
			return
		}
	}
	existingId, _ := model.GetUserIdByAffCode(code)
	if existingId > 0 && existingId != userId {
		common.ApiErrorMsg(c, "该邀请码已被其他用户使用，请换一个")
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).
		Update("aff_code", code).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{"aff_code": code})
}

func SetKolRebateRate(c *gin.Context) {
	userId := c.GetInt("id")
	var req KolSetRebateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}
	maxRate := setting.KolCommissionRate
	if req.RebateRate < 0 || req.RebateRate > maxRate {
		common.ApiErrorMsg(c, fmt.Sprintf("让利比例必须在 0~%.0f%% 之间", maxRate*100))
		return
	}
	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).
		Update("kol_rebate_rate", req.RebateRate).Error; err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// [Stripe Connect - disabled]
// KolStripeConnectOnboard and KolStripeConnectStatus have been replaced by PayPal manual payouts.
// Original code preserved below for reference.

/*
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
		model.DB.Model(&model.User{}).Where("id = ?", userId).
			Update("stripe_connect_account_id", accountId)
	}
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
	common.ApiSuccess(c, gin.H{"onboarding_url": link.URL})
}

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
		common.ApiSuccess(c, gin.H{"onboarded": false, "account_id": ""})
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
*/
