package local

import (
	"bytes"
	"context"
	"github.com/nvthongswansea/xtreme/internal/author"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/directory"
	"github.com/nvthongswansea/xtreme/internal/repository/file"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/fileUtils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuidUtils"
	log "github.com/sirupsen/logrus"
	"io"
	"path/filepath"
)

const (
	invalidDirNameErrorMessage       = "directory name is invalid"
	invalidFileNameErrorMessage      = "filename is invalid"
	invalidUserUUIDErrorMessage      = "user UUID is not valid"
	invalidParentDirUUIDErrorMessage = "parent directory UUID is not valid"
	invalidFileUUIDErrorMessage      = "file UUID is not valid"
	invalidDirUUIDErrorMessage       = "directory UUID is not valid"
	invalidPathErrorMessage          = "path is not valid"

	pathNotFoundMessage            = "path not found"
	nameAlreadyExistErrorMessage   = "name already exists in desired location"
	forbiddenOperationErrorMessage = "forbidden operation"
)

// MultiOSFileManager implements interface local.FileManagerService.
// This implementation of local.FileManagerService supports multiple Operating Systems.
type MultiOSFileManager struct {
	fileRepo     file.Repository
	dirRepo      directory.Repository
	txRepo       transaction.TxRepository
	uuidTool     uuidUtils.UUIDGenerateValidator
	fileOps      fileUtils.FileSaveReadCpRmer
	fileCompress fileUtils.FileCompressor
	author       author.Authorizer
}

// NewMultiOSFileManager creates a new MultiOSFileManager.
func NewMultiOSFileManager(fileRepo file.Repository, dirRepo directory.Repository, txRepo transaction.TxRepository,
	uuidTool uuidUtils.UUIDGenerateValidator, fileOps fileUtils.FileSaveReadCpRmer, fileCompress fileUtils.FileCompressor,
	author author.Authorizer) *MultiOSFileManager {
	return &MultiOSFileManager{
		fileRepo:     fileRepo,
		dirRepo:      dirRepo,
		txRepo:       txRepo,
		uuidTool:     uuidTool,
		fileOps:      fileOps,
		fileCompress: fileCompress,
		author:       author,
	}
}

// CreateNewFile creates a new empty file with a given filename in a specific user storage space.
func (m *MultiOSFileManager) CreateNewFile(ctx context.Context, userUUID, filename, parentDirUUID string) (string, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":           "local-service_handler-multi_os",
		"Operation":     "CreateNewFile",
		"userUUID":      userUUID,
		"filename":      filename,
		"parentDirUUID": parentDirUUID,
	})
	logger.Debug("Start creating a new file")
	defer logger.Debug("Finish creating a new file")
	emptyBytes := make([]byte, 0) // Create an empty byte slice
	return m.UploadFile(ctx, userUUID, filename, parentDirUUID, bytes.NewReader(emptyBytes))
}

func (m *MultiOSFileManager) UploadFile(ctx context.Context, userUUID, filename, parentDirUUID string, contentReader io.Reader) (string, error) {
	logger := log.WithFields(log.Fields{
		"Loc":           "local-service_handler-multi_os",
		"Operation":     "UploadFile",
		"userUUID":      userUUID,
		"filename":      filename,
		"parentDirUUID": parentDirUUID,
	})
	logger.Debug("Start uploading file")
	defer logger.Debug("Finish uploading file")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return "", err
	}
	if !fileUtils.IsFilenameOk(filename) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
		return "", err
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
		return "", err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, parentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return "", err
	}

	// Check if the file already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, parentDirUUID, filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, parentDirUUID, filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", filename)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return "", err
	}
	// Save file to the local storage.
	saveFileFn := func(filepath string) (int64, error) {
		size, err := m.fileOps.SaveFile(filepath, contentReader)
		if err != nil {
			logger.Errorf("[-INTERNAL-] SaveFile failed with error %s", err.Error())
			err = models.XtremeError{
				Code:    models.InternalServerErrorCode,
				Message: err.Error(),
			}
			return 0, err
		}
		return size, nil
	}

	// Get path of the parent dir UUID
	parentDirMeta, err := m.dirRepo.GetDirMetadata(ctx, tx, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	newFile := models.File{
		Metadata: models.FileMetadata{
			Filename:   filename,
			ParentUUID: parentDirUUID,
			OwnerUUID:  userUUID,
			Path:       filepath.Join(parentDirMeta.Path, filename),
		},
	}
	fUUID, err := m.fileRepo.InsertFile(ctx, tx, newFile, saveFileFn)
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	logger.WithField("newFileUUID", fUUID)

	return fUUID, nil
}

