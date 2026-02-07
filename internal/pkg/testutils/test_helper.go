package testutils

import "github.com/stretchr/testify/mock"

func SafeGet[T any](args mock.Arguments, index int) *T {
	val := args.Get(index)
	if val == nil {
		return nil
	}
	return val.(*T)
}