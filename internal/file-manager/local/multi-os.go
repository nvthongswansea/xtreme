package local

import (
	"bytes"
	"context"
	"github.com/nvthongswansea/xtreme/internal/author"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/directory"
	"github.com/nvthongswansea/xtreme/internal/repository/file"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
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
	invalidDirUUIDErrorMessage       = "file UUID is not valid"
	invalidPathErrorMessage          = "path is not valid"

	pathNotFoundMessage            = "path not found"
	nameAlreadyExistErrorMessage   = "name already exists in desired location"
	forbiddenOperationErrorMessage = "forbidden operation"
)

// MultiOSFileManager implements interface local.FileManagerService.
// This implementation of local.FileManagerService supports multiple Operating Systems.
type MultiOSFileManager struct {
	fileRepo file.Repository
	dirRepo directory.Repository
	uuidTool        uuidUtils.UUIDGenerateValidator
	fileOps         fileUtils.FileSaveReadCpRmer
	fileCompress    fileUtils.FileCompressor
	author          author.Authorizer
}

// NewMultiOSFileManager creates a new MultiOSFileManager.
func NewMultiOSFileManager(fileRepo file.Repository, dirRepo directory.Repository, uuidTool uuidUtils.UUIDGenerateValidator,
	fileOps fileUtils.FileSaveReadCpRmer, fileCompress fileUtils.FileCompressor, author author.Authorizer) *MultiOSFileManager {
	return &MultiOSFileManager{
		fileRepo,
		dirRepo,
		uuidTool,
		fileOps,
		fileCompress,
		author,
	}
}

