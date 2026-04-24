package operation_setting

import "github.com/QuantumNous/new-api/setting/config"

type PaymentSetting struct {
	AmountOptions     []int           `json:"amount_options"`
	AmountDiscount    map[int]float64 `json:"amount_discount"`     // 充值金额对应的折扣，例如 100 元 0.9 表示 100 元充值享受 9 折优惠
	AmountOptionNames []string        `json:"amount_option_names"` // 每个充值档位的展示名称，与 AmountOptions 一一对应
	AmountOptionDescs []string        `json:"amount_option_descs"` // 每个充值档位的副标题描述，与 AmountOptions 一一对应
}

// 默认配置
var paymentSetting = PaymentSetting{
	AmountOptions:     []int{10, 20, 50, 100, 200, 500},
	AmountDiscount:    map[int]float64{},
	AmountOptionNames: []string{},
	AmountOptionDescs: []string{},
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("payment_setting", &paymentSetting)
}

func GetPaymentSetting() *PaymentSetting {
	return &paymentSetting
}
