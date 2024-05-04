package portfoliotest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/internal"
	"github.com/oxisto/money-gopher/service/portfolio"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewServer(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()
	mux.Handle(portfoliov1connect.NewPortfolioServiceHandler(portfolio.NewService(
		portfolio.Options{
			DB:               internal.NewTestDB(t),
			SecuritiesClient: portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, portfolio.DefaultSecuritiesServiceURL),
		},
	)))

	return httptest.NewServer(h2c.NewHandler(mux, &http2.Server{}))
}
