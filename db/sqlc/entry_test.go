package db

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateEntry(t *testing.T) {
	account := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    10,
	}

	updatedEntry, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)
	require.NotEmpty(t, updatedEntry)

	entry, err := testQueries.GetEntry(context.Background(), updatedEntry.ID)
	require.NoError(t, err)
	require.NotEmpty(t, entry)
	require.Equal(t, updatedEntry, entry)
}

func TestListEntries(t *testing.T) {
	account := createRandomAccount(t)

	arg := CreateEntryParams{
		AccountID: account.ID,
		Amount:    10,
	}
	_, err := testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)

	_, err = testQueries.CreateEntry(context.Background(), arg)
	require.NoError(t, err)

	entries, err := testQueries.ListEntries(context.Background(), ListEntriesParams{
		AccountID: account.ID,
		Limit:     2,
		Offset:    0,
	})
	require.NoError(t, err)
	require.Len(t, entries, 2)
}
