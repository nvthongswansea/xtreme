package usecase

import (
	"fmt"
	"io"

	"github.com/nvthongswansea/xtreme/internal/fman"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
	log "github.com/sirupsen/logrus"
)

// FManLocalUsecase provides usecase(logic) for file manager on local storage.
type FManLocalUsecase struct {
	dbFileRepo fman.FManFileDBRepo
	dbDirRepo  fman.FManDirDBRepo
	dbValRepo  fman.FManValidateDBRepo
	uuidGen    uuidUtils.UUIDGenerator
	fileOps    fileUtils.FileSaveReadRemover
}

// NewFManLocalUsecase create a new FManLocalUsecase.
func NewFManLocalUsecase(dbFileRepo fman.FManFileDBRepo, dbDirRepo fman.FManDirDBRepo, dbValRepo fman.FManValidateDBRepo,
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
	// Generate a new UUID.
	newFileUUID := u.uuidGen.NewUUID()
	logger := log.WithFields(log.Fields{
		"Layer":      "usecase-local",
		"Operation":  "UploadFile",
		"filename":   filename,
		"fileUUID":   newFileUUID,
		"parentUUID": parentUUID,
	})
	logger.Debug("Start uploading file")
	defer logger.Debug("Finish uploading file")
	// Validate parent UUID.
	parentUUIDok, err := u.dbValRepo.IsParentUUIDExist(parentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentUUIDExist failed with error %s", err.Error())
		return err
	}
	if !parentUUIDok {
		logger.Infof("[-USER-] parent UUID (%s) does not exist", parentUUID)
		return fmt.Errorf("parent UUID (%s) does not exist", parentUUID)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := u.dbValRepo.IsNameExist(filename, parentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", filename)
		return fmt.Errorf("%s already exists in the desired location", filename)
	}
	// Save file to the local disk.
	size, realPath, err := u.fileOps.SaveFile(newFileUUID, contentReader)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveFile failed with error %s", err.Error())
		return err
	}
	// Insert new file record to the DB.
	if err := u.dbFileRepo.InsertFileRecord(newFileUUID, filename, parentUUID, realPath, size); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s", newFileUUID)
			err := u.fileOps.RemoveFile(newFileUUID)
			if err != nil {
				logger.Errorf("[-INTERNAL-] RemoveFile failed with error %s", err.Error())
			}
		}()
		logger.Errorf("[-INTERNAL-] InsertFileRecord failed with error %s", err.Error())
		return err
	}
	return nil
}

func (u *FManLocalUsecase) CopyFile(srcUUID, dstParentUUID string) error {
	// Generate a new UUID for the destination file.
	newFileUUID := u.uuidGen.NewUUID()
	logger := log.WithFields(log.Fields{
		"Layer":          "usecase-local",
		"Operation":      "CopyFile",
		"sourceFileUUID": srcUUID,
		"fileUUID":       newFileUUID,
		"dstParentUUID":  dstParentUUID,
	})
	logger.Debug("Start copying file")
	defer logger.Debug("Finish copying file")
	// Validate parent UUID.
	parentUUIDok, err := u.dbValRepo.IsParentUUIDExist(dstParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentUUIDExist failed with error %s", err.Error())
		return err
	}
	if !parentUUIDok {
		logger.Infof("[-USER-] parent UUID (%s) does not exist", dstParentUUID)
	}
	// Get the source filename.
	srcFile, err := u.dbFileRepo.ReadFileRecord(srcUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFileRecord failed with error %s", err.Error())
		return err
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := u.dbValRepo.IsNameExist(srcFile.Filename, dstParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return fmt.Errorf("%s already exists in the desired location", srcFile.Filename)
	}
	// Get source file pointer to read its content.
	srcFReadCloser, err := u.fileOps.ReadFile(srcUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		return err
	}
	defer srcFReadCloser.Close()
	// Save the dst file to the disk.
	size, realPath, err := u.fileOps.SaveFile(newFileUUID, srcFReadCloser)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveFile failed with error %s", err.Error())
		return err
	}
	// Insert a new file record to the DB.
	if err = u.dbFileRepo.InsertFileRecord(newFileUUID, srcFile.Filename, dstParentUUID, realPath, size); err != nil {
		// If error reprents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s", newFileUUID)
			err := u.fileOps.RemoveFile(newFileUUID)
			if err != nil {
				logger.Errorf("[-INTERNAL-] RemoveFile failed with error %s", err.Error())
			}
		}()
		logger.Errorf("[-INTERNAL-] InsertFileRecord failed with error %s", err.Error())
		return err
	}
	return nil
}

func (u *FManLocalUsecase) CreateNewDirectory(dirname, parentUUID string) error {
	// Generate a new UUID.
	newDirUUID := u.uuidGen.NewUUID()
	logger := log.WithFields(log.Fields{
		"Layer":      "usecase-local",
		"Operation":  "CreateNewDirectory",
		"dirname":    dirname,
		"parentUUID": parentUUID,
	})
	logger.Debug("Start creating a new directory")
	defer logger.Debug("Finish creating a new directory")
	// Validate parent UUID.
	parentUUIDok, err := u.dbValRepo.IsParentUUIDExist(parentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentUUIDExist failed with error %s", err.Error())
		return err
	}
	if !parentUUIDok {
		logger.Infof("[-USER-] parent UUID (%s) does not exist", parentUUID)
		return fmt.Errorf("parent UUID (%s) does not exist", parentUUID)
	}
	// Check if the directory already exists in a desired location in the db.
	isExist, err := u.dbValRepo.IsNameExist(dirname, parentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirname)
		return fmt.Errorf("%s already exists in the desired location", dirname)
	}

	// Insert new file record to the DB.
	err = u.dbDirRepo.InsertDirRecord(newDirUUID, dirname, parentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertDirRecord failed with error %s", err.Error())
	}
	return nil
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
