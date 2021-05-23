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

func (e EntSQLFileRepo) InsertFile(ctx context.Context, tx transaction.RollbackCommitter, newFile models.File, saveFileFn SaveFileToDiskFn) (string, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	newUUID := e.uuidTool.NewUUID()
	relPathOD := newUUID
	size, err := saveFileFn(relPathOD)
	if err != nil {
		return "", err
	}
	f, err := client.File.
		Create().
		SetID(newUUID).
		SetName(newFile.Metadata.Filename).
		SetOwnerID(newFile.Metadata.OwnerUUID).
		SetParentID(newFile.Metadata.ParentUUID).
		SetRelPathOnDisk(relPathOD).
		SetSize(size).
		SetPath(newFile.Metadata.Path).
		Save(ctx)
	if err != nil {
		return "", err
	}
	return f.ID, nil
}

func (e EntSQLFileRepo) GetFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) (models.File, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	f, err := client.File.
		Query().
		Where(file.ID(fileUUID)).
		WithOwner().
		WithParent().
		First(ctx)
	if err != nil {
		return models.File{}, err
	}
	return models.File{
		Metadata:    models.FileMetadata{
			UUID:          f.ID,
			Filename:      f.Name,
			MIMEType:      f.MimeType,
			Path:          f.Path,
			RelPathOnDisk: f.RelPathOnDisk,
			ParentUUID:    f.Edges.Parent.ID,
			Size:          f.Size,
			OwnerUUID:     f.Edges.Owner.ID,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		},
	}, nil
}

func (e EntSQLFileRepo) GetFileMetadata(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) (models.FileMetadata, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	f, err := client.File.
		Query().
		Where(file.ID(fileUUID)).
		First(ctx)
	if err != nil {
		return models.FileMetadata{}, err
	}
	return models.FileMetadata{
			UUID:          f.ID,
			Filename:      f.Name,
			MIMEType:      f.MimeType,
			Path:          f.Path,
			RelPathOnDisk: f.RelPathOnDisk,
			Size:          f.Size,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		}, nil
}

func (e EntSQLFileRepo) GetFileMetadataBatch(ctx context.Context, tx transaction.RollbackCommitter, fileUUIDs []string) ([]models.FileMetadata, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	fileList, err := client.File.
		Query().
		Where(file.IDIn(fileUUIDs...)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var fileMetadataList []models.FileMetadata
	for _, f := range fileList {
		fileMetadataList = append(fileMetadataList, models.FileMetadata{
			UUID:          f.ID,
			Filename:      f.Name,
			MIMEType:      f.MimeType,
			Path:          f.Path,
			RelPathOnDisk: f.RelPathOnDisk,
			Size:          f.Size,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		})
	}
	return fileMetadataList, nil
}

func (e EntSQLFileRepo) GetFileMetadataListByDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) ([]models.FileMetadata, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	fileList, err := client.File.
		Query().
		Where(file.HasParentWith(directory.ID(dirUUID))).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var fileMetadataList []models.FileMetadata
	for _, f := range fileList {
		fileMetadataList = append(fileMetadataList, models.FileMetadata{
			UUID:          f.ID,
			Filename:      f.Name,
			MIMEType:      f.MimeType,
			Path:          f.Path,
			RelPathOnDisk: f.RelPathOnDisk,
			Size:          f.Size,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		})
	}
	return fileMetadataList, nil
}

func (e EntSQLFileRepo) GetFileMetadataListByName(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, filename string) ([]models.FileMetadata, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	fileList, err := client.File.Query().
		Where(file.And(
			file.HasParentWith(directory.ID(parentDirUUID)),
			file.NameContains(filename),
		)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var fileMetadataList []models.FileMetadata
	for _, f := range fileList {
		fileMetadataList = append(fileMetadataList, models.FileMetadata{
			UUID:          f.ID,
			Filename:      f.Name,
			MIMEType:      f.MimeType,
			Path:          f.Path,
			RelPathOnDisk: f.RelPathOnDisk,
			Size:          f.Size,
			CreatedAt:     f.CreatedAt,
			UpdatedAt:     f.UpdatedAt,
		})
	}
	return fileMetadataList, nil
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
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	err := client.File.UpdateOneID(fileUUID).
		SetName(newFilename).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e EntSQLFileRepo) UpdateFileRelPathOD(ctx context.Context, tx transaction.RollbackCommitter, relPathOD, fileUUID string) error {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	err := client.File.UpdateOneID(fileUUID).
		SetRelPathOnDisk(relPathOD).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e EntSQLFileRepo) UpdateFileSize(ctx context.Context, tx transaction.RollbackCommitter, size int64, fileUUID string) error {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	err := client.File.UpdateOneID(fileUUID).
		SetSize(size).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e EntSQLFileRepo) UpdateParentDirUUID(ctx context.Context, tx transaction.RollbackCommitter, newParentDirUUID, fileUUID string) error {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	err := client.File.UpdateOneID(fileUUID).
		SetParentID(newParentDirUUID).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e EntSQLFileRepo) SoftRemoveFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string) error {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	err := client.File.UpdateOneID(fileUUID).
		SetIsDeleted(true).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (e EntSQLFileRepo) HardRemoveFile(ctx context.Context, tx transaction.RollbackCommitter, fileUUID string, rmFileFn func(string) error) error {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	err := client.File.DeleteOneID(fileUUID).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}
