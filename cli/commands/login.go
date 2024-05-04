package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/oxisto/money-gopher/cli"
	oauth2 "github.com/oxisto/oauth2go"
)

type LoginCmd struct {
	ClientID string `default:"cli"`
	AuthURL  string `default:"http://localhost:8000/authorize"`
	TokenURL string `default:"http://localhost:8000/token"`
	Callback string `default:"http://localhost:10000/callback"`
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

func (l *LoginCmd) Run(s *cli.Session) error {
	var (
		err     error
		session *cli.Session
		sock    net.Listener
		code    string
		config  *oauth2.Config
	)

	// Create an OAuth 2 config. TODO: Use oauth2 metadata discovery instead
	config = &oauth2.Config{
		ClientID: l.ClientID,
		Endpoint: oauth2.Endpoint{
			AuthURL:  l.AuthURL,
			TokenURL: l.TokenURL,
		},
		RedirectURL: l.Callback,
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

	session = cli.NewSession(&cli.SessionOptions{
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
