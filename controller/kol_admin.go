package controller

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/gin-gonic/gin"
)

// AdminUpdateKolAffCode allows super admin to set a custom aff_code for a KOL user
func AdminUpdateKolAffCode(c *gin.Context) {
	idStr := c.Param("id")
	userId, err := strconv.Atoi(idStr)
	if err != nil || userId <= 0 {
		common.ApiErrorMsg(c, "无效的用户 ID")
		return
	}

	var req struct {
		AffCode string `json:"aff_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorMsg(c, "参数错误")
		return
	}

	if len(req.AffCode) > 32 {
		common.ApiErrorMsg(c, "邀请码长度不能超过 32 位")
		return
	}

	user, err := model.GetUserById(userId, false)
	if err != nil || user == nil {
		common.ApiErrorMsg(c, "用户不存在")
		return
	}

	if user.Group != "kol" {
		common.ApiErrorMsg(c, "该用户不是达人账号")
		return
	}

	// Check uniqueness
	existingId, _ := model.GetUserIdByAffCode(req.AffCode)
	if existingId > 0 && existingId != userId {
		common.ApiErrorMsg(c, "该邀请码已被其他用户使用")
		return
	}

	if err := model.DB.Model(&model.User{}).Where("id = ?", userId).
		Update("aff_code", req.AffCode).Error; err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{"aff_code": req.AffCode})
}

// AdminGetAllWithdrawals returns all withdrawal requests for admin review
func AdminGetAllWithdrawals(c *gin.Context) {
	pageInfo := common.GetPageQuery(c)
	status := c.Query("status")

	requests, total, err := model.GetAllWithdrawalRequests(status, pageInfo.Page, pageInfo.PageSize)
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

// AdminReconcileWithdrawalTransfer retries Stripe transfer synchronization for an approved withdrawal.
func AdminReconcileWithdrawalTransfer(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		common.ApiErrorMsg(c, "无效的提现申请 ID")
		return
	}

	withdrawal, err := model.GetWithdrawalRequestById(id)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if withdrawal.Status == model.WithdrawalStatusPaid {
		common.ApiSuccess(c, gin.H{
			"withdrawal":  withdrawal,
			"transfer_id": withdrawal.StripeTransferId,
		})
		return
	}
	if withdrawal.Status != model.WithdrawalStatusApproved {
		common.ApiErrorMsg(c, "只有已通过但未完成打款同步的提现单才能重试")
		return
	}

	user, err := model.GetUserById(withdrawal.UserId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if user.StripeConnectAccountId == "" {
		common.ApiErrorMsg(c, "该用户未绑定 Stripe Connect 账户")
		return
	}

	transfer, err := service.CreateWithdrawalTransfer(withdrawal, user)
	if err != nil {
		if service.IsStripeTransferDefinitiveFailure(err) {
			common.ApiErrorMsg(c, "Stripe 打款失败: "+err.Error())
			return
		}
		common.SysError(fmt.Sprintf(
			"withdrawal %d: admin reconciliation could not confirm Stripe transfer result: %v",
			withdrawal.Id, err,
		))
		common.ApiErrorMsg(c, "Stripe 打款状态仍未确认，请稍后重试同步")
		return
	}

	if dbErr := model.MarkWithdrawalPaid(withdrawal.Id, transfer.ID); dbErr != nil {
		common.SysError(fmt.Sprintf(
			"withdrawal %d: admin reconciliation got Stripe transfer %s but DB sync failed: %v",
			withdrawal.Id, transfer.ID, dbErr,
		))
		common.ApiErrorMsg(c, "Stripe 已返回打款成功，但本地状态同步失败，请稍后重试同步")
		return
	}

	withdrawal.Status = model.WithdrawalStatusPaid
	withdrawal.StripeTransferId = transfer.ID
	common.ApiSuccess(c, gin.H{
		"withdrawal":  withdrawal,
		"transfer_id": transfer.ID,
	})
}
