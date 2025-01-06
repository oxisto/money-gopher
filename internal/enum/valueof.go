package enum

import (
	"encoding/json"
	"flag"
	"fmt"
)

// Enum is an interface for enum types that support [flag.Value].
type Enum interface {
	flag.Value
}

// ValueOf returns the index of the value v in the name/index slice.
func ValueOf(v string, name string, index []uint8) int {
	for i := range len(index) - 1 {
		if name[index[i]:index[i+1]] == v {
			return i + 1
		}
	}

	return -1
}

// Set sets the target enum to the value represented by v (using [ValueOf]).
func Set[T ~int](enum *T, v string, name string, index []uint8) error {
	i := ValueOf(v, name, index)
	if i == -1 {
		return fmt.Errorf("unknown value: %s", v)
	} else {
		*enum = T(i)
		return nil
	}
}

// MarshalJSON marshals the enum to JSON using the string representation.
func MarshalJSON[T fmt.Stringer](enum T) ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, enum.String())), nil
}

// UnmarshalJSON unmarshals the enum from JSON. It expects a string
// representation.
func UnmarshalJSON[T Enum](enum T, data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return enum.Set(s)
}
