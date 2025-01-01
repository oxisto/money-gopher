// Copyright 2024 Christian Banse
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// This file is part of The Money Gopher.

package commands

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/oxisto/money-gopher/persistence"
	"github.com/oxisto/money-gopher/server"

	"github.com/lmittmann/tint"
	"github.com/mattn/go-colorable"
	"github.com/mattn/go-isatty"
	"github.com/urfave/cli/v3"
)

// opts holds the options for the server.
var opts server.Options

// ServerCmd is the command to start the Money Gopher server.
var ServerCmd = &cli.Command{
	Name:  "moneyd",
	Usage: "Starts the Money Gopher server.",
	Flags: []cli.Flag{
		&cli.BoolFlag{Name: "debug", Aliases: []string{"d"},
			Destination: &opts.Debug},
		&cli.StringFlag{
			Name:        "embedded-oauth2-server-dashboard-callback",
			Value:       "http://localhost:3000/api/auth/callback/money-gopher",
			Usage:       "Specifies the callback URL for the dashboard, if the embedded oauth2 server is used",
			Destination: &opts.EmbeddedOAuth2ServerDashboardCallback,
		},
		&cli.StringFlag{
			Name:        "private-key-file",
			Value:       "private.key",
			Destination: &opts.PrivateKeyFile,
		},
		&cli.StringFlag{
			Name:        "private-key-password",
			Value:       "moneymoneymoney",
			Destination: &opts.PrivateKeyPassword,
		},
	},
	Action: RunServer,
}

// RunServer is the action for the server command.
func RunServer(ctx context.Context, cmd *cli.Command) error {
	var (
		w     = os.Stdout
		level = slog.LevelInfo
	)

	if opts.Debug {
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

	db, err := persistence.OpenDB(persistence.Options{})
	if err != nil {
		slog.Error("Error while opening database", tint.Err(err))
		return err
	}

	return server.StartServer(db, opts)
}