func (m *MultiOSFileManager) CopyFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error) {
	logger := log.WithFields(log.Fields{
		"Loc":              "local-service_handler-multi_os",
		"Operation":        "CopyFile",
		"userUUID":         userUUID,
		"fileUUID":         fileUUID,
		"dstParentDirUUID": dstParentDirUUID,
	})
	logger.Debug("Start copying file")
	defer logger.Debug("Finish copying file")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return "", err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return "", err
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
		return "", err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, dstParentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return "", err
	}
	isAuthorized, err = m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.CopyFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return "", err
	}
	return m.copyFile(ctx, tx, logger, userUUID, fileUUID, dstParentDirUUID)
}

func (m *MultiOSFileManager) copyFile(ctx context.Context, tx transaction.RollbackCommitter, logger *log.Entry, userUUID, fileUUID, dstParentDirUUID string) (string, error) {
	// Get the source file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, tx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	// Check if the file already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, dstParentDirUUID, srcFile.Filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, dstParentDirUUID, srcFile.Filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return "", err
	}
	// Get source file reader to read its content.
	srcFReadCloser, err := m.fileOps.ReadFile(srcFile.RelPathOnDisk)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	defer srcFReadCloser.Close()

	saveFileFn := func(filepath string) (int64, error) {
		// Save the dst file to the disk.
		size, err := m.fileOps.SaveFile(filepath, srcFReadCloser)
		if err != nil {
			logger.Errorf("[-INTERNAL-] SaveFile failed with error %s", err.Error())
			err = models.XtremeError{
				Code:    models.InternalServerErrorCode,
				Message: err.Error(),
			}
			return 0, err
		}
		return size, nil
	}

	// Get path of the parent dir UUID
	parentDirMeta, err := m.dirRepo.GetDirMetadata(ctx, tx, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}

	newFile := models.File{
		Metadata: models.FileMetadata{
			Filename:   srcFile.Filename,
			ParentUUID: dstParentDirUUID,
			OwnerUUID:  userUUID, // other user can copy the owner's file (if permitted)
			Path:       filepath.Join(parentDirMeta.Path, srcFile.Filename),
		},
	}
	newFileUUID, err := m.fileRepo.InsertFile(ctx, tx, newFile, saveFileFn)
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}

	return newFileUUID, nil
}

func (m *MultiOSFileManager) RenameFile(ctx context.Context, userUUID, fileUUID, newFilename string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":         "local-service_handler-multi_os",
		"Operation":   "RenameFile",
		"userUUID":    userUUID,
		"fileUUID":    fileUUID,
		"newFilename": newFilename,
	})
	logger.Debug("Start renaming file")
	defer logger.Debug("Finish renaming file")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return err
	}
	if !fileUtils.IsFilenameOk(newFilename) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
		return err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.UpdateFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return err
	}

	// Get the source file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, tx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	// Check if the file already exists in a current location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, srcFile.ParentUUID, newFilename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, srcFile.ParentUUID, newFilename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", newFilename)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return err
	}
	err = m.fileRepo.UpdateFilename(ctx, tx, newFilename, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateFilename failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	return nil
}

func (m *MultiOSFileManager) SoftRemoveFile(ctx context.Context, userUUID, fileUUID string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "SoftRemoveFile",
		"userUUID":  userUUID,
		"fileUUID":  fileUUID,
	})
	logger.Debug("Start removing file (SOFT)")
	defer logger.Debug("Finish removing file (SOFT)")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return err
	}

	err = m.fileRepo.SoftRemoveFile(ctx, tx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SoftRemoveFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	return nil
}

func (m *MultiOSFileManager) HardRemoveFile(ctx context.Context, userUUID, fileUUID string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "HardRemoveFile",
		"userUUID":  userUUID,
		"fileUUID":  fileUUID,
	})
	logger.Debug("Start removing file (HARD)")
	defer logger.Debug("Finish removing file (HARD)")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return err
	}

	err = m.fileRepo.HardRemoveFile(ctx, tx, fileUUID, m.fileOps.RemoveFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] HardRemoveFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	return nil
}

func (m *MultiOSFileManager) MoveFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) error {
	logger := log.WithFields(log.Fields{
		"Loc":              "local-service_handler-multi_os",
		"Operation":        "MoveFile",
		"userUUID":         userUUID,
		"fileUUID":         fileUUID,
		"dstParentDirUUID": dstParentDirUUID,
	})
	logger.Debug("Start moving file")
	defer logger.Debug("Finish moving file")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return err
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
		return err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, dstParentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return err
	}
	isAuthorized, err = m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.CopyFileAction, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return err
	}

	// Get the file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, tx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	// Check if the file already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, dstParentDirUUID, srcFile.Filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, dstParentDirUUID, srcFile.Filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return err
	}
	err = m.fileRepo.UpdateParentDirUUID(ctx, tx, dstParentDirUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateParentDirUUID failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	return nil
}

