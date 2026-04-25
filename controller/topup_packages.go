package controller

import (
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/QuantumNous/new-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
)

// resolveI18nNames parses a raw JSON string that is either a plain []string
// or a multilingual map {"zh":[...],"en":[...]} and returns the slice for the
// requested lang, falling back to "zh" and then the first available language.
func resolveI18nNames(raw, lang string) []string {
	if raw == "" {
		return nil
	}
	var i18nMap map[string][]string
	if err := common.UnmarshalJsonStr(raw, &i18nMap); err == nil && len(i18nMap) > 0 {
		if v, ok := i18nMap[lang]; ok {
			return v
		}
		if v, ok := i18nMap["zh"]; ok {
			return v
		}
		for _, v := range i18nMap {
			return v
		}
	}
	var simple []string
	if err := common.UnmarshalJsonStr(raw, &simple); err == nil {
		return simple
	}
	return nil
}

// localizedAmountOptions reads amount_option_names and amount_option_descs from
// the live OptionMap and resolves them for the requested language.
// Also called by GetTopUpInfo in topup.go — keep the touch surface on that
// upstream function minimal (one call site).
func localizedAmountOptions(c *gin.Context) (names, descs []string) {
	lang := c.DefaultQuery("lang", "zh")
	common.OptionMapRWMutex.RLock()
	namesRaw := common.OptionMap["payment_setting.amount_option_names"]
	descsRaw := common.OptionMap["payment_setting.amount_option_descs"]
	common.OptionMapRWMutex.RUnlock()
	return resolveI18nNames(namesRaw, lang), resolveI18nNames(descsRaw, lang)
}

// GetPublicTopupPackages returns Stripe topup packages with default-group pricing.
// No authentication required — only final computed prices are exposed, no internal config.
func GetPublicTopupPackages(c *gin.Context) {
	if setting.StripeUnitPrice <= 0 {
		common.ApiErrorMsg(c, "Stripe not configured")
		return
	}

	amountOptions := operation_setting.GetPaymentSetting().AmountOptions

	type PackageInfo struct {
		Amount      int    `json:"amount"`
		Credits     int    `json:"credits"`
		Price       string `json:"price"`
		Tag         string `json:"tag,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	}

	amountOptionNames, amountOptionDescs := localizedAmountOptions(c)
	packages := make([]PackageInfo, 0, len(amountOptions))
	for i, amount := range amountOptions {
		payMoney := getStripePayMoney(float64(amount), "default")
		if payMoney <= 0 {
			continue
		}
		name := ""
		if i < len(amountOptionNames) {
			name = amountOptionNames[i]
		}
		desc := ""
		if i < len(amountOptionDescs) {
			desc = amountOptionDescs[i]
		}
		packages = append(packages, PackageInfo{
			Amount:      amount,
			Credits:     amount,
			Price:       strconv.FormatFloat(payMoney, 'f', 2, 64),
			Name:        name,
			Description: desc,
		})
	}

	n := len(packages)
	if n >= 2 {
		packages[n-2].Tag = "最受欢迎"
		packages[n-1].Tag = "最划算"
	}

	common.ApiSuccess(c, gin.H{"packages": packages, "credit_divisor": model.CreditDivisor()})
}
