package commands

import (
	"context"
	"crypto/ecdsa"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/oxisto/money-gopher/gen/portfoliov1connect"
	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/server"
	"github.com/oxisto/money-gopher/service/portfolio"
	"github.com/oxisto/money-gopher/service/securities"

	"connectrpc.com/connect"
	"connectrpc.com/vanguard"
	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"
	"github.com/oxisto/oauth2go/storage"
	"github.com/urfave/cli/v3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var cfg moneydConfig

type moneydConfig struct {
	Debug bool

	EmbeddedOAuth2ServerDashboardCallback string

	PrivateKeyFile     string
	PrivateKeyPassword string
}

var ServerCmd = &cli.Command{
	Name:  "moneyd",
	Usage: "Starts the Money Gopher server.",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "debug", Aliases: []string{"d"},
			Destination: &cfg.Debug},
		&cli.StringFlag{
			Name:        "embedded-oauth2-server-dashboard-callback",
			Value:       "http://localhost:3000/api/auth/callback/money-gopher",
			Usage:       "Specifies the callback URL for the dashboard, if the embedded oauth2 server is used",
			Destination: &cfg.EmbeddedOAuth2ServerDashboardCallback,
		},
		&cli.StringFlag{
			Name:        "private-key-file",
			Value:       "private.key",
			Destination: &cfg.PrivateKeyFile,
		},
		&cli.StringFlag{
			Name:        "private-key-password",
			Value:       "moneymoneymoney",
			Destination: &cfg.PrivateKeyPassword,
		},
	},
	Action: RunServer,
}

func RunServer(ctx context.Context, cmd *cli.Command) error {
	var (
		w       = os.Stdout
		level   = slog.LevelInfo
		authSrv *oauth2.AuthorizationServer
	)

	if cfg.Debug {
		level = slog.LevelDebug
	}

	logger := slog.New(
		tint.NewHandler(colorable.NewColorable(w), &tint.Options{
			TimeFormat: time.TimeOnly,
			Level:      level,
			NoColor:    !isatty.IsTerminal(w.Fd()),
		}),
	)

	slog.SetDefault(logger)
	slog.Info("Welcome to the Money Gopher", "money", "🤑")

	pdb, err := persistence.OpenDB(persistence.Options{})
	if err != nil {
		slog.Error("Error while opening database", tint.Err(err))
		return err
	}

	authSrv = oauth2.NewServer(
		":8000",
		oauth2.WithClient("dashboard", "", cfg.EmbeddedOAuth2ServerDashboardCallback),
		oauth2.WithClient("cli", "", "http://localhost:10000/callback"),
		oauth2.WithPublicURL("http://localhost:8000"),
		login.WithLoginPage(
			login.WithUser("money", "money"),
		),
		oauth2.WithAllowedOrigins("*"),
		oauth2.WithSigningKeysFunc(func() map[int]*ecdsa.PrivateKey {
			return storage.LoadSigningKeys(cfg.PrivateKeyFile, cfg.PrivateKeyPassword, true)
		}),
	)
	go authSrv.ListenAndServe()

	interceptors := connect.WithInterceptors(
		server.NewSimpleLoggingInterceptor(),
		server.NewAuthInterceptor(),
	)

	portfolioService := vanguard.NewService(
		portfoliov1connect.NewPortfolioServiceHandler(portfolio.NewService(
			portfolio.Options{
				DB:               pdb,
				SecuritiesClient: portfoliov1connect.NewSecuritiesServiceClient(http.DefaultClient, portfolio.DefaultSecuritiesServiceURL),
			},
		), interceptors))
	securitiesService := vanguard.NewService(
		portfoliov1connect.NewSecuritiesServiceHandler(securities.NewService(pdb), interceptors),
	)

	transcoder, err := vanguard.NewTranscoder([]*vanguard.Service{
		portfolioService,
		securitiesService,
	}, vanguard.WithCodec(func(tr vanguard.TypeResolver) vanguard.Codec {
		codec := vanguard.NewJSONCodec(tr)
		codec.MarshalOptions.EmitDefaultValues = true
		return codec
	}))
	if err != nil {
		slog.Error("transcoder failed", tint.Err(err))
		return err
	}

	mux := http.NewServeMux()
	mux.Handle("/", transcoder)

	err = http.ListenAndServe(
		":8080",
		h2c.NewHandler(server.HandleCORS(mux), &http2.Server{}),
	)

	slog.Error("listen failed", tint.Err(err))

	return err
}
