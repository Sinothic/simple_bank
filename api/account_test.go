package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/Sinothic/simplebank/db/sqlc"
	"github.com/Sinothic/simplebank/db/sqlc/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestGetAccountApi(t *testing.T) {

	randomAccount := createRandomAccount()

	type apiTest[A any, R any] struct {
		Argument A
		Response R
		Err      error
		Times    int
	}
	testCases := []struct {
		name               string
		GetAccount         apiTest[int64, db.Account]
		expectedStatusCode int
	}{
		{
			name:               "bad request error",
			GetAccount:         apiTest[int64, db.Account]{},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "no record found error",
			GetAccount: apiTest[int64, db.Account]{
				Argument: 1,
				Err:      sql.ErrNoRows,
				Times:    1,
			},
			expectedStatusCode: http.StatusNotFound,
		},
		{
			name: "db return random error",
			GetAccount: apiTest[int64, db.Account]{
				Argument: 1,
				Err:      sql.ErrConnDone,
				Times:    1,
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
		{
			name: "successful request",
			GetAccount: apiTest[int64, db.Account]{
				Argument: 1,
				Response: randomAccount,
				Times:    1,
			},
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockStore := mocks.NewMockStore(ctrl)
			mockStore.EXPECT().
				GetAccount(gomock.Any(), tc.GetAccount.Argument).
				Return(tc.GetAccount.Response, tc.GetAccount.Err).
				Times(tc.GetAccount.Times)

			server := NewServer(mockStore)
			recorder := httptest.NewRecorder()

			url := fmt.Sprintf("/accounts/%d", tc.GetAccount.Argument)
			request, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			server.router.ServeHTTP(recorder, request)
			require.Equal(t, tc.expectedStatusCode, recorder.Code)
			if tc.GetAccount.Response != (db.Account{}) {
				body, err := ioutil.ReadAll(recorder.Body)
				require.NoError(t, err)

				var response db.Account
				err = json.Unmarshal(body, &response)
				require.NoError(t, err)

				require.Equal(t, tc.GetAccount.Response.ID, response.ID)
				require.Equal(t, tc.GetAccount.Response.Owner, response.Owner)
				require.Equal(t, tc.GetAccount.Response.Balance, response.Balance)
				require.Equal(t, tc.GetAccount.Response.Currency, response.Currency)
				require.WithinDuration(t, tc.GetAccount.Response.CreatedAt, response.CreatedAt, time.Second)

			}

		})
	}
}

func createRandomAccount() db.Account {
	return db.Account{
		ID:        1,
		Owner:     "owner",
		Balance:   544,
		Currency:  "USD",
		CreatedAt: time.Now(),
	}
}
