package utils

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"
)

func GetDataWithSF[T any](ctx context.Context, sf *singleflight.Group, key string, fn func()(T, error)) (T, error) {
	var zero T

	ch := sf.DoChan(key, func() (interface{}, error) {
		return fn()
	}) 

	select {
	case <- ctx.Done():
		return  zero, ctx.Err()

	case res, ok := <-ch:
		if  !ok {
			return zero, fmt.Errorf("singleflight channel closed")
		}

		if res.Err != nil {
			return zero, res.Err
		}

		return res.Val.(T), nil
	}
}