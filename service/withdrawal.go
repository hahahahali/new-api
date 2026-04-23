package service

// [Stripe Connect - disabled]
// The Stripe Connect automatic transfer logic has been replaced by manual PayPal payouts.
// Admin manually enters a PayPal transaction ID after completing the transfer.
// Original code preserved below for reference.

/*
import (
	"errors"
	"fmt"

	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting"
	"github.com/stripe/stripe-go/v81"
	stripeTransfer "github.com/stripe/stripe-go/v81/transfer"
)

func GetWithdrawalTransferIdempotencyKey(withdrawalID int) string {
	return fmt.Sprintf("kol-withdrawal-%d-transfer", withdrawalID)
}

func IsStripeTransferDefinitiveFailure(err error) bool {
	var stripeErr *stripe.Error
	if !errors.As(err, &stripeErr) {
		return false
	}
	if stripeErr.HTTPStatusCode >= 500 {
		return false
	}
	return stripeErr.Code != stripe.ErrorCodeIdempotencyKeyInUse
}

func CreateWithdrawalTransfer(withdrawal *model.WithdrawalRequest, user *model.User) (*stripe.Transfer, error) {
	stripe.Key = setting.StripeApiSecret

	amountCents := int64(withdrawal.Amount * 100)
	params := &stripe.TransferParams{
		Amount:        stripe.Int64(amountCents),
		Currency:      stripe.String(string(stripe.CurrencyUSD)),
		Destination:   stripe.String(user.StripeConnectAccountId),
		Description:   stripe.String(fmt.Sprintf("KOL withdrawal #%d", withdrawal.Id)),
		TransferGroup: stripe.String(GetWithdrawalTransferIdempotencyKey(withdrawal.Id)),
	}
	params.SetIdempotencyKey(GetWithdrawalTransferIdempotencyKey(withdrawal.Id))
	params.AddMetadata("withdrawal_id", fmt.Sprintf("%d", withdrawal.Id))
	params.AddMetadata("user_id", fmt.Sprintf("%d", withdrawal.UserId))

	return stripeTransfer.New(params)
}
*/
