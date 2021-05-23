package transaction

import (
	"context"
	"github.com/nvthongswansea/xtreme/pkg/txUtils"
)

type RollbackCommitter interface {
	Commit() error
	txUtils.Rollbacker
}

type TxRepository interface {
	StartTransaction(ctx context.Context) (RollbackCommitter, error)
	FinishTransaction(tx RollbackCommitter, err error) error
}