func (m *MultiOSFileManager) GetRootDirectory(ctx context.Context, userUUID string) (models.Directory, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetRootDirectory",
		"userUUID":  userUUID,
	})
	logger.Debug("Start retrieving root directory")
	defer logger.Debug("Finish retrieving root directory")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.Directory{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	rooDir, err := m.dirRepo.GetRootDirectoryByUserUUID(ctx, userUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetRootDirectoryByUserUUID failed with error %s", err.Error())
		return models.Directory{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return rooDir, nil
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	if !fileUtils.IsFilenameOk(newFilename) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.UpdateFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get the source file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	// Check if the file already exists in a current location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, newFilename, srcFile.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, newFilename, srcFile.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", newFilename)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	err = m.fileRepo.UpdateFilename(ctx, fileUUID, newFilename)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateFilename failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	err = m.fileRepo.SoftRemoveFile(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SoftRemoveFile failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
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
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	err = m.fileRepo.HardRemoveFile(ctx, fileUUID, m.fileOps.RemoveFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] HardRemoveFile failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return nil
}

func (m *MultiOSFileManager) GetDirectory(ctx context.Context, userUUID, dirUUID string) (models.Directory, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetDirectory",
		"userUUID":  userUUID,
		"dirUUID":   dirUUID,
	})
	logger.Debug("Start getting directory")
	defer logger.Debug("Finish getting directory")
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.Directory{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return models.Directory{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.Directory{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.Directory{}, models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	directory, err := m.dirRepo.GetDirectory(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirectory failed with error %s", err.Error())
		return models.Directory{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return directory, nil
}

func (m *MultiOSFileManager) CopyDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	logger := log.WithFields(log.Fields{
		"Loc":              "local-service_handler-multi_os",
		"Operation":        "CopyDirectory",
		"userUUID":         userUUID,
		"dirUUID":          dirUUID,
		"dstParentDirUUID": dstParentDirUUID,
	})
	logger.Debug("Start moving directory")
	defer logger.Debug("Finish moving directory")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.CopyDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return "", models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}
	return m.copyDirectory(ctx, logger, userUUID, dirUUID, dstParentDirUUID)
}

func (m *MultiOSFileManager) copyDirectory(ctx context.Context, logger *log.Entry, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	// Get to-be-copied dir metadata.
	copiedDirMeta, err := m.dirRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	logger.WithField("currentCopyPath", copiedDirMeta.Path)
	// Copy the current dir to the new location
	nDirCopyUUID, err := m.createNewDirectory(ctx, logger, userUUID, copiedDirMeta.Dirname, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] createNewDirectory failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}

	// Copy files to the newly created directory.
	childFileMetaList, err := m.fileRepo.GetFileMetadataListByDir(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataListByDir failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	for _, childMeta := range childFileMetaList {
		_, err := m.copyFile(ctx, logger, userUUID, childMeta.UUID, nDirCopyUUID)
		if err != nil {
			logger.Errorf("[-INTERNAL-] copyFile failed with error %s", err.Error())
			return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
		}
	}

	// Get direct child-directories' UUIDs of the current to-be-copied dir.
	dChildDirUUIDList, err := m.dirRepo.GetDirectChildDirUUIDList(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirectChildDirUUIDList failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	for _, childDirUUID := range dChildDirUUIDList {
		// Recursively copy the child directories with their child files/directories.
		_, err := m.copyDirectory(ctx, logger, userUUID, childDirUUID, nDirCopyUUID)
		if err != nil {
			logger.Errorf("[-INTERNAL-] copyDirectory failed with error %s", err.Error())
			return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
		}
	}
	return nDirCopyUUID, nil
}

func (m *MultiOSFileManager) RenameDirectory(ctx context.Context, userUUID, dirUUID, newDirname string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":        "local-service_handler-multi_os",
		"Operation":  "RenameDirectory",
		"userUUID":   userUUID,
		"dirUUID":    dirUUID,
		"newDirname": newDirname,
	})
	logger.Debug("Start renaming directory")
	defer logger.Debug("Finish renaming directory")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	if !fileUtils.IsFilenameOk(newDirname) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.UpdateDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get the source directory metadata.
	dirMeta, err := m.dirRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	// Check if the name already exists in a current location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, newDirname, dirMeta.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, newDirname, dirMeta.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", newDirname)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	err = m.dirRepo.UpdateDirname(ctx, dirUUID, newDirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateDirname failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return nil
}

func (m *MultiOSFileManager) MoveDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	logger := log.WithFields(log.Fields{
		"Loc":              "local-service_handler-multi_os",
		"Operation":        "MoveDirectory",
		"userUUID":         userUUID,
		"dirUUID":          dirUUID,
		"dstParentDirUUID": dstParentDirUUID,
	})
	logger.Debug("Start moving directory")
	defer logger.Debug("Finish moving directory")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.CopyDirAction, author.RemoveDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return "", models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get the directory metadata.
	dirMeta, err := m.dirRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	// Check if the directory name already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, dirMeta.Dirname, dirMeta.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return "",models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, dirMeta.Dirname, dirMeta.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirMeta.Dirname)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	err = m.dirRepo.UpdateParentDirUUID(ctx, dirUUID, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateParentDirUUID failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return dirUUID, nil
}

func (m *MultiOSFileManager) DownloadDirectory(ctx context.Context, userUUID, dirUUID string) (models.TmpFilePayload, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "DownloadDirectory",
		"userUUID":  userUUID,
		"dirUUID":   dirUUID,
	})
	logger.Debug("Start getting directory")
	defer logger.Debug("Finish getting directory")
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.TmpFilePayload{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return models.TmpFilePayload{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.TmpFilePayload{}, models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get dir metadata.
	dirMeta, err := m.dirRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}

	// Get all child-files metadata.
	fileMetadataList, err := m.fileRepo.GetFileMetadataListByDir(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataListByDir failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}

	inZipPaths := make(map[string]string)
	for _, childFileMeta := range fileMetadataList {
		relPath, err := filepath.Rel(dirMeta.Path, childFileMeta.Path)
		if err != nil {
			logger.Errorf("[-INTERNAL-] filepath.Rel failed with error %s", err.Error())
			return models.TmpFilePayload{}, models.XtremeError{
				Code: models.InternalServerErrorCode,
				Message: err.Error(),
			}
		}
		inZipPaths[relPath] = childFileMeta.RelPathOnDisk
	}
	tmpFilePath, err := m.fileCompress.CompressFiles(inZipPaths)
	if err != nil {
		logger.Errorf("[-INTERNAL-] CompressFiles failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	tmpFHanlder, err := m.fileOps.GetFileReadCloserRmer(tmpFilePath)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileReadCloserRmer failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	payload := models.TmpFilePayload{
		Filename: filepath.Base(tmpFilePath),
		TmpFile:  tmpFHanlder,
	}
	return payload, nil
}

func (m *MultiOSFileManager) SoftRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "SoftRemoveDir",
		"userUUID":  userUUID,
		"dirUUID":   dirUUID,
	})
	logger.Debug("Start removing directory (SOFT)")
	defer logger.Debug("Finish removing directory (SOFT)")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.RemoveDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	err = m.dirRepo.SoftRemoveDir(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SoftRemoveDir failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return nil
}

