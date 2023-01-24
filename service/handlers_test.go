package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"wdiet/store/mockstore"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

var l, _ = zap.NewProduction()

var testServer = New(&mockstore.Mockstore{}, l)

func TestPing(t *testing.T) {
	testcases := []struct {
		name                   string
		pingOverrideFunc       func() error
		expectedResponseBody   string
		expectedResponseStatus int
	}{
		{
			"happyPath",
			nil,
			`{"message":"pong"}`,
			http.StatusOK,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/ping", nil)
			w := httptest.NewRecorder()

			testServer.db = &mockstore.Mockstore{PingOverride: testcase.pingOverrideFunc}
			testServer.r.ServeHTTP(w, req)

			assert.Equal(t, testcase.expectedResponseStatus, w.Code)
			assert.Equal(t, testcase.expectedResponseBody, w.Body.String())
		})
	}
}
