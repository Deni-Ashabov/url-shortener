package tests

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"url_shorter/internal/http-server/handlers/url/save"
	"url_shorter/internal/lib/api"
	"url_shorter/internal/lib/random"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/require"
)

const (
	host = "localhost:8082"
)

func TestMain(m *testing.M) {
	wd, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	envPath := filepath.Join(wd, "..", ".env")

	if _, err := os.Stat(envPath); os.IsNotExist(err) {
		fmt.Println(".env not found at", envPath)
	} else {
		_ = godotenv.Load(envPath)
	}

	os.Exit(m.Run())
}

func TestURLShortener_HappyPath(t *testing.T) {
	godotenv.Load(".env")

	u := url.URL{
		Scheme: "http",
		Host:   host,
	}
	e := httpexpect.Default(t, u.String())

	e.POST("/url").
		WithJSON(save.Request{
			URL:   gofakeit.URL(),
			Alias: random.NewRandomString(10),
		}).
		WithBasicAuth(
			os.Getenv("HTTP_SERVER_USER"),
			os.Getenv("HTTP_SERVER_PASSWORD"),
		).
		Expect().
		Status(200).
		JSON().Object().
		ContainsKey("alias")
}

func TestURLShortener_SaveRedirect(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		alias        string
		error        string
		expectDelete bool
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Invalid URL",
			url:   "invalid_url",
			alias: gofakeit.Word(),
			error: "field URL is not a valid URL",
		},
		{
			name:  "Empty Alias",
			url:   gofakeit.URL(),
			alias: "",
		},
		{
			name:         "Delete non-existing alias",
			url:          gofakeit.URL(),
			alias:        gofakeit.Word(),
			expectDelete: false,
		},
		{
			name:  "Invalid Redirect",
			url:   gofakeit.URL(),
			alias: gofakeit.Word(),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// Save

			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth(
					os.Getenv("HTTP_SERVER_USER"),
					os.Getenv("HTTP_SERVER_PASSWORD"),
				).
				Expect().Status(http.StatusOK).
				JSON().Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")

				resp.Value("error").String().IsEqual(tc.error)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(tc.alias)
			} else {
				resp.Value("alias").String().NotEmpty()

				alias = resp.Value("alias").String().Raw()
			}

			// Redirect

			testRedirect(t, alias, tc.url)

			// Delete

			e.DELETE("/{alias}", alias).
				WithBasicAuth(
					os.Getenv("HTTP_SERVER_USER"),
					os.Getenv("HTTP_SERVER_PASSWORD"),
				).
				Expect().
				Status(http.StatusNoContent)

			u = url.URL{
				Scheme: "http",
				Host:   host,
				Path:   alias,
			}
			_, err := api.GetRedirect(u.String())

			require.Error(t, err)

			// Redirect failed

			testRedirectNotFound(t, alias)
		})
	}
}

func testRedirect(t *testing.T, alias string, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedToURL, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToURL)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	_, err := api.GetRedirect(u.String())

	require.ErrorIs(t, err, api.ErrInvalidStatusCode)
}
