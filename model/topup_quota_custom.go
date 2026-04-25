package model

// creditQuotaRatio 是 1 积分对应的 quota 数量（冻结值）。
// 来源：StripeUnitPrice(0.0008167) × QuotaPerUnit(500000) = 408.35
// 冻结后 StripeUnitPrice 调价只影响新套餐价格，不影响已有余额的积分显示。
const creditQuotaRatio = 408.35

func CalcQuotaByAmount(topUp *TopUp) float64 {
	return float64(topUp.Amount) * creditQuotaRatio
}

func CreditDivisor() float64 {
	return creditQuotaRatio
}
