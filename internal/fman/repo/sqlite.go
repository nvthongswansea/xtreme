package repo

import "github.com/nvthongswansea/xtreme/internal/models"

type FManSQLiteRepo struct{}

// NewFManSQLiteRepo returns a new FManSQLiteRepo.
func NewFManSQLiteRepo() *FManSQLiteRepo {
	return &FManSQLiteRepo{}
}

// InsertFileRecord insert a new file record to SQLite DB.
func (m *FManSQLiteRepo) InsertFileRecord(newFile models.File) (models.File, error) {
	return models.File{}, nil
}

func (m *FManSQLiteRepo) ReadFileRecord(UUID string) (models.File, error) {
	return models.File{}, nil
}

func (m *FManSQLiteRepo) UpdateFileRecord(file models.File) error {
	return nil
}

func (m *FManSQLiteRepo) SoftRemoveFileRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) HardRemoveFileRecord(UUID string) error {
	return nil
}
