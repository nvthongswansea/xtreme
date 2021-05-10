package database

import "github.com/nvthongswansea/xtreme/internal/models"

type FManSQLiteRepo struct{}

// NewFManSQLiteRepo returns a new FManSQLiteRepo.
func NewFManSQLiteRepo() *FManSQLiteRepo {
	return &FManSQLiteRepo{}
}

// InsertFileRecord insert a new file record to SQLite DB.
func (m *FManSQLiteRepo) InsertFileRecord(UUID, filename, parentUUID, realPath string, fileSize int64) error {
	return nil
}

func (m *FManSQLiteRepo) ReadFileRecord(UUID string) (models.File, error) {
	return models.File{}, nil
}

func (m *FManSQLiteRepo) UpdateFileRecord(filename, parentUUID string) error {
	return nil
}

func (m *FManSQLiteRepo) SoftRemoveFileRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) HardRemoveFileRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) InsertDirRecord(UUID, dirname, parentUUID string) error {
	return nil
}

func (m *FManSQLiteRepo) ReadDirRecord(UUID string) (models.Directory, error) {
	return models.Directory{}, nil
}

func (m *FManSQLiteRepo) UpdateDirRecord(filename, parentUUID string) error {
	return nil
}

func (m *FManSQLiteRepo) SoftRemoveDirRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) HardRemoveDirRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) IsNameExist(filename, parentUUID string) (bool, error) {
	return false, nil
}

func (m *FManSQLiteRepo) IsParentUUIDExist(parentUUID string) (bool, error) {
	return true, nil
}
