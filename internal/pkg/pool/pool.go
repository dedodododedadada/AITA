package pool

import (
	"log/slog"
	"time"

	"github.com/panjf2000/ants/v2"
)

func NewBackfillPool(size int) (*ants.Pool, error) {
	p, err := ants.NewPool(size,
		ants.WithPreAlloc(true),
		ants.WithExpiryDuration(10 * time.Second),
		ants.WithPanicHandler(func (i any) {
			slog.Error("ルーチンプール内でパニックが発生しました", "err", i) 
		}),
	)
	
	if err != nil {
		slog.Error("ルーチンプールの初期化に失敗しました", "エラー", err)
		return nil, err
	}

	slog.Info("ルーチンプールの初期化が正常に完了しました", "プールサイズ", size)
	return p, nil
}