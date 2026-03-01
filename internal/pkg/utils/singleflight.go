package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/singleflight"
)

var limitSem = make(chan struct{}, 100)
func GetDataWithSF[T any](ctx context.Context, sf *singleflight.Group, key string, fn func(ctx context.Context)(T, error)) (T, error) {
	var zero T

	select {
	case limitSem <- struct{}{}:
		defer func() { <- limitSem}()
	case <-ctx.Done():
		return zero, ctx.Err()
	}

	innerCtx, cancel := context.WithTimeout(ctx, 1* time.Second)
	defer cancel()

	ch := sf.DoChan(key, func() (interface{}, error) {
		return fn(innerCtx)
	}) 

	select {
	case <- innerCtx.Done():
		if errors.Is(innerCtx.Err(), context.DeadlineExceeded) {
			return zero, fmt.Errorf("AITA SF timeout: %w", innerCtx.Err())
		}
		return  zero, innerCtx.Err()

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