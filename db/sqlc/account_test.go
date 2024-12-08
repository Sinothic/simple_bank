package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func createRandomAccount(t *testing.T) Account {
	balance, err := faker.RandomInt(1, 1000, 1)
	require.NoError(t, err)

	user, err := testQueries.CreateUser(context.Background(), CreateUserParams{
		HashedPassword: "xxxxxxxx",
		Username:       faker.Username(),
		FullName:       faker.Name(),
		Email:          faker.Email(),
	})
	require.NoError(t, err)

	args := CreateAccountParams{
		UserID:   user.ID,
		Balance:  int64(balance[0]),
		Currency: faker.Currency(),
	}

	account, err := testQueries.CreateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, args.UserID, account.UserID)
	require.Equal(t, args.Balance, account.Balance)
	require.Equal(t, args.Currency, account.Currency)

	require.NotZero(t, account.ID)
	require.NotZero(t, account.CreatedAt)

	return account
}

func TestCreateAccount(t *testing.T) {
	createRandomAccount(t)
}

func TestGetAccount(t *testing.T) {
	expectedAccount := createRandomAccount(t)
	dbAccount, err := testQueries.GetAccount(context.Background(), expectedAccount.ID)
	require.NoError(t, err)

	require.NotEmpty(t, dbAccount)
	require.Equal(t, expectedAccount.ID, dbAccount.ID)
	require.Equal(t, expectedAccount.UserID, dbAccount.UserID)
	require.Equal(t, expectedAccount.Balance, dbAccount.Balance)
	require.Equal(t, expectedAccount.Currency, dbAccount.Currency)
	require.WithinDuration(t, expectedAccount.CreatedAt, dbAccount.CreatedAt, time.Second)
}

func TestUpdateAccount(t *testing.T) {
	account := createRandomAccount(t)

	args := UpdateAccountParams{
		ID:      account.ID,
		Balance: account.Balance + 100,
	}

	updatedAccount, err := testQueries.UpdateAccount(context.Background(), args)
	require.NoError(t, err)
	require.NotEmpty(t, account)

	require.Equal(t, account.ID, updatedAccount.ID)
	require.Equal(t, args.Balance, updatedAccount.Balance)
}

func TestDeleteAccount(t *testing.T) {
	account := createRandomAccount(t)

	err := testQueries.DeleteAccount(context.Background(), account.ID)
	require.NoError(t, err)

	deletedAccount, err := testQueries.GetAccount(context.Background(), account.ID)
	require.Error(t, err)
	require.Empty(t, deletedAccount)
	require.EqualError(t, err, sql.ErrNoRows.Error())
}

func TestListAccounts(t *testing.T) {
	for i := 0; i < 10; i++ {
		createRandomAccount(t)
	}

	args := ListAccountsParams{
		Limit:  5,
		Offset: 5,
	}

	accounts, err := testQueries.ListAccounts(context.Background(), args)
	require.NoError(t, err)
	require.Len(t, accounts, 5)

	for _, account := range accounts {
		require.NotEmpty(t, account)
		require.NotZero(t, account.ID)
		require.NotZero(t, account.CreatedAt)
	}
}
