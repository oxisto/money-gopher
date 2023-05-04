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

// package assert contains logic to assert test values.
package assert

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

type Want[T any] func(*testing.T, T) bool

// Equals compares expected to actual and returns true if they are equal. If the
// expected type is a protobuf message, [proto.Equals] will be used for
// comparison Otherwise, the test fails (but continues) and false is returned.
func Equals[T comparable](t *testing.T, expected T, actual T) (ok bool) {
	if msg, isMsg := any(expected).(proto.Message); isMsg {
		ok = proto.Equal(msg, any(actual).(proto.Message))
	} else {
		ok = expected == actual
	}

	if !ok {
		t.Errorf("%T = %v, want %v", actual, actual, expected)
	}

	return ok
}

// Is asserts that value is of type T. If it succeeds, it returns the value
// casted to T. If it fails, we fatally fail the test, because we cannot
// continue.
func Is[T any](t *testing.T, value any) T {
	cast, ok := value.(T)
	if !ok {
		// We cannot continue
		t.Fatalf("%v is not of type %T", value, new(T))
	}

	return cast
}

// NoError asserts that err does not contain an error.
func NoError(t *testing.T, err error) bool {
	return Equals(t, nil, err)
}
