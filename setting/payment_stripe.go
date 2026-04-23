package setting

var StripeApiSecret = ""
var StripeWebhookSecret = ""
var StripeUnitPrice = 8.0
var StripeMinTopUp = 1
var StripePromotionCodesEnabled = false

// StripePriceId is preserved from upstream for compile compatibility only.
// Our code path uses dynamic price_data via genStripeLinkPriceData and never
// reads this value. Kept so upstream genStripeLink remains buildable and
// future merges with QuantumNous/new-api stay conflict-free.
var StripePriceId = ""
