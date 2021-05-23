package file

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/ent/directory"
	"github.com/nvthongswansea/xtreme/internal/ent/file"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuidUtils"
)

type EntSQLFileRepo struct {
	client   *ent.Client
	uuidTool uuidUtils.UUIDGenerator
}

func NewEntSQLFileRepo(client *ent.Client, uuidTool uuidUtils.UUIDGenerator) EntSQLFileRepo {
	return EntSQLFileRepo{
		client:   client,
		uuidTool: uuidTool,
	}
}

func (e EntSQLFileRepo) InsertFile(ctx context.Context, tx transaction.RollbackCommitter, newFile models.File) (string, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	f, err := client.File.
		Create().
		SetID(e.uuidTool.NewUUID()).
		SetName(newFile.Metadata.Filename).
		SetOwnerID(newFile.Metadata.OwnerUUID).
		SetParentID(newFile.Metadata.ParentUUID).
		Save(ctx)
	if err != nil {
		return "", err
	}
	return f.ID, nil
}

func (e EntSQLFileRepo) GetFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) (models.File, error) {
	panic("implement me")
}

func (e EntSQLFileRepo) GetFileMetadata(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) (models.FileMetadata, error) {
	panic("implement me")
}

func (e EntSQLFileRepo) GetFileMetadataBatch(ctx context.Context, tx transaction.RollbackCommitter, fileUUIDs []string) ([]models.FileMetadata, error) {
	panic("implement me")
}

func (e EntSQLFileRepo) GetFileMetadataListByDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) ([]models.FileMetadata, error) {
	panic("implement me")
}

func (e EntSQLFileRepo) GetFileMetadataListByName(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, filename string) ([]models.FileMetadata, error) {
	panic("implement me")
}

func (e EntSQLFileRepo) IsFilenameExist(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, name string) (bool, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	count, err := client.File.Query().
		Where(
			file.And(
				file.HasParentWith(directory.ID(parentDirUUID)),
				file.Name(name),
			),
		).
		Count(ctx)
	if err != nil {
		return false, err
	}
	if count != 0 {
		return true, err
	}
	return false, err
}

func (e EntSQLFileRepo) UpdateFilename(ctx context.Context, tx transaction.RollbackCommitter, newFilename, fileUUID string) error {
	panic("implement me")
}

func (e EntSQLFileRepo) UpdateParentDirUUID(ctx context.Context, tx transaction.RollbackCommitter, newParentDirUUID, fileUUID string) error {
	panic("implement me")
}

func (e EntSQLFileRepo) SoftRemoveFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) error {
	panic("implement me")
}

func (e EntSQLFileRepo) HardRemoveFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string, rmFileFn func(string) error) error {
	panic("implement me")
}
