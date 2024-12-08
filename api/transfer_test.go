package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/Sinothic/simplebank/db/sqlc"
	"github.com/Sinothic/simplebank/db/sqlc/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestCreateTransfer(t *testing.T) {

	type apiTest[A any, R any] struct {
		Argument A
		Response R
		Err      error
		Times    int
	}

	type responseError struct {
		Error string `json:"error"`
	}

	testCases := []struct {
		name               string
		transferRequest    transferRequest
		TransferTx         apiTest[db.TransferTxParams, db.TransferTxResult]
		GetAccountFrom     apiTest[int64, db.Account]
		GetAccountTo       apiTest[int64, db.Account]
		expectedStatusCode int
		expectedError      *responseError
	}{
		{
			name: "bad request error (invalid from account id)",
			transferRequest: transferRequest{
				FromAccountID: 0,
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "invalid account from (random error)",
			transferRequest: transferRequest{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        100,
				Currency:      "USD",
			},
			expectedStatusCode: http.StatusInternalServerError,
			GetAccountFrom: apiTest[int64, db.Account]{
				Argument: 1,
				Err:      errors.New("some error"),
				Times:    1,
			},
		},
		{
			name: "invalid account from (not found)",
			transferRequest: transferRequest{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        100,
				Currency:      "USD",
			},
			expectedStatusCode: http.StatusNotFound,
			GetAccountFrom: apiTest[int64, db.Account]{
				Argument: 1,
				Err:      sql.ErrNoRows,
				Times:    1,
			},
		},
		{
			name: "invalid account from (invalid currency)",
			transferRequest: transferRequest{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        100,
				Currency:      "EUR",
			},
			expectedStatusCode: http.StatusBadRequest,
			GetAccountFrom: apiTest[int64, db.Account]{
				Argument: 1,
				Response: db.Account{Currency: "USD"},
				Times:    1,
			},
		},
		{
			name: "invalid account to (random error)",
			transferRequest: transferRequest{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        100,
				Currency:      "USD",
			},
			expectedStatusCode: http.StatusInternalServerError,
			GetAccountFrom: apiTest[int64, db.Account]{
				Argument: 1,
				Response: db.Account{Currency: "USD"},
				Times:    1,
			},
			GetAccountTo: apiTest[int64, db.Account]{
				Err:   errors.New("some error"),
				Times: 1,
			},
		},
		{
			name: "error while transferring money",
			transferRequest: transferRequest{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        100,
				Currency:      "USD",
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedError: &responseError{
				Error: "some error while transferring the money",
			},
			GetAccountFrom: apiTest[int64, db.Account]{
				Argument: 1,
				Response: db.Account{Currency: "USD"},
				Times:    1,
			},
			GetAccountTo: apiTest[int64, db.Account]{
				Response: db.Account{Currency: "USD"},
				Times:    1,
			},
			TransferTx: apiTest[db.TransferTxParams, db.TransferTxResult]{
				Argument: db.TransferTxParams{
					FromAccountID: 1,
					ToAccountID:   2,
					Amount:        100,
				},
				Err:   errors.New("some error while transferring the money"),
				Times: 1,
			},
		},
		{
			name: "successfully make as transfer",
			transferRequest: transferRequest{
				FromAccountID: 1,
				ToAccountID:   2,
				Amount:        100,
				Currency:      "USD",
			},
			expectedStatusCode: http.StatusOK,
			GetAccountFrom: apiTest[int64, db.Account]{
				Argument: 1,
				Response: db.Account{
					ID:       1,
					UserID:   1,
					Balance:  100,
					Currency: "USD",
				},
				Times: 1,
			},
			GetAccountTo: apiTest[int64, db.Account]{
				Response: db.Account{
					ID:       2,
					UserID:   2,
					Balance:  100,
					Currency: "USD",
				},
				Times: 1,
			},
			TransferTx: apiTest[db.TransferTxParams, db.TransferTxResult]{
				Argument: db.TransferTxParams{
					FromAccountID: 1,
					ToAccountID:   2,
					Amount:        100,
				},
				Response: db.TransferTxResult{
					Transfer: db.Transfer{
						ID:            1,
						FromAccountID: 1,
						ToAccountID:   2,
						Amount:        100,
					},
					FromAccount: db.Account{
						ID:       1,
						UserID:   1,
						Balance:  0,
						Currency: "USD",
					},
					ToAccount: db.Account{
						ID:       2,
						UserID:   2,
						Balance:  200,
						Currency: "USD",
					},
					FromEntry: db.Entry{
						ID:        1,
						AccountID: 1,
						Amount:    -100,
					},
					ToEntry: db.Entry{
						ID:        2,
						AccountID: 2,
						Amount:    100,
					},
				},
				Times: 1,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().
				TransferTx(gomock.Any(), tc.TransferTx.Argument).
				Return(tc.TransferTx.Response, tc.TransferTx.Err).
				Times(tc.TransferTx.Times)

			mockStore.EXPECT().
				GetAccount(gomock.Any(), tc.transferRequest.FromAccountID).
				Return(tc.GetAccountFrom.Response, tc.GetAccountFrom.Err).
				Times(tc.GetAccountFrom.Times)

			mockStore.EXPECT().
				GetAccount(gomock.Any(), tc.transferRequest.ToAccountID).
				Return(tc.GetAccountTo.Response, tc.GetAccountTo.Err).
				Times(tc.GetAccountTo.Times)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/transfers")

			bodyJson, err := json.Marshal(tc.transferRequest)
			if err != nil {
				require.NoError(t, err)
			}

			request, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(bodyJson))
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			responseBody, err := io.ReadAll(recorder.Body)
			require.NoError(t, err)

			var responseStruct db.TransferTxResult
			err = json.Unmarshal(responseBody, &responseStruct)
			require.NoError(t, err)

			require.Equal(t, tc.expectedStatusCode, recorder.Code)
			if tc.expectedStatusCode == http.StatusOK {
				require.Equal(t, tc.TransferTx.Response, responseStruct)
			}

			if tc.expectedError != nil {
				var responseError responseError
				err = json.Unmarshal(responseBody, &responseError)
				require.NoError(t, err)
				require.Equal(t, tc.expectedError, &responseError)
			}

		})
	}

}
