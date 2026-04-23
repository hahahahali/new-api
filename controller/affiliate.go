package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"
	"github.com/gin-gonic/gin"
)

const affiliateEmailVerifyPurpose = "affiliate_ev"

func affiliateVerifyRedisKey(email string) string {
	return "affiliate_email_verify:" + email
}

// checkAffiliateEmailCode checks the stored code without consuming it.
func checkAffiliateEmailCode(email, code string) bool {
	if common.RedisEnabled {
		stored, err := common.RedisGet(affiliateVerifyRedisKey(email))
		return err == nil && stored == code
	}
	return common.VerifyCodeWithKey(email, code, affiliateEmailVerifyPurpose)
}

// consumeAffiliateEmailCode deletes the code so it cannot be reused.
func consumeAffiliateEmailCode(email string) {
	if common.RedisEnabled {
		_ = common.RedisDel(affiliateVerifyRedisKey(email))
	} else {
		common.DeleteKey(email, affiliateEmailVerifyPurpose)
	}
}

// SubmitAffiliateApplication handles POST /api/affiliate/apply (public, no auth).
func SubmitAffiliateApplication(c *gin.Context) {
	var req struct {
		Name        string `json:"name"`
		Email       string `json:"email"`
		Phone       string `json:"phone"`
		Country     string `json:"country"`
		Instagram   string `json:"instagram"`
		Tiktok      string `json:"tiktok"`
		Youtube     string `json:"youtube"`
		OtherSocial string `json:"other_social"`
		VerifyCode  string `json:"verify_code"`
		Lang        string `json:"lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的请求参数"})
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)
	req.Instagram = strings.TrimSpace(req.Instagram)
	req.Tiktok = strings.TrimSpace(req.Tiktok)
	req.Youtube = strings.TrimSpace(req.Youtube)
	req.OtherSocial = strings.TrimSpace(req.OtherSocial)
	req.VerifyCode = strings.TrimSpace(req.VerifyCode)
	if req.Name == "" || req.Email == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "姓名和邮箱不能为空"})
		return
	}
	if err := common.Validate.Var(req.Email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "邮箱格式不正确"})
		return
	}
	if req.Instagram == "" && req.Tiktok == "" && req.Youtube == "" && req.OtherSocial == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请至少填写一个社交账号"})
		return
	}
	if req.VerifyCode == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请先获取并填写邮箱验证码"})
		return
	}
	if !checkAffiliateEmailCode(req.Email, req.VerifyCode) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "验证码错误或已过期，请重新获取"})
		return
	}
	consumeAffiliateEmailCode(req.Email)

	app := model.AffiliateApplication{
		Name:        req.Name,
		Email:       req.Email,
		Phone:       req.Phone,
		Country:     req.Country,
		Instagram:   req.Instagram,
		Tiktok:      req.Tiktok,
		Youtube:     req.Youtube,
		OtherSocial: req.OtherSocial,
		Lang:        req.Lang,
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

// GetAffiliateApplicationStatus handles POST /api/affiliate/status (public, email verification required).
func GetAffiliateApplicationStatus(c *gin.Context) {
	var req struct {
		Email      string `json:"email"`
		VerifyCode string `json:"verify_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的请求参数"})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.VerifyCode = strings.TrimSpace(req.VerifyCode)
	if req.Email == "" || req.VerifyCode == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "请先完成邮箱验证"})
		return
	}
	if err := common.Validate.Var(req.Email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "邮箱格式不正确"})
		return
	}
	if !checkAffiliateEmailCode(req.Email, req.VerifyCode) {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "验证码错误或已过期，请重新获取"})
		return
	}

	app, err := model.GetApplicationByEmail(req.Email)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if app == nil {
		c.JSON(http.StatusOK, gin.H{"success": true, "data": nil})
		return
	}
	// Email ownership verified — safe to return the applicant's own data
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"status":       app.Status,
			"name":         app.Name,
			"phone":        app.Phone,
			"country":      app.Country,
			"instagram":    app.Instagram,
			"tiktok":       app.Tiktok,
			"youtube":      app.Youtube,
			"other_social": app.OtherSocial,
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
	var body struct {
		Reason string `json:"reason"`
	}
	// Ignore parse error — reason is optional
	_ = c.ShouldBindJSON(&body)

	app, err := model.RejectApplication(id, body.Reason)
	if err != nil {
		if err == model.ErrAffiliateNotPending {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "该申请不处于待审核状态"})
			return
		}
		common.ApiError(c, err)
		return
	}

	go sendAffiliateRejectionEmail(app)

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "已拒绝申请，通知邮件已发送"})
}

func sendAffiliateApprovalEmail(app *model.AffiliateApplication) {
	if app.KolInviteToken == nil {
		return
	}
	serverAddr := strings.TrimRight(system_setting.ServerAddress, "/")
	link := fmt.Sprintf("%s/register?kol_token=%s", serverAddr, *app.KolInviteToken)
	subject, content := common.BuildAffiliateApprovalEmail(app.Lang, app.Name, common.SystemName, link)
	if err := common.SendEmail(subject, app.Email, content); err != nil {
		common.SysError(fmt.Sprintf("failed to send affiliate approval email to %s: %v", app.Email, err))
	}
}

func sendAffiliateRejectionEmail(app *model.AffiliateApplication) {
	subject, content := common.BuildAffiliateRejectionEmail(app.Lang, app.Name, common.SystemName, app.RejectReason)
	if err := common.SendEmail(subject, app.Email, content); err != nil {
		common.SysError(fmt.Sprintf("failed to send affiliate rejection email to %s: %v", app.Email, err))
	}
}

// SendAffiliateEmailCode handles POST /api/affiliate/send-email-code (public, rate-limited).
func SendAffiliateEmailCode(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Lang  string `json:"lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "无效的请求参数"})
		return
	}
	email := strings.TrimSpace(req.Email)
	if email == "" {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "邮箱不能为空"})
		return
	}
	if err := common.Validate.Var(email, "required,email"); err != nil {
		c.JSON(http.StatusOK, gin.H{"success": false, "message": "邮箱格式不正确"})
		return
	}

	code := fmt.Sprintf("%06d", common.GetRandomInt(1000000))
	if common.RedisEnabled {
		if err := common.RedisSet(affiliateVerifyRedisKey(email), code, 10*time.Minute); err != nil {
			c.JSON(http.StatusOK, gin.H{"success": false, "message": "服务暂时不可用，请稍后重试"})
			return
		}
	} else {
		common.RegisterVerificationCodeWithKey(email, code, affiliateEmailVerifyPurpose)
	}

	go sendAffiliateVerificationCodeEmail(email, code, req.Lang)
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "验证码已发送，有效期 10 分钟"})
}

func sendAffiliateVerificationCodeEmail(email, code, lang string) {
	subject, content := common.BuildAffiliateVerificationEmail(lang, common.SystemName, code)
	if err := common.SendEmail(subject, email, content); err != nil {
		common.SysError(fmt.Sprintf("failed to send affiliate verification code to %s: %v", email, err))
	}
}
