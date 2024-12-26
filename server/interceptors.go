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

package server

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"connectrpc.com/connect"
	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/lmittmann/tint"
)

// NewSimpleLoggingInterceptor returns a new simple logging interceptor.
func NewSimpleLoggingInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			slog.Debug("Handling RPC Request",
				slog.Group("req",
					"procedure", req.Spec().Procedure,
					"httpmethod", req.HTTPMethod(),
				))
			return next(ctx, req)
		})
	}

	return connect.UnaryInterceptorFunc(interceptor)
}

// NewAuthInterceptor returns a new auth interceptor.
func NewAuthInterceptor() connect.UnaryInterceptorFunc {
	interceptor := func(next connect.UnaryFunc) connect.UnaryFunc {
		k, err := keyfunc.NewDefault([]string{"http://localhost:8000/certs"})
		if err != nil {
			slog.Error("Error while setting up JWKS", tint.Err(err))
		}

		return connect.UnaryFunc(func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			var (
				claims jwt.RegisteredClaims
				auth   string
				token  string
				err    error
				ok     bool
			)
			auth = req.Header().Get("Authorization")
			if auth == "" {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}

			token, ok = strings.CutPrefix(auth, "Bearer ")
			if !ok {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					errors.New("no token provided"),
				)
			}

			_, err = jwt.ParseWithClaims(token, &claims, k.Keyfunc)
			if err != nil {
				return nil, connect.NewError(
					connect.CodeUnauthenticated,
					err,
				)
			}

			ctx = context.WithValue(ctx, "claims", claims)
			return next(ctx, req)
		})
	}
	return connect.UnaryInterceptorFunc(interceptor)
}