func (m *MultiOSFileManager) HardRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "HardRemoveDir",
		"userUUID":  userUUID,
		"dirUUID":   dirUUID,
	})
	logger.Debug("Start removing directory (HARD)")
	defer logger.Debug("Finish removing directory (HARD)")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.RemoveDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	err = m.dirRepo.HardRemoveDir(ctx, dirUUID, m.fileOps.RemoveFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] HardRemoveDir failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return nil
}

func (m *MultiOSFileManager) GetDirUUIDByPath(ctx context.Context, userUUID, path string) (string, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetDirUUIDByPath",
		"userUUID":  userUUID,
		"path":      path,
	})
	logger.Debug("Start retrieving UUID via path")
	defer logger.Debug("Finish retrieving UUID via path")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	path = filepath.Clean(path)
	if path == "" {
		logger.Info("[-USER-]", invalidPathErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidPathErrorMessage,
		}
	}
	uuid, err := m.dirRepo.GetDirUUIDByPath(ctx, userUUID, path)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirUUIDByPath failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if uuid == "" {
		logger.Info("[-USER-]", pathNotFoundMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: pathNotFoundMessage,
		}
	}
	return uuid, nil
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !fileUtils.IsFilenameOk(filename) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, parentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return "", models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Check if the file already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, filename, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, filename, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", filename)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	// Generate a new UUID.
	newFileUUID := m.uuidTool.NewUUID()
	logger = logger.WithField("newFileUUID", newFileUUID)
	// Save file to the local storage.
	relPathOD := filepath.Join(userUUID, newFileUUID)
	size, err := m.fileOps.SaveFile(relPathOD, contentReader)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveFile failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	newFile := models.File{
		Metadata: models.FileMetadata{
			UUID:          newFileUUID,
			Filename:      filename,
			RelPathOnDisk: relPathOD,
			ParentUUID:    parentDirUUID,
			Size:          size,
			OwnerUUID:     userUUID,
		},
	}
	// Insert new file metadata to the DB.
	if err := m.fileRepo.InsertFile(ctx, newFile); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s due to InsertFile error %s", newFileUUID, err.Error())
			err = m.fileOps.RemoveFile(newFileUUID)
			if err != nil {
				logger.Errorf("[-INTERNAL-] RemoveFile failed with error %s", err.Error())
			}
		}()
		logger.Errorf("[-INTERNAL-] InsertFile failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return newFileUUID, nil
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dstParentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return "", models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}
	isAuthorized, err = m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.CopyFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return "", models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}
	return m.copyFile(ctx, logger, userUUID, fileUUID, dstParentDirUUID)
}

