// Copyright 2023 Christian Banse
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

package moneygopher

func Ref[T any](value T) *T {
	return &value
}

func Map[K comparable, V any](slice []V, key func(V) K) (m map[K]V) {
	m = make(map[K]V)

	for _, v := range slice {
		m[key(v)] = v
	}

	return
}
