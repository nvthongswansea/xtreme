package repo

import (
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/pkg/config"
)

type FManSQLiteRepo struct{}

// NewFManSQLiteRepo returns a new FManSQLiteRepo.
func NewFManSQLiteRepo() *FManSQLiteRepo {
	return &FManSQLiteRepo{}
}

// InsertFileRecord insert a new file record to SQLite DB.
func (m *FManSQLiteRepo) InsertFileRecord(newFile models.File) (bool, error) {

	db, err := config.GetDB()
	if err != nil {
		return false, err
	}

	result, err2 := db.Exec("insert into file(UUID, Filename, IsDir, Path, RealPath, ParentUUID, CreatedAt, UpdatedAt) values (?,?,?,?,?,?,?,?)", newFile.UUID, newFile.Filename, newFile.IsDir, newFile.Path, newFile.RealPath, newFile.ParentUUID, newFile.CreatedAt, newFile.UpdatedAt )
	if err2 != nil {
		return 	false, err2
	}
	rowsAffected, err3 := result.RowsAffected()
	if err3 != nil {
		return false, err3
	}
	return rowsAffected > 0 , nil
}

func (m *FManSQLiteRepo) ReadFileRecord(UUID string) (models.File, error) {

	db, err := config.GetDB()
	if err != nil {
	return models.File{}, err
	} else {
			rows, err2 := db.Query("select * from File where UUID=? ",UUID)
			if err2 != nil {
		return models.File{}, err2
		} else {
			var file models.File
			for rows.Next(){
			var file models.File
			rows.Scan(&file.UUID, &file.Filename, &file.IsDir, &file.Path, &file.RealPath, &file.ParentUUID, &file.CreatedAt, &file.UpdatedAt)
			}
		return file, nil
		}
	}

}

func (m *FManSQLiteRepo) UpdateFileRecord(file models.File) (bool, error) {

	db, err := config.GetDB()
	if err != nil {
		return false, err
	}

	result, err2 := db.Exec("update File set UUID=?, Filename=?, IsDir=?, Path=?, RealPath=?, ParentUUID=?, CreatedAt=?, UpdatedAt=?", file.UUID, file.Filename, file.IsDir, file.Path, file.RealPath, file.ParentUUID, file.CreatedAt, file.UpdatedAt )
	if err2 != nil {
		return 	false, err2
	}
	rowsAffected, err3 := result.RowsAffected()
	if err3 != nil {
		return false, err3
	}
	return rowsAffected > 0 , nil
}

func (m *FManSQLiteRepo) SoftRemoveFileRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) HardRemoveFileRecord(UUID string) error {
	return nil
}
