package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

// SubmitAffiliateApplication handles POST /api/affiliate/apply (public, no auth).
func SubmitAffiliateApplication(c *gin.Context) {
	var app model.AffiliateApplication
	if err := c.ShouldBindJSON(&app); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的请求参数"})
		return
	}
	app.Name = strings.TrimSpace(app.Name)
	app.Email = strings.TrimSpace(app.Email)
	app.Instagram = strings.TrimSpace(app.Instagram)
	app.Tiktok = strings.TrimSpace(app.Tiktok)
	app.Youtube = strings.TrimSpace(app.Youtube)
	app.OtherSocial = strings.TrimSpace(app.OtherSocial)
	if app.Name == "" || app.Email == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "姓名和邮箱不能为空"})
		return
	}
	if err := common.Validate.Var(app.Email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "邮箱格式不正确"})
		return
	}
	if app.Instagram == "" && app.Tiktok == "" && app.Youtube == "" && app.OtherSocial == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请至少填写一个社交账号"})
		return
	}

	if err := model.CreateOrUpdateApplication(&app); err != nil {
		if err == model.ErrAffiliateAlreadyApproved {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "您的申请已通过审核，请查收邮件"})
			return
		}
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "申请已提交"})
}

// GetAffiliateApplicationStatus handles GET /api/affiliate/apply?email=xxx (public, no auth).
func GetAffiliateApplicationStatus(c *gin.Context) {
	email := strings.TrimSpace(c.Query("email"))
	if email == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "缺少邮箱参数"})
		return
	}
	app, err := model.GetApplicationByEmail(email)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if app == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}
	// Return only status — never expose personal data to unauthenticated callers
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status": app.Status,
		},
	})
}

// GetAffiliateApplications handles GET /api/affiliate/applications (reviewer/root only).
func GetAffiliateApplications(c *gin.Context) {
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	apps, total, err := model.GetAllApplications(status, page, pageSize)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    apps,
		"total":   total,
	})
}

// ApproveAffiliateApplication handles POST /api/affiliate/applications/:id/approve (reviewer/root only).
func ApproveAffiliateApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	app, err := model.ApproveApplication(id)
	if err != nil {
		if err == model.ErrAffiliateNotPending {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "该申请不处于待审核状态"})
			return
		}
		common.ApiError(c, err)
		return
	}

	// Send approval email asynchronously
	go sendAffiliateApprovalEmail(app)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已通过申请，邮件已发送"})
}

// RejectAffiliateApplication handles POST /api/affiliate/applications/:id/reject (reviewer/root only).
func RejectAffiliateApplication(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil || id <= 0 {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的 ID"})
		return
	}
	if err := model.RejectApplication(id); err != nil {
		if err == model.ErrAffiliateNotPending {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "该申请不处于待审核状态"})
			return
		}
		common.ApiError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已拒绝申请"})
}

func sendAffiliateApprovalEmail(app *model.AffiliateApplication) {
	if app.KolInviteToken == nil {
		return
	}
	serverAddr := strings.TrimRight(system_setting.ServerAddress, "/")
	link := fmt.Sprintf("%s/register?kol_token=%s", serverAddr, *app.KolInviteToken)
	subject := fmt.Sprintf("【%s】您的达人合作申请已通过", common.SystemName)
	content := fmt.Sprintf(`<p>您好 %s，</p>
<p>恭喜！您申请加入 <strong>%s</strong> 达人计划已通过审核。</p>
<p>请点击以下链接完成注册，注册后您将自动加入达人分组并获得专属佣金功能：</p>
<p><a href="%s">%s</a></p>
<p>如果链接无法点击，请将以下地址复制到浏览器打开：<br>%s</p>
<p><strong>注意：该链接为一次性链接，仅可使用一次。</strong></p>
<p>如有疑问请联系我们。</p>`,
		app.Name, common.SystemName, link, link, link)
	if err := common.SendEmail(subject, app.Email, content); err != nil {
		common.SysError(fmt.Sprintf("failed to send affiliate approval email to %s: %v", app.Email, err))
	}
}
