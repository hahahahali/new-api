package model

// creditQuotaRatio 是 1 积分对应的 quota 数量（冻结值）。
// 由 QuotaPerUnit(500,000) / 100 推导，确保前端 model_price×100 的积分显示
// 与 quota/creditQuotaRatio 的余额显示完全一致。
const creditQuotaRatio = 5000

func CalcQuotaByAmount(topUp *TopUp) int {
	return int(topUp.Amount) * creditQuotaRatio
}

func CreditDivisor() int {
	return creditQuotaRatio
}
