package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

type transactionResult struct {
	Result TransferTxResult
	Error  error
}

func TestTransferTx(t *testing.T) {
	store := NewStore(testDBConnection)

	// Create two accounts
	account1 := createRandomAccount(t)
	account2 := createRandomAccount(t)

	ttTransactions := 5
	amount := int64(10)

	// Run concurrent transactions
	results := make(chan transactionResult)

	for i := 0; i < ttTransactions; i++ {
		go func() {
			ctx := context.Background()
			params := TransferTxParams{
				FromAccountID: account1.ID,
				ToAccountID:   account2.ID,
				Amount:        amount,
			}
			result, err := store.TransferTx(ctx, params)
			results <- transactionResult{Result: result, Error: err}
		}()
	}

	// Check the results
	existed := make(map[int]bool)
	for i := 0; i < ttTransactions; i++ {
		result := <-results
		require.NoError(t, result.Error)
		require.NotEmpty(t, result.Result)

		// Check the transfer details
		require.Equal(t, account1.ID, result.Result.Transfer.FromAccountID)
		require.Equal(t, account2.ID, result.Result.Transfer.ToAccountID)
		require.Equal(t, amount, result.Result.Transfer.Amount)

		// check if transfer was created
		_, err := store.GetTransfer(context.Background(), result.Result.Transfer.ID)
		require.NoError(t, err)

		// check transfer entries
		fromEntry := result.Result.FromEntry
		toEntry := result.Result.ToEntry

		require.NotEmpty(t, fromEntry.ID)
		require.NotEmpty(t, toEntry.ID)
		require.Equal(t, account1.ID, fromEntry.AccountID)
		require.Equal(t, -amount, fromEntry.Amount)
		require.Equal(t, account2.ID, toEntry.AccountID)
		require.Equal(t, amount, toEntry.Amount)
		require.NotZero(t, fromEntry.CreatedAt)
		require.NotZero(t, toEntry.CreatedAt)

		// check if entries were created
		_, err = store.GetEntry(context.Background(), fromEntry.ID)
		require.NoError(t, err)
		_, err = store.GetEntry(context.Background(), toEntry.ID)
		require.NoError(t, err)

		fromAccount := result.Result.FromAccount
		require.NotEmpty(t, fromAccount)
		require.Equal(t, account1.ID, fromAccount.ID)

		toAccount := result.Result.ToAccount
		require.NotEmpty(t, toAccount)
		require.Equal(t, account2.ID, toAccount.ID)

		//check account balances
		diff1 := account1.Balance - fromAccount.Balance
		diff2 := toAccount.Balance - account2.Balance
		require.Equal(t, diff1, diff2)
		require.True(t, diff1 > 0)
		require.True(t, diff1%amount == 0)

		k := int(diff1 / amount)
		require.True(t, k >= 1 && k <= ttTransactions)
		require.NotContains(t, existed, k)
		existed[k] = true
	}

	// check the final updated balances
	updatedAccount1, err := store.GetAccount(context.Background(), account1.ID)
	require.NoError(t, err)

	updatedAccount2, err := store.GetAccount(context.Background(), account2.ID)
	require.NoError(t, err)

	require.Equal(t, account1.Balance-int64(ttTransactions)*amount, updatedAccount1.Balance)
	require.Equal(t, account2.Balance+int64(ttTransactions)*amount, updatedAccount2.Balance)

}
