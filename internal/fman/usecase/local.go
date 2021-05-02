package usecase

import (
	"fmt"
	"io"

	"github.com/nvthongswansea/xtreme/internal/fman"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
)

// FManLocalUsecase provides usecase(logic) for file manager on local storage.
type FManLocalUsecase struct {
	dbFileRepo fman.FManFileDBRepo
	dbDirRepo  fman.FManDirDBRepo
	dbValRepo  fman.FManValidator
	uuidGen    uuidUtils.UUIDGenerator
	fileOps    fileUtils.FileSaveReadRemover
}

// NewFManLocalUsecase create a new FManLocalUsecase.
func NewFManLocalUsecase(dbFileRepo fman.FManFileDBRepo, dbDirRepo fman.FManDirDBRepo, dbValRepo fman.FManValidator,
	uuidGen uuidUtils.UUIDGenerator, fileOps fileUtils.FileSaveReadRemover) *FManLocalUsecase {
	return &FManLocalUsecase{
		dbFileRepo,
		dbDirRepo,
		dbValRepo,
		uuidGen,
		fileOps,
	}
}

func (u *FManLocalUsecase) UploadFile(filename, parentUUID string, contentReader io.Reader) error {
	// Validate parent UUID.
	parentUUIDok, err := u.dbValRepo.IsParentUUIDExist(parentUUID)
	if err != nil {
		return err
	}
	if !parentUUIDok {
		return fmt.Errorf("parent UUID (%s) does not exist", parentUUID)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := u.dbValRepo.IsNameExist(filename, parentUUID)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("%s already exists in the desired location", filename)
	}
	// Generate a new UUID.
	newFileUUID := u.uuidGen.NewUUID()
	// Save file to the disk.
	size, realPath, err := u.fileOps.SaveFile(newFileUUID, contentReader)
	if err != nil {
		return err
	}
	// Insert new file record to the DB.
	if err := u.dbFileRepo.InsertFileRecord(newFileUUID, filename, parentUUID, realPath, size); err != nil {
		// If error reprents while inserting a new record,
		// remove the file from the storage.
		return u.fileOps.RemoveFile(newFileUUID)
	}
	return err
}

func (u *FManLocalUsecase) CopyFile(srcUUID, dstParentUUID string) error {
	// Validate parent UUID.
	parentUUIDok, err := u.dbValRepo.IsParentUUIDExist(dstParentUUID)
	if err != nil {
		return err
	}
	if !parentUUIDok {
		return fmt.Errorf("parent UUID (%s) does not exist", dstParentUUID)
	}
	// Get the source filename.
	srcFile, err := u.dbFileRepo.ReadFileRecord(srcUUID)
	if err != nil {
		return err
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := u.dbValRepo.IsNameExist(srcFile.Filename, dstParentUUID)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("%s already exists in the desired location", srcFile.Filename)
	}
	// Generate a new UUID for the destination file.
	newFileUUID := u.uuidGen.NewUUID()
	// Get source file pointer to read its content.
	srcFReadCloser, err := u.fileOps.ReadFile(srcUUID)
	if err != nil {
		return err
	}
	defer srcFReadCloser.Close()
	// Save the dst file to the disk.
	size, realPath, err := u.fileOps.SaveFile(newFileUUID, srcFReadCloser)
	if err != nil {
		return err
	}
	// Insert a new file record to the DB.
	if err = u.dbFileRepo.InsertFileRecord(newFileUUID, srcFile.Filename, dstParentUUID, realPath, size); err != nil {
		// If error reprents while inserting a new record,
		// remove the file from the storage.
		return u.fileOps.RemoveFile(newFileUUID)
	}
	return err
}

func (u *FManLocalUsecase) CreateNewDirectory(dirname, parentUUID string) error {
	// Validate parent UUID.
	parentUUIDok, err := u.dbValRepo.IsParentUUIDExist(parentUUID)
	if err != nil {
		return err
	}
	if !parentUUIDok {
		return fmt.Errorf("parent UUID (%s) does not exist", parentUUID)
	}
	// Check if the directory already exists in a desired location in the db.
	isExist, err := u.dbValRepo.IsNameExist(dirname, parentUUID)
	if err != nil {
		return err
	}
	if isExist {
		return fmt.Errorf("%s already exists in the desired location", dirname)
	}
	// Generate a new UUID.
	newDirUUID := u.uuidGen.NewUUID()
	// Insert new file record to the DB.
	if err := u.dbDirRepo.InsertDirRecord(newDirUUID, dirname, parentUUID); err != nil {
		// If error reprents while inserting a new record,
		// remove the file from the storage.
		return u.fileOps.RemoveFile(newDirUUID)
	}
	return err
}

func (u *FManLocalUsecase) MoveFile() {
	panic("not implemented")
}

func (u *FManLocalUsecase) RemoveFile() {
	panic("not implemented")
}

func (u *FManLocalUsecase) MoveFileToRecyleBin() {
	panic("not implemented")
}
