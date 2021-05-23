package transaction

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/pkg/txUtils"
)

type EntSQLTxRepo struct {
	client *ent.Client
}

func (e EntSQLTxRepo) StartTransaction(ctx context.Context) (RollbackCommitter, error) {
	return e.client.Tx(ctx)
}

func (e EntSQLTxRepo) FinishTransaction(tx RollbackCommitter, err error) error {
	if err != nil {
		return txUtils.Rollback(tx, err)
	}
	return tx.Commit()
}
