package delete_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	del "url_shorter/internal/http-server/handlers/url/delete"
	"url_shorter/internal/http-server/handlers/url/delete/mocks"
	"url_shorter/internal/lib/logger/handlers/slogdiscard"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestDeleteURLHandler(t *testing.T) {
	cases := []struct {
		name           string
		alias          string
		mockError      error
		expectedStatus int
	}{
		{
			name:           "Success",
			alias:          "abc123",
			mockError:      nil,
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Internal error",
			alias:          "abc123",
			mockError:      errors.New("db error"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mockDeleter := mocks.NewDeleteURL(t)

			mockDeleter.
				On("DeleteURL", tc.alias).
				Return(tc.mockError).
				Once()

			handler := del.New(slogdiscard.NewDiscardLogger(), mockDeleter)

			req := httptest.NewRequest(http.MethodDelete, "/"+tc.alias, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()

			handler(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)
		})
	}
}
