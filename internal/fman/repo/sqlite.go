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
func (m *FManSQLiteRepo) InsertFileRecord(UUID, filename, parentUUID, realPath string, fileSize int64) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}
	result, err2 := db.Exec("INSERT INTO File(UUID, Filename, ParentUUID, RealPath, FileSize) VALUES (?,?,?,?,?)", UUID, filename, parentUUID, realPath, fileSize )
	if err2 != nil {
		return err2
	}
	_, err3 := result.RowsAffected()
	if err3 != nil {
		return err3
	}
	return nil

}

func (m *FManSQLiteRepo) ReadFileRecord(UUID string) (models.File, error) {

	db, err := config.GetDB()
	if err != nil {
	return models.File{}, err
	} else {
			rows, err2 := db.Query("SELECT * FROM File WHERE UUID=? ",UUID)
			if err2 != nil {
		return models.File{}, err2
		} else {
			var file models.File
			for rows.Next(){
			var file models.File
			rows.Scan(&file.UUID, &file.Filename, &file.Path, &file.RealPath, &file.FileSize, &file.ParentUUID, &file.CreatedAt, &file.UpdatedAt)
			}
		return file, nil
		}
	}

}

func (m *FManSQLiteRepo) UpdateFileRecord(filename, parentUUID string) error {

	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result, err2 := db.Exec("UPDATE File SET Filename=?, ParentUUID=?", filename, parentUUID )
	if err2 != nil {
		return err2
	}
	_, err3 := result.RowsAffected()
	if err3 != nil {
		return err3
	}
	
	return nil
}

func (m *FManSQLiteRepo) SoftRemoveFileRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) HardRemoveFileRecord(UUID string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result, err2 := db.Exec("DELETE FROM File WHERE UUID=? ",UUID)
	if err2 != nil {
		return err2
	}
	_, err3 := result.RowsAffected()
	if err3 != nil {
		return err3
	}
	
	return nil
}

func (m *FManSQLiteRepo) InsertDirRecord(UUID, dirname, parentUUID string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}
	result, err2 := db.Exec("INSERT INTO Directory(UUID, Dirname, ParentUUID) VALUES (?,?,?,?,?)", UUID, dirname, parentUUID)
	if err2 != nil {
		return err2
	}
	_, err3 := result.RowsAffected()
	if err3 != nil {
		return err3
	}
	return nil
}

func (m *FManSQLiteRepo) ReadDirRecord(UUID string) (models.Directory, error) {
	db, err := config.GetDB()
	if err != nil {
	return models.Directory{}, err
	} else {
			rows, err2 := db.Query("SELECT * FROM Directory WHERE UUID=? ",UUID)
			if err2 != nil {
		return models.Directory{}, err2
		} else {
			var dir models.Directory
			for rows.Next(){
			var dir models.Directory
			rows.Scan(&dir.UUID, &dir.Dirname, &dir.Path, &dir.ParentUUID, &dir.CreatedAt, &dir.UpdatedAt)
			}
		return dir, nil
		}
	}
}

func (m *FManSQLiteRepo) UpdateDirRecord(dirname, parentUUID string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result, err2 := db.Exec("UPDATE Directory SET Dirname=?, ParentUUID=?", dirname, parentUUID )
	if err2 != nil {
		return err2
	}
	_, err3 := result.RowsAffected()
	if err3 != nil {
		return err3
	}
	
	return nil
}

func (m *FManSQLiteRepo) SoftRemoveDirRecord(UUID string) error {
	return nil
}

func (m *FManSQLiteRepo) HardRemoveDirRecord(UUID string) error {
	db, err := config.GetDB()
	if err != nil {
		return err
	}

	result, err2 := db.Exec("DELETE FROM Directory WHERE UUID=? ",UUID)
	if err2 != nil {
		return err2
	}
	_, err3 := result.RowsAffected()
	if err3 != nil {
		return err3
	}
	
	return nil
}

func (m *FManSQLiteRepo) IsNameExist(filename, parentUUID string) (bool, error) {
	db, err := config.GetDB()
	if err != nil {
	return true, err
	}
	rows, err2 := db.Query("SELECT * FROM File WHERE UUID=? and ParentUUID=? ",filename,parentUUID)
	if err2 != nil {
		return true, err2
	}
	var files [] models.File
	for rows.Next(){
		var file models.File
		rows.Scan(&file.Filename, &file.ParentUUID)
		files = append(files, file)
	}
	if len(files) == 0 {
		return false, nil
	} else {
		return true, nil
	}
}


func (m *FManSQLiteRepo) IsParentUUIDExist(parentUUID string) (bool, error) {
	db, err := config.GetDB()
	if err != nil {
		return true, err
	}
	_, err2 := db.Query("select * from File where ParentUUID=? ",parentUUID)
	if err2 != nil {
		return false, err2
	}
	return true, nil
}
