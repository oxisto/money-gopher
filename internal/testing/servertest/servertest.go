package servertest

import (
	"net/http"
	"net/http/httptest"

	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/securities/quote"
	"github.com/oxisto/money-gopher/server"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func NewServer(db *persistence.DB) *httptest.Server {
	mux := http.NewServeMux()
	srv := httptest.NewServer(h2c.NewHandler(mux, &http2.Server{}))

	qu := quote.NewQuoteUpdater(db)

	server.ConfigureGraphQL(mux, db, qu)

	return srv
}
