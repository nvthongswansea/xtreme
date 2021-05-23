package author

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
)

type StubAuthorizer struct {
}

func (s StubAuthorizer) AuthorizeActionsOnFile(ctx context.Context, tx transaction.RollbackCommitter, userUUID, fileUUID string, actions ...fileAction) (bool, error) {
	return true, nil
}

func (s StubAuthorizer) AuthorizeActionsOnDir(ctx context.Context, tx transaction.RollbackCommitter, userUUID, dirUUID string, actions ...dirAction) (bool, error) {
	return true, nil
}