func (m *MultiOSFileManager) copyFile(ctx context.Context, logger *log.Entry, userUUID, fileUUID, dstParentDirUUID string) (string, error) {
	// Get the source file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	// Check if the file already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	// Get source file reader to read its content.
	srcFReadCloser, err := m.fileOps.ReadFile(srcFile.RelPathOnDisk)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	defer srcFReadCloser.Close()

	// Generate a new UUID for the destination file.
	newFileUUID := m.uuidTool.NewUUID()
	logger = logger.WithField("fileUUID", newFileUUID)
	// Save the dst file to the disk.
	relDstPathOD := filepath.Join(userUUID, fileUUID)
	size, err := m.fileOps.SaveFile(relDstPathOD, srcFReadCloser)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveFile failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	newFile := models.File{
		Metadata: models.FileMetadata{
			UUID:          newFileUUID,
			Filename:      srcFile.Filename,
			RelPathOnDisk: relDstPathOD,
			ParentUUID:    dstParentDirUUID,
			Size:          size,
			OwnerUUID:     userUUID, // other user can copy the owner's file (if permitted)
		},
	}
	// Insert a new file metadata to the DB.
	if err = m.fileRepo.InsertFile(ctx, newFile); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s due to InsertFile error %s", newFileUUID, err.Error())
			err = m.fileOps.RemoveFile(newFileUUID)
			if err != nil {
				logger.Errorf("[-INTERNAL-] RemoveFile failed with error %s", err.Error())
			}
		}()
		logger.Errorf("[-INTERNAL-] InsertFile failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return newFileUUID, nil
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dstParentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}
	isAuthorized, err = m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.CopyFileAction, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get the file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	// Check if the file already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	err = m.fileRepo.UpdateParentDirUUID(ctx, fileUUID, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateParentDirUUID failed with error %s", err.Error())
		return models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.File{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return models.File{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.File{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.File{}, models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get the file metadata.
	srcFile, err := m.fileRepo.GetFile(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFile failed with error %s", err.Error())
		return models.File{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.FilePayload{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
		return models.FilePayload{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.FilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return models.FilePayload{}, models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	// Get the file metadata.
	srcFile, err := m.fileRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return models.FilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	fileRCloser, err := m.fileOps.ReadFile(srcFile.RelPathOnDisk)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		return models.FilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
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
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return models.TmpFilePayload{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	for _, uuid := range fileUUIDs {
		if !m.uuidTool.ValidateUUID(uuid) {
			logger.Info("[-USER-]", invalidFileUUIDErrorMessage)
			return models.TmpFilePayload{}, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileUUIDErrorMessage,
		}
		}
	}
	// Authorization.
	for _, uuid := range fileUUIDs {
		isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, uuid, author.ViewFileAction)
		if err != nil {
			logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
			return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
		}
		if !isAuthorized {
			logger.Info(forbiddenOperationErrorMessage)
			return models.TmpFilePayload{}, models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		}
	}
	// Get the files' metadata.
	srcFiles, err := m.fileRepo.GetFileMetadataBatch(ctx, fileUUIDs)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataBatch failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	inZipPaths := make(map[string]string)
	for _, srcFile := range srcFiles {
		inZipPaths[srcFile.Filename] = srcFile.RelPathOnDisk
	}
	tmpFilePath, err := m.fileCompress.CompressFiles(inZipPaths)
	if err != nil {
		logger.Errorf("[-INTERNAL-] CompressFiles failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	tmpFile, err := m.fileOps.GetFileReadCloserRmer(tmpFilePath)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileReadCloserRmer failed with error %s", err.Error())
		return models.TmpFilePayload{}, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
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
	logger.Debug("Start retrieving UUID via path")
	defer logger.Debug("Finish retrieving UUID via path")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return nil, nil, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !fileUtils.IsFilenameOk(filename) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		return nil, nil, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		return nil, nil, models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, parentDirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return nil, nil, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return nil, nil, models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}

	fileMetadataList, err := m.fileRepo.GetFileMetadataListByName(ctx, filename, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataListByName failed with error %s", err.Error())
		return nil, nil, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	dirMetadataList, err := m.dirRepo.GetDirMetadataListByName(ctx, filename, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadataListByName failed with error %s", err.Error())
		return nil, nil, models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return fileMetadataList, dirMetadataList, nil
}

func (m *MultiOSFileManager) CreateNewDirectory(ctx context.Context, userUUID, dirname, parentDirUUID string) (string, error) {
	logger := log.WithFields(log.Fields{
		"Loc":           "local-service_handler-multi_os",
		"Operation":     "CreateNewDirectory",
		"userUUID":      userUUID,
		"dirname":       dirname,
		"parentDirUUID": parentDirUUID,
	})
	logger.Debug("Start creating a new directory")
	defer logger.Debug("Finish creating a new directory")
	// Pre-validate inputs
	if m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
	}
	if !fileUtils.IsFilenameOk(dirname) {
		logger.Info("[-USER-]", invalidDirNameErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirNameErrorMessage,
		}
	}
	if m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", invalidParentDirUUIDErrorMessage)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidParentDirUUIDErrorMessage,
		}
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, parentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		return "", models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
	}
	return m.createNewDirectory(ctx, logger, userUUID, dirname, parentDirUUID)
}

func (m *MultiOSFileManager) createNewDirectory(ctx context.Context, logger *log.Entry, userUUID, dirname, parentDirUUID string) (string, error) {
	// Check if the directory already exists in a desired location in the db.
	var isExist bool
	var err error
	isExist, err = m.fileRepo.IsFilenameExist(ctx, dirname, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, dirname, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirname)
		return "", models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
	}
	// Generate a new UUID.
	newDirUUID := m.uuidTool.NewUUID()
	logger = logger.WithField("newDirUUID", newDirUUID)
	// Insert new file record to the DB.
	newDirMetadata := models.DirectoryMetadata{
		UUID:       newDirUUID,
		Dirname:    dirname,
		ParentUUID: parentDirUUID,
		OwnerUUID:  userUUID,
	}
	err = m.dirRepo.InsertDirectoryMetadata(ctx, newDirMetadata)
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertDirectoryMetadata failed with error %s", err.Error())
		return "", models.XtremeError{
			Code: models.InternalServerErrorCode,
			Message: err.Error(),
		}
	}
	return newDirUUID, nil
}
