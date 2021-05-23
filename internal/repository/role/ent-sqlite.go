package role

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuidUtils"
)

type EntSQLRoleRepo struct {
	client   *ent.Client
	uuidTool uuidUtils.UUIDGenerator
}

func NewEntSQLRoleRepo(client *ent.Client, uuidTool uuidUtils.UUIDGenerator) EntSQLRoleRepo {
	return EntSQLRoleRepo{
		client:   client,
		uuidTool: uuidTool,
	}
}

func (e EntSQLRoleRepo) GetUserRoleByFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID, userUUID string) (string, error) {
	panic("implement me")
}

func (e EntSQLRoleRepo) GetUserRoleByDirectory(ctx context.Context, tx transaction.RollbackCommitter, dirUUID, userUUID string) (string, error) {
	panic("implement me")
}
