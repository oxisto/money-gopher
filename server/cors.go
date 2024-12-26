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

// package server provides utilities for handling HTTP requests.
package server

import (
	"net/http"
	"strings"
)

// handleCORS is a middleware that handles CORS headers.
func handleCORS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Add("Vary", "Origin")

		if r.Method == "OPTIONS" && r.Header.Get("Access-Control-Request-Method") != "" {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
				"Connect-Protocol-Version",
				"Content-Type",
				"Accept",
				"Authorization",
			}, ","))
			w.Header().Set("Access-Control-Allow-Methods", strings.Join([]string{
				"GET",
				"POST",
				"PUT",
				"DELETE",
			}, ","))
		} else {
			h.ServeHTTP(w, r)
		}
	})
}
