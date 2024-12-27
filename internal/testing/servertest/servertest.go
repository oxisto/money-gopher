package servertest

import (
	"net/http"
	"net/http/httptest"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/server"
	"github.com/oxisto/money-gopher/service/portfolio"
	"github.com/oxisto/money-gopher/service/securities"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewServer(db *persistence.DB) *httptest.Server {
	mux := http.NewServeMux()
	srv := httptest.NewServer(h2c.NewHandler(mux, &http2.Server{}))
	server.ConfigureGraphQL(mux, db)

	mux.Handle(portfoliov1connect.NewPortfolioServiceHandler(portfolio.NewService(
		portfolio.Options{
			DB:               db,
			SecuritiesClient: portfoliov1connect.NewSecuritiesServiceClient(srv.Client(), srv.URL),
		},
	)))
	mux.Handle(portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService(db)))

	return srv
}
