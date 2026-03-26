package redirect_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"url_shorter/internal/http-server/handlers/url/redirect"
	"url_shorter/internal/http-server/handlers/url/redirect/mocks"
	"url_shorter/internal/lib/logger/handlers/slogdiscard"
	"url_shorter/internal/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name           string
		alias          string
		mockURL        string
		mockError      error
		expectedStatus int
		expectedURL    string
	}{
		{
			name:           "Success",
			alias:          "abc123",
			mockURL:        "https://google.com",
			expectedStatus: http.StatusFound,
			expectedURL:    "https://google.com",
		},
		{
			name:           "Not found",
			alias:          "unknown",
			mockError:      storage.ErrUrlNotFound,
			expectedStatus: http.StatusNotFound,
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

			mockGetter := mocks.NewGetURL(t)

			mockGetter.
				On("GetUrl", tc.alias).
				Return(tc.mockURL, tc.mockError).
				Once()

			handler := redirect.New(slogdiscard.NewDiscardLogger(), mockGetter)

			req := httptest.NewRequest(http.MethodGet, "/"+tc.alias, nil)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)

			req = req.WithContext(
				context.WithValue(req.Context(), chi.RouteCtxKey, rctx),
			)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.expectedStatus, rr.Code)

			if tc.expectedURL != "" {
				require.Equal(t, tc.expectedURL, rr.Header().Get("Location"))
			}
		})
	}
}
