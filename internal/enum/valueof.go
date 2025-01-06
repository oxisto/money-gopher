package enum

import "fmt"

// ValueOf returns the index of the value v in the name/index slice.
func ValueOf(v string, name string, index []uint8) int {
	for i := range len(index) - 1 {
		if name[index[i]:index[i+1]] == v {
			return i + 1
		}
	}

	return -1
}

// Set sets the target to the value represented by v (using [ValueOf]).
func Set[T ~int](target *T, v string, name string, index []uint8) error {
	i := ValueOf(v, name, index)
	if i == -1 {
		return fmt.Errorf("unknown value: %s", v)
	} else {
		*target = T(i)
		return nil
	}
}
