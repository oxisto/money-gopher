package commands

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	mcli "github.com/oxisto/money-gopher/cli"

	"github.com/oxisto/assert"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"
)

func TestLoginAction(t *testing.T) {
	var (
		err      error
		verifier string
		authSrv  *oauth2.AuthorizationServer
		port     uint16
		timeout  <-chan time.Time
	)

	authSrv, port, err = startAuthServer()
	assert.NoError(t, err)

	var (
		clientID = "cli"
		authURL  = fmt.Sprintf("http://localhost:%d/authorize", port)
		tokenURL = fmt.Sprintf("http://localhost:%d/token", port)
		callback = "http://localhost:10000/callback"
	)
	cmd := LoginCmd

	verifier = "012345678901234567890123456789"
	VerifierGenerator = func() string {
		return verifier
	}

	code := authSrv.IssueCode(oauth2.GenerateCodeChallenge(verifier))

	// Simulate a callback with a timeout of 5 seconds
	timeout = time.After(5 * time.Second)
	done := make(chan bool)
	go func() {
		go func() {
			<-callbackServerReady
			_, err = http.Get(fmt.Sprintf("%s?code=%s", callback, code))
			if err != nil {
				assert.NoError(t, err)
			}
		}()

		err = cmd.Run(context.Background(), []string{
			"login",
			"--client-id", clientID,
			"--auth-url", authURL,
			"--token-url", tokenURL,
			"--callback", callback,
		})
		assert.NoError(t, err)

		// Resume the session
		_, err := mcli.ContinueSession()
		assert.NoError(t, err)

		done <- true
	}()

	select {
	case <-timeout:
		t.Fatal("Did not finish in time")
	case <-done:
	}
}

func startAuthServer() (srv *oauth2.AuthorizationServer, port uint16, err error) {
	var (
		nl net.Listener
	)

	nl, err = net.Listen("tcp", "localhost:0")
	if err != nil {
		return nil, 0, fmt.Errorf("could not listen: %w", err)
	}

	port = nl.Addr().(*net.TCPAddr).AddrPort().Port()

	srv = oauth2.NewServer(fmt.Sprintf(":%d", port),
		oauth2.WithClient("cli", "", "http://localhost:10000/callback"),
		oauth2.WithPublicURL(fmt.Sprintf("http://localhost:%d", port)),
		login.WithLoginPage(
			login.WithUser("money", "money"),
		),
	)

	go func() {
		_ = srv.Serve(nl)
	}()

	return
}
