package usecase

import (
	"io"

	"github.com/nvthongswansea/xtreme/internal/fman"
	"github.com/nvthongswansea/xtreme/internal/models"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
)

// FManLocalUsecase provides usecase(logic) for file manager on local storage.
type FManLocalUsecase struct {
	dbRepo  fman.FManDBRepo
	uuidGen uuidUtils.UUIDGenerator
	fileOps fileUtils.FileSaveReadRemover
}

// NewFManLocalUsecase create a new FManLocalUsecase.
func NewFManLocalUsecase(dbRepo fman.FManDBRepo, uuidGen uuidUtils.UUIDGenerator, fileOps fileUtils.FileSaveReadRemover) *FManLocalUsecase {
	return &FManLocalUsecase{
		dbRepo,
		uuidGen,
		fileOps,
	}
}

func (u *FManLocalUsecase) UploadFile(newFile models.File, contentReader io.Reader) error {
	// Generate a new UUID.
	newFile.UUID = u.uuidGen.NewUUID()
	// Save file to the disk.
	err := u.fileOps.SaveFile(newFile.UUID, contentReader)
	if err != nil {
		return err
	}
	// Insert new file record to the DB.
	_, err = u.dbRepo.InsertFileRecord(newFile)
	return err
}

func (u *FManLocalUsecase) CopyFile(dstFile models.File, srcFile models.File) error {
	// Generate a new UUID for the destination file.
	dstFile.UUID = u.uuidGen.NewUUID()
	// Get source file pointer to read its content.
	srcFReadCloser, err := u.fileOps.ReadFile(srcFile.Filename)
	if err != nil {
		return err
	}
	defer srcFReadCloser.Close()
	// Save the dst file to the disk.
	err = u.fileOps.SaveFile(dstFile.UUID, srcFReadCloser)
	if err != nil {
		return err
	}
	// Insert a new file record to the DB.
	_, err = u.dbRepo.InsertFileRecord(dstFile)
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
