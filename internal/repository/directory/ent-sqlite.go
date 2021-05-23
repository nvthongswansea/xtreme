package directory

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/ent"
	"github.com/nvthongswansea/xtreme/internal/ent/directory"
	"github.com/nvthongswansea/xtreme/internal/ent/user"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuidUtils"
)

type EntSQLDirectoryRepo struct {
	client   *ent.Client
	uuidTool uuidUtils.UUIDGenerator
}

func NewEntSQLDirectoryRepo(client *ent.Client, uuidTool uuidUtils.UUIDGenerator) EntSQLDirectoryRepo {
	return EntSQLDirectoryRepo{
		client:   client,
		uuidTool: uuidTool,
	}
}

func (e EntSQLDirectoryRepo) InsertRootDirectory(ctx context.Context, tx transaction.RollbackCommitter, userUUID string) (string, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	root, err := client.Directory.
		Create().
		SetID(e.uuidTool.NewUUID()).
		SetName("root").
		SetPath("/").
		SetOwnerID(userUUID).
		Save(ctx)
	if err != nil {
		return "", err
	}
	return root.ID, nil
}

func (e EntSQLDirectoryRepo) InsertDirectory(ctx context.Context, tx transaction.RollbackCommitter, newDir models.Directory) (string, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	dir, err := client.Directory.
		Create().
		SetID(e.uuidTool.NewUUID()).
		SetName(newDir.Metadata.Dirname).
		SetOwnerID(newDir.Metadata.OwnerUUID).
		SetParentID(newDir.Metadata.ParentUUID).
		Save(ctx)
	if err != nil {
		return "", err
	}
	return dir.ID, nil
}

func (e EntSQLDirectoryRepo) GetDirectory(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) (models.Directory, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	dir, err := client.Directory.Query().
		Where(directory.ID(dirUUID)).
		WithOwner().
		WithParent().
		WithChildDirs().
		WithChildFiles().
		First(ctx)
	if err != nil {
		return models.Directory{}, err
	}
	var fileMetadataList []models.FileMetadata
	if dir.Edges.ChildFiles != nil {
		for _, child := range dir.Edges.ChildFiles {
			fileMetadataList = append(fileMetadataList, models.FileMetadata{
				UUID:          child.ID,
				Filename:      child.Name,
				MIMEType:      child.MimeType,
				Path:          child.Path,
				RelPathOnDisk: child.RelPathOnDisk,
				Size:          int64(child.Size),
				CreatedAt:     child.CreatedAt,
				UpdatedAt:     child.UpdatedAt,
			})
		}
	}
	var dirMetadataList []models.DirectoryMetadata
	if dir.Edges.ChildDirs != nil {
		for _, child := range dir.Edges.ChildDirs {
			dirMetadataList = append(dirMetadataList, models.DirectoryMetadata{
				UUID:      child.ID,
				Dirname:   child.Name,
				Path:      child.Path,
				CreatedAt: child.CreatedAt,
				UpdatedAt: child.UpdatedAt,
			})
		}
	}
	return models.Directory{
		Metadata: models.DirectoryMetadata{
			UUID:       dir.ID,
			Dirname:    dir.Name,
			Path:       dir.Path,
			ParentUUID: dir.Edges.Parent.ID,
			OwnerUUID:  dir.Edges.Owner.ID,
			CreatedAt:  dir.CreatedAt,
			UpdatedAt:  dir.UpdatedAt,
		},
		Content: models.DirectoryContent{
			ListOfFiles: fileMetadataList,
			ListOfDirs:  dirMetadataList,
		},
	}, nil
}

func (e EntSQLDirectoryRepo) GetDirMetadata(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) (models.DirectoryMetadata, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	dir, err := client.Directory.Query().
		Where(directory.ID(dirUUID)).
		WithOwner().
		WithParent().
		First(ctx)
	if err != nil {
		return models.DirectoryMetadata{}, err
	}
	metadata := models.DirectoryMetadata{
		UUID:      dir.ID,
		Dirname:   dir.Name,
		Path:      dir.Path,
		OwnerUUID: dir.Edges.Owner.ID,
		CreatedAt: dir.CreatedAt,
		UpdatedAt: dir.UpdatedAt,
	}
	if dir.Edges.Parent != nil {
		metadata.ParentUUID = dir.Edges.Parent.ID
	}
	return metadata, nil
}

