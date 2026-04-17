package model

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func insertWithdrawalTestUser(t *testing.T, id int, balance float64) {
	t.Helper()
	user := &User{
		Id:         id,
		Username:   fmt.Sprintf("withdraw_test_user_%d", id),
		Password:   "password123",
		Group:      "kol",
		KolBalance: balance,
	}
	require.NoError(t, DB.Create(user).Error)
}

func TestCreateWithdrawalWithDeductionAndRejectRefund(t *testing.T) {
	truncateTables(t)
	insertWithdrawalTestUser(t, 3001, 120)

	withdrawal, err := CreateWithdrawalWithDeduction(3001, 25)
	require.NoError(t, err)
	require.Equal(t, WithdrawalStatusPending, withdrawal.Status)

	var user User
	require.NoError(t, DB.First(&user, 3001).Error)
	assert.InDelta(t, 95.0, user.KolBalance, 0.000001)

	require.NoError(t, RejectWithdrawal(withdrawal.Id, "stripe failed"))

	var reloaded WithdrawalRequest
	require.NoError(t, DB.First(&reloaded, withdrawal.Id).Error)
	assert.Equal(t, WithdrawalStatusRejected, reloaded.Status)
	assert.Equal(t, "stripe failed", reloaded.RejectReason)

	require.NoError(t, DB.First(&user, 3001).Error)
	assert.InDelta(t, 120.0, user.KolBalance, 0.000001)
}

func TestMarkWithdrawalPaidRequiresApprovedStatus(t *testing.T) {
	truncateTables(t)
	insertWithdrawalTestUser(t, 3002, 50)

	withdrawal := &WithdrawalRequest{
		UserId: 3002,
		Amount: 10,
		Status: WithdrawalStatusPending,
	}
	require.NoError(t, DB.Create(withdrawal).Error)

	err := MarkWithdrawalPaid(withdrawal.Id, "tr_pending")
	require.Error(t, err)
	assert.Contains(t, err.Error(), WithdrawalStatusPending)
}

func TestMarkWithdrawalPaidUpdatesApprovedWithdrawal(t *testing.T) {
	truncateTables(t)
	insertWithdrawalTestUser(t, 3003, 50)

	withdrawal := &WithdrawalRequest{
		UserId: 3003,
		Amount: 10,
		Status: WithdrawalStatusApproved,
	}
	require.NoError(t, DB.Create(withdrawal).Error)

	require.NoError(t, MarkWithdrawalPaid(withdrawal.Id, "tr_paid"))

	var reloaded WithdrawalRequest
	require.NoError(t, DB.First(&reloaded, withdrawal.Id).Error)
	assert.Equal(t, WithdrawalStatusPaid, reloaded.Status)
	assert.Equal(t, "tr_paid", reloaded.StripeTransferId)
}
