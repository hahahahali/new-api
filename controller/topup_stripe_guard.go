package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/gin-gonic/gin"
)

// GuardedRequestStripePay 在调用上游 RequestStripePay 之前，
// 校验 amount 是否在 AmountOptions 白名单中，替代上游硬编码的 amount > 10000 上限。
func GuardedRequestStripePay(c *gin.Context) {
	var req StripePayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{"message": "error", "data": "参数错误"})
		return
	}

	if !isAmountInOptions(req.Amount) {
		c.JSON(http.StatusOK, gin.H{"message": "无效的充值数量", "data": 10})
		return
	}

	stripeAdaptor.RequestPay(c, &req)
}

func isAmountInOptions(amount int64) bool {
	for _, opt := range operation_setting.GetPaymentSetting().AmountOptions {
		if int64(opt) == amount {
			return true
		}
	}
	return false
}
