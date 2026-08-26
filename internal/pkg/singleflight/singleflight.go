package sf

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

	if err := ctx.Err(); err != nil {
		return zero, err
	}

	ch := sf.DoChan(key, func() (any, error) {
		execCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5 * time.Second)
		defer cancel()
		select {
		case limitSem <- struct{}{}:
			defer func() {<- limitSem}()
		case <- execCtx.Done():
			return zero, execCtx.Err()
		}
		return fn(execCtx)
	}) 

	select {
	case <- ctx.Done():
		err := ctx.Err()
		if errors.Is(err, context.DeadlineExceeded) {
			return zero, fmt.Errorf("AITA SF timeout: %w", err)
		}
		return  zero, fmt.Errorf("AITA SF closed: %w", err)

	case res, ok := <-ch:
		if  !ok {
			return zero, fmt.Errorf("singleflight channel closed")
		}

		if res.Err != nil {
			return zero, res.Err
		}

		val, ok := res.Val.(T)
		if !ok {
			return zero, fmt.Errorf("AITA SF: type assertion failed, expected %T", zero)
        }
		return val, nil
	}
}