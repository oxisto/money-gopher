package commands

import (
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/oxisto/assert"
	oauth2 "github.com/oxisto/oauth2go"
	"github.com/oxisto/oauth2go/login"
)

func TestLoginCmd(t *testing.T) {
	var (
		err      error
		verifier string
		authSrv  *oauth2.AuthorizationServer
		port     uint16
		timeout  <-chan time.Time
	)

	authSrv, port, err = startAuthServer()
	assert.NoError(t, err)

	cmd := &LoginCmd{
		ClientID: "cli",
		AuthURL:  fmt.Sprintf("http://localhost:%d/authorize", port),
		TokenURL: fmt.Sprintf("http://localhost:%d/token", port),
		Callback: "http://localhost:10000/callback",
	}

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
			_, err = http.Get(fmt.Sprintf("%s?code=%s", cmd.Callback, code))
			if err != nil {
				assert.NoError(t, err)
			}
		}()

		err = cmd.Run(nil)
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