func (m *MultiOSFileManager) GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error) {
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetFile",
		"userUUID":  userUUID,
		"fileUUID":  fileUUID,
	})
	logger.Debug("Start retrieving file metadata")
	defer logger.Debug("Finish retrieving file metadata")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.File{}, err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return models.File{}, err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return models.File{}, err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.File{}, err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return models.File{}, err
	}

	// Get the file metadata.
	srcFile, err := m.fileRepo.GetFile(ctx, tx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.File{}, err
	}
	return srcFile, nil
}

func (m *MultiOSFileManager) DownloadFile(ctx context.Context, userUUID, fileUUID string) (models.FilePayload, error) {
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "DownloadFile",
		"userUUID":  userUUID,
		"fileUUID":  fileUUID,
	})
	logger.Debug("Start getting file payload")
	defer logger.Debug("Finish getting file payload")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.FilePayload{}, err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return models.FilePayload{}, err
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		return models.FilePayload{}, err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, fileUUID, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.FilePayload{}, err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return models.FilePayload{}, err
	}

	// Get the file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, tx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.FilePayload{}, err
	}
	fileRCloser, err := m.fileOps.ReadFile(srcFile.RelPathOnDisk)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.FilePayload{}, err
	}
	payload := models.FilePayload{
		Filename: srcFile.Filename,
		File:     fileRCloser,
	}
	return payload, nil
}

func (m *MultiOSFileManager) DownloadFileBatch(ctx context.Context, userUUID string, fileUUIDs []string) (models.TmpFilePayload, error) {
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "DownloadFileBatch",
		"userUUID":  userUUID,
		"fileUUIDs": fileUUIDs,
	})
	logger.Debug("Start getting files' payloads")
	defer logger.Debug("Finish getting files' payloads")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.TmpFilePayload{}, err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return models.TmpFilePayload{}, err
	}
	for _, uuid := range fileUUIDs {
		if !m.uuidTool.ValidateUUID(uuid) {
			logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
			err = models.XtremeError{
				Code:    models.BadInputErrorCode,
				Message: invalidFileUUIDErrorMessage,
			}
			return models.TmpFilePayload{}, err
		}
	}
	// Authorization.
	for _, uuid := range fileUUIDs {
		isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, tx, userUUID, uuid, author.ViewFileAction)
		if err != nil {
			logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
			err = models.XtremeError{
				Code:    models.InternalServerErrorCode,
				Message: err.Error(),
			}
			return models.TmpFilePayload{}, err
		}
		if !isAuthorized {
			logger.Info(forbiddenOperationErrorMessage)
			err = models.XtremeError{
				Code:    models.ForbiddenOperationErrorCode,
				Message: forbiddenOperationErrorMessage,
			}
			return models.TmpFilePayload{}, err
		}
	}
	// Get the files' metadata.
	srcFiles, err := m.fileRepo.GetFileMetadataBatch(ctx, tx, fileUUIDs)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataBatch failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.TmpFilePayload{}, err
	}
	inZipPaths := make(map[string]string)
	for _, srcFile := range srcFiles {
		inZipPaths[srcFile.Path] = srcFile.RelPathOnDisk
	}
	tmpFilePath, err := m.fileCompress.CompressFiles(inZipPaths)
	if err != nil {
		logger.Errorf("[-INTERNAL-] CompressFiles failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.TmpFilePayload{}, err
	}
	tmpFile, err := m.fileOps.GetFileReadCloserRmer(tmpFilePath)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileReadCloserRmer failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.TmpFilePayload{}, err
	}
	payload := models.TmpFilePayload{
		Filename: filepath.Base(tmpFilePath),
		TmpFile:  tmpFile,
	}
	return payload, nil
}

func (m *MultiOSFileManager) SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.FileMetadata, []models.DirectoryMetadata, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":           "local-service_handler-multi_os",
		"Operation":     "SearchByName",
		"userUUID":      userUUID,
		"parentDirUUID": parentDirUUID,
	})
	logger.Debug("Start retrieving UserUUID via path")
	defer logger.Debug("Finish retrieving UserUUID via path")
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return nil, nil, err
	}
	defer func() {
		err = m.txRepo.FinishTransaction(tx, err)
		if err != nil {
			logger.Errorf("[-INTERNAL-] FinishTransaction failed with error %s", err.Error())
		}
	}()
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return nil, nil, err
	}
	if !fileUtils.IsFilenameOk(filename) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
		return nil, nil, err
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
		return nil, nil, err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, parentDirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return nil, nil, err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return nil, nil, err
	}

	fileMetadataList, err := m.fileRepo.GetFileMetadataListByName(ctx, tx, parentDirUUID, filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataListByName failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return nil, nil, err
	}
	dirMetadataList, err := m.dirRepo.GetDirMetadataListByName(ctx, tx, parentDirUUID, filename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadataListByName failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return nil, nil, err
	}
	return fileMetadataList, dirMetadataList, nil
}
