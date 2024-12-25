package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	mcli "github.com/oxisto/money-gopher/cli"
	oauth2 "github.com/oxisto/oauth2go"

	"github.com/urfave/cli/v3"
)

var LoginCmd = &cli.Command{
	Name:   "login",
	Usage:  "Login to the Money Gopher server",
	Action: LoginAction,
	Flags: []cli.Flag{
		&cli.StringFlag{Name: "client-id", Usage: "The client ID to use for the OAuth 2.0 flow", Value: "cli"},
		&cli.StringFlag{Name: "auth-url", Usage: "The authorization URL for the OAuth 2.0 flow", Value: "http://localhost:8000/authorize"},
		&cli.StringFlag{Name: "token-url", Usage: "The token URL for the OAuth 2.0 flow", Value: "http://localhost:8000/token"},
		&cli.StringFlag{Name: "callback", Usage: "The callback URL for the OAuth 2.0 flow", Value: "http://localhost:10000/callback"},
	},
}

var (
	// VerifierGenerator is a function that generates a new verifier.
	VerifierGenerator = oauth2.GenerateSecret

	// callbackServerReady is an internally used channel to indicate that the callback server is ready.
	callbackServerReady = make(chan bool)
)

type callbackServer struct {
	http.Server

	verifier string
	config   *oauth2.Config
	code     chan string
}

func LoginAction(ctx context.Context, cmd *cli.Command) error {
	var (
		err     error
		session *mcli.Session
		sock    net.Listener
		code    string
		config  *oauth2.Config
	)

	// Create an OAuth 2 config. TODO: Use oauth2 metadata discovery instead
	config = &oauth2.Config{
		ClientID: cmd.String("client-id"),
		Endpoint: oauth2.Endpoint{
			AuthURL:  cmd.String("auth-url"),
			TokenURL: cmd.String("token-url"),
		},
		RedirectURL: cmd.String("callback"),
	}

	srv := newCallbackServer(config)

	go func() {
		sock, err = net.Listen("tcp", srv.Addr)
		if err != nil {
			fmt.Printf("Could not start web server for OAuth 2.0 authorization code flow: %v", err)
		}
		go func() {
			callbackServerReady <- true
		}()

		err = srv.Serve(sock)
		if err != http.ErrServerClosed {
			fmt.Printf("Could not start web server for OAuth 2.0 authorization code flow: %v", err)
			return
		}
	}()
	defer srv.Close()

	// waiting for our code
	code = <-srv.code
	token, err := srv.config.Exchange(context.Background(), code,
		oauth2.SetAuthURLParam("code_verifier", srv.verifier),
	)

	if err != nil {
		return err
	}

	session = mcli.NewSession(&mcli.SessionOptions{
		OAuth2Config: config,
		Token:        token,
	})

	if err = session.Save(); err != nil {
		return fmt.Errorf("could not save session: %w", err)
	}

	fmt.Print("\nLogin successful\n")

	return err
}

func newCallbackServer(config *oauth2.Config) *callbackServer {
	var mux = http.NewServeMux()

	var srv = &callbackServer{
		Server: http.Server{
			Handler:           mux,
			Addr:              "localhost:10000",
			ReadHeaderTimeout: 2 * time.Second,
		},
		verifier: VerifierGenerator(),
		config:   config,
		code:     make(chan string),
	}

	mux.HandleFunc("/callback", srv.handleCallback)

	challenge := oauth2.GenerateCodeChallenge(srv.verifier)
	authURL := srv.config.AuthCodeURL("",
		oauth2.SetAuthURLParam("code_challenge", challenge),
		oauth2.SetAuthURLParam("code_challenge_method", "S256"),
	)

	fmt.Printf("Please open %s in your browser 🤑 to continue\n", authURL)

	return srv
}

func (srv *callbackServer) handleCallback(w http.ResponseWriter, r *http.Request) {
	var err error

	_, err = w.Write([]byte("Success. You can close this browser tab now"))
	if err != nil {
		w.WriteHeader(500)
	}

	srv.code <- r.URL.Query().Get("code")
}
