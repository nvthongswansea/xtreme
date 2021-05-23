package role

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/ent/directory"
	"github.com/nvthongswansea/xtreme/internal/ent/file"
	"github.com/nvthongswansea/xtreme/internal/ent/user"
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
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	count, err := client.File.
		Query().
		Where(file.And(
			file.ID(fileUUID),
			file.HasOwnerWith(user.ID(userUUID)),
			)).
		Count(ctx)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", nil
	}
	return OwnerRol, nil
}

func (e EntSQLRoleRepo) GetUserRoleByDirectory(ctx context.Context, tx transaction.RollbackCommitter, dirUUID, userUUID string) (string, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	count, err := client.Directory.
		Query().
		Where(directory.And(
			directory.ID(dirUUID),
			directory.HasOwnerWith(user.ID(userUUID)),
		)).
		Count(ctx)
	if err != nil {
		return "", err
	}
	if count == 0 {
		return "", nil
	}
	return OwnerRol, nil
}
