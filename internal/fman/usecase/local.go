package usecase

import (
	"io"

	fileUtils "github.com/nvthongswansea/xtreme/internal/file-utils"
	"github.com/nvthongswansea/xtreme/internal/fman"
	"github.com/nvthongswansea/xtreme/internal/models"
	uuidUtils "github.com/nvthongswansea/xtreme/internal/uuid-utils"
)

type FManLocalUsecase struct {
	dbRepo    fman.FManDBRepo
	uuidGen   uuidUtils.UUIDGenerator
	fileSaver fileUtils.FileSaverRemover
}

// NewFManLocalUsecase create a new FManLocalUsecase
func NewFManLocalUsecase(dbRepo fman.FManDBRepo, uuidGen uuidUtils.UUIDGenerator, fileSaver fileUtils.FileSaverRemover) *FManLocalUsecase {
	return &FManLocalUsecase{
		dbRepo,
		uuidGen,
		fileSaver,
	}
}

func (u *FManLocalUsecase) UploadFile(newFile models.File, contentReader io.Reader) error {
	// generate a new UUID.
	newFile.UUID = u.uuidGen.NewUUID()
	// Save file to the disk.
	err := u.fileSaver.SaveFile(newFile.UUID, contentReader)
	if err != nil {
		return err
	}
	// Insert new file record to the DB.
	_, err = u.dbRepo.InsertFileRecord(newFile)
	return err
}

func (u *FManLocalUsecase) CopyFile() {
	panic("not implemented")
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