func (e EntSQLDirectoryRepo) GetDirMetadataListByName(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, dirname string) ([]models.DirectoryMetadata, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	dirs, err := client.Directory.Query().
		Where(directory.And(
			directory.HasParentWith(directory.ID(parentDirUUID)),
			directory.NameContains(dirname),
			)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	var matchedDirs []models.DirectoryMetadata
	for _, dir := range dirs {
		metadata := models.DirectoryMetadata{
			UUID:      dir.ID,
			Dirname:   dir.Name,
			Path:      dir.Path,
			CreatedAt: dir.CreatedAt,
			UpdatedAt: dir.UpdatedAt,
		}
		matchedDirs = append(matchedDirs, metadata)
	}
	return matchedDirs, nil
}

func (e EntSQLDirectoryRepo) GetRootDirectoryByUserUUID(ctx context.Context, tx transaction.RollbackCommitter, userUUID string) (models.Directory, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	dir, err := client.Directory.Query().
		Where(directory.And(
			directory.Name("root"),
			directory.HasOwnerWith(user.ID(userUUID)),
		)).
		WithOwner().
		WithChildDirs().
		WithChildFiles().
		First(ctx)
	if err != nil {
		return models.Directory{}, err
	}
	var fileMetadataList []models.FileMetadata
	if dir.Edges.ChildFiles != nil {
		for _, child := range dir.Edges.ChildFiles {
			fileMetadataList = append(fileMetadataList, models.FileMetadata{
				UUID:          child.ID,
				Filename:      child.Name,
				MIMEType:      child.MimeType,
				Path:          child.Path,
				RelPathOnDisk: child.RelPathOnDisk,
				Size:          int64(child.Size),
				CreatedAt:     child.CreatedAt,
				UpdatedAt:     child.UpdatedAt,
			})
		}
	}
	var dirMetadataList []models.DirectoryMetadata
	if dir.Edges.ChildDirs != nil {
		for _, child := range dir.Edges.ChildDirs {
			dirMetadataList = append(dirMetadataList, models.DirectoryMetadata{
				UUID:      child.ID,
				Dirname:   child.Name,
				Path:      child.Path,
				CreatedAt: child.CreatedAt,
				UpdatedAt: child.UpdatedAt,
			})
		}
	}
	return models.Directory{
		Metadata: models.DirectoryMetadata{
			UUID:      dir.ID,
			Dirname:   dir.Name,
			Path:      dir.Path,
			OwnerUUID: dir.Edges.Owner.ID,
			CreatedAt: dir.CreatedAt,
			UpdatedAt: dir.UpdatedAt,
		},
		Content: models.DirectoryContent{
			ListOfFiles: fileMetadataList,
			ListOfDirs:  dirMetadataList,
		},
	}, nil
}

func (e EntSQLDirectoryRepo) GetDirUUIDByPath(ctx context.Context, tx transaction.RollbackCommitter, path, userUUID string) (string, error) {
	panic("implement me")
}

func (e EntSQLDirectoryRepo) GetDirectChildDirUUIDList(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) ([]string, error) {
	panic("implement me")
}

func (e EntSQLDirectoryRepo) IsDirNameExist(ctx context.Context, tx transaction.RollbackCommitter, parentDirUUID, name string) (bool, error) {
	client := e.client
	if tx != nil {
		client = tx.(*ent.Tx).Client()
	}
	count, err := client.Directory.Query().
		Where(
			directory.And(
				directory.HasParentWith(directory.ID(parentDirUUID)),
				directory.Name(name),
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

func (e EntSQLDirectoryRepo) UpdateDirname(ctx context.Context, tx transaction.RollbackCommitter, newDirname, dirUUID string) error {
	panic("implement me")
}

func (e EntSQLDirectoryRepo) UpdateParentDirUUID(ctx context.Context, tx transaction.RollbackCommitter, newParentDirUUID, dirUUID string) error {
	panic("implement me")
}

func (e EntSQLDirectoryRepo) SoftRemoveDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) error {
	panic("implement me")
}

func (e EntSQLDirectoryRepo) HardRemoveDir(ctx context.Context, tx transaction.RollbackCommitter, dirUUID string) error {
	panic("implement me")
}
