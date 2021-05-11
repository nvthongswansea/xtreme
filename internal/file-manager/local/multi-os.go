package local

import (
	"bytes"
	"context"
	"errors"
	"github.com/nvthongswansea/xtreme/internal/author"
	"github.com/nvthongswansea/xtreme/internal/database"
	"github.com/nvthongswansea/xtreme/internal/models"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"path/filepath"
)

const (
	InvalidDirNameErrorMessage       = "directory name is invalid"
	InvalidFileNameErrorMessage      = "filename is invalid"
	InvalidUserUUIDErrorMessage      = "user UUID is not valid"
	InvalidParentDirUUIDErrorMessage = "parent directory UUID is not valid"
	InvalidFileUUIDErrorMessage      = "file UUID is not valid"
	InvalidDirUUIDErrorMessage       = "file UUID is not valid"
	InvalidPathErrorMessage          = "path is not valid"

	PathNotFoundMessage = "path not found"
	NameAlreadyExistErrorMessage            = "name already exists in desired location"
	ForbiddenOperationErrorMessage = "forbidden operation"

	InternalErrorMessage = "internal error"
)

// MultiOSLocalFManServiceHandler implements interface LocalFManServiceHandler.
// This implementation of LocalFManServiceHandler support multiple Operating Systems.
type MultiOSLocalFManServiceHandler struct {
	localFManDBRepo database.LocalFManRepository
	uuidTool        uuidUtils.UUIDGenerateValidator
	fileOps         fileUtils.FileSaveReadRemover
	fileCompress    fileUtils.FileCompressor
	author author.Authorizer
}

func (m *MultiOSLocalFManServiceHandler) GetRootDirectory(ctx context.Context, userUUID string) (models.Directory, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":         "local-service_handler-multi_os",
		"Operation":   "GetRootDirectory",
		"userUUID":    userUUID,
	})
	logger.Debug("Start retrieving root directory")
	defer logger.Debug("Finish retrieving root directory")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.Directory{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	// No need to authorize.
	rooDir, err := m.localFManDBRepo.GetRootDirectory(ctx, userUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetRootDirectory failed with error %s", err.Error())
		return models.Directory{}, err
	}
	return rooDir, nil
}

// NewMultiOSLocalFManServiceHandler creates a new MultiOSLocalFManServiceHandler.
func NewMultiOSLocalFManServiceHandler(localFManDBRepo database.LocalFManRepository, uuidTool uuidUtils.UUIDGenerateValidator,
	fileOps fileUtils.FileSaveReadRemover, fileCompress fileUtils.FileCompressor, author author.Authorizer) *MultiOSLocalFManServiceHandler {
	return &MultiOSLocalFManServiceHandler{
		localFManDBRepo,
		uuidTool,
		fileOps,
		fileCompress,
		author,
	}
}
func (m *MultiOSLocalFManServiceHandler) RenameFile(ctx context.Context, userUUID, fileUUID, newFileName string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":         "local-service_handler-multi_os",
		"Operation":   "RenameFile",
		"userUUID":    userUUID,
		"fileUUID":    fileUUID,
		"newFileName": newFileName,
	})
	logger.Debug("Start renaming file")
	defer logger.Debug("Finish renaming file")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return errors.New(InvalidFileUUIDErrorMessage)
	}
	if !fileUtils.IsFilenameOk(newFileName) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return errors.New(InvalidFileNameErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.UpdateFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return errors.New(ForbiddenOperationErrorMessage)
	}

	// Get the source file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	// Check if the file already exists in a current location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, newFileName, srcFile.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", newFileName)
		return errors.New(NameAlreadyExistErrorMessage)
	}
	srcFile.Filename = newFileName
	err = m.localFManDBRepo.UpdateFileMetadata(ctx, srcFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateFileMetadata failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) SoftRemoveFile(ctx context.Context, userUUID, fileUUID string) error {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return errors.New(InvalidFileUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return errors.New(ForbiddenOperationErrorMessage)
	}

	err = m.localFManDBRepo.SoftRemoveFile(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SoftRemoveFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) HardRemoveFile(ctx context.Context, userUUID, fileUUID string) error {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return errors.New(InvalidFileUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return errors.New(ForbiddenOperationErrorMessage)
	}

	err = m.localFManDBRepo.HardRemoveFile(ctx, fileUUID, m.fileOps.RemoveFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] HardRemoveFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) GetDirectory(ctx context.Context, userUUID, dirUUID string) (models.Directory, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.Directory{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return models.Directory{}, errors.New(InvalidDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.Directory{},errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return models.Directory{}, errors.New(ForbiddenOperationErrorMessage)
	}

	directory, err := m.localFManDBRepo.GetDirectory(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return models.Directory{}, errors.New(InternalErrorMessage)
	}
	return directory, nil
}

func (m *MultiOSLocalFManServiceHandler) CopyDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return "", errors.New(InvalidDirUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return "", errors.New(InvalidDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.CopyDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}
	return m.copyDirectory(ctx, logger, userUUID, dirUUID, dstParentDirUUID)
}

func (m *MultiOSLocalFManServiceHandler) copyDirectory(ctx context.Context, logger *log.Entry, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	// Get to-be-copied dir metadata.
	copiedDirMeta, err := m.localFManDBRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	logger.WithField("currentCopyPath", copiedDirMeta.Path)
	// Copy the current dir to the new location
	nDirCopyUUID, err := m.createNewDirectory(ctx, logger, userUUID, copiedDirMeta.Dirname, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] createNewDirectory failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}

	// Copy files to the newly created directory.
	childFileMetaList, err := m.localFManDBRepo.GetChildFileMetadataList(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetChildFileMetadataList failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	for _, childMeta := range childFileMetaList {
		_, err := m.copyFile(ctx, logger, userUUID, childMeta.UUID, nDirCopyUUID)
		if err != nil {
			logger.Errorf("[-INTERNAL-] copyFile failed with error %s", err.Error())
			return "", errors.New(InternalErrorMessage)
		}
	}

	// Get direct child-directories' UUIDs of the current to-be-copied dir.
	dChildDirUUIDList, err := m.localFManDBRepo.GetDirectChildDirUUIDList(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirectChildDirUUIDList failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	for _, childDirUUID := range dChildDirUUIDList {
		// Recursively do copy the child directories with their child files/directories.
		_, err := m.copyDirectory(ctx, logger, userUUID, childDirUUID, nDirCopyUUID)
		if err != nil {
			logger.Errorf("[-INTERNAL-] copyDirectory failed with error %s", err.Error())
			return "", errors.New(InternalErrorMessage)
		}
	}
	return nDirCopyUUID, nil
}

func (m *MultiOSLocalFManServiceHandler) RenameDirectory(ctx context.Context, userUUID, dirUUID, newDirName string) error {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":        "local-service_handler-multi_os",
		"Operation":  "RenameDirectory",
		"userUUID":   userUUID,
		"dirUUID":    dirUUID,
		"newDirName": newDirName,
	})
	logger.Debug("Start renaming directory")
	defer logger.Debug("Finish renaming directory")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return errors.New(InvalidDirUUIDErrorMessage)
	}
	if !fileUtils.IsFilenameOk(newDirName) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return errors.New(InvalidFileNameErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.UpdateDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return errors.New(ForbiddenOperationErrorMessage)
	}

	// Get the source directory metadata.
	dirMeta, err := m.localFManDBRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if (dirMeta == models.DirectoryMetadata{}) {
		logger.Infof("[-USER-] directory UUID (%s) does not exist", dirUUID)
		return errors.New(InvalidDirUUIDErrorMessage)
	}
	// Check if the name already exists in a current location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, newDirName, dirMeta.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", newDirName)
		return errors.New(NameAlreadyExistErrorMessage)
	}
	dirMeta.Dirname = newDirName
	err = m.localFManDBRepo.UpdateDirMetadata(ctx, dirMeta)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateDirMetadata failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) MoveDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return "", errors.New(InvalidDirUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return "", errors.New(InvalidDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.CopyDirAction, author.RemoveDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}

	// Get the directory metadata.
	dirMeta, err := m.localFManDBRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	// Check if the directory name already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, dirMeta.Dirname, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirMeta.Dirname)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	dirMeta.ParentUUID = dstParentDirUUID
	err = m.localFManDBRepo.UpdateDirMetadata(ctx, dirMeta)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateDirMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	return dirUUID, nil
}

func (m *MultiOSLocalFManServiceHandler) DownloadDirectory(ctx context.Context, userUUID, dirUUID string) (models.TmpFilePayload, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.TmpFilePayload{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return models.TmpFilePayload{}, errors.New(InvalidDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return models.TmpFilePayload{},errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return models.TmpFilePayload{}, errors.New(ForbiddenOperationErrorMessage)
	}

	// Get dir metadata.
	dirMeta, err := m.localFManDBRepo.GetDirMetadata(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}

	// Get all child-files metadata.
	fileMetadataList, err := m.localFManDBRepo.GetChildFileMetadataList(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetChildFileMetadataList failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}

	inZipPaths := make(map[string]string)
	for _, childFileMeta := range fileMetadataList {
		relPath, err := filepath.Rel(dirMeta.Path, childFileMeta.Path)
		if err != nil {
			logger.Errorf("[-INTERNAL-] filepath.Rel failed with error %s", err.Error())
			return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
		}
		inZipPaths[relPath] = childFileMeta.AbsPathOnDisk
	}
	tmpFile, err := m.fileCompress.CompressFiles(inZipPaths)
	if err != nil {
		logger.Errorf("[-INTERNAL-] CompressFiles failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	// Get zipped file size
	tmpFStat, err := tmpFile.Stat()
	if err != nil {
		logger.Errorf("[-INTERNAL-] Failed to read file stat %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	payload := models.TmpFilePayload{
		Filename:      tmpFile.Name(),
		ContentLength: tmpFStat.Size(),
		File:          tmpFile,
	}
	return payload, nil
}

func (m *MultiOSLocalFManServiceHandler) SoftRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return errors.New(InvalidDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.RemoveDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return errors.New(ForbiddenOperationErrorMessage)
	}

	err = m.localFManDBRepo.SoftRemoveDir(ctx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SoftRemoveFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) HardRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", InvalidDirUUIDErrorMessage)
		return errors.New(InvalidDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dirUUID, author.RemoveDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return errors.New(ForbiddenOperationErrorMessage)
	}

	err = m.localFManDBRepo.HardRemoveDir(ctx, dirUUID, m.fileOps.RemoveFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SoftRemoveFile failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) GetUUIDByPath(ctx context.Context, userUUID, path string) (string, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetUUIDByPath",
		"userUUID":  userUUID,
		"path":      path,
	})
	logger.Debug("Start retrieving UUID via path")
	defer logger.Debug("Finish retrieving UUID via path")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	path = filepath.Clean(path)
	if path == "" {
		logger.Info("[-USER-]", InvalidPathErrorMessage)
		return "", errors.New(InvalidPathErrorMessage)
	}
	// Get root dir metadata of the user
	rootDirMeta, err := m.localFManDBRepo.GetRootDirMetadata(ctx, userUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetUUIDByPath failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}

	uuid, isDir, err := m.localFManDBRepo.GetUUIDByPath(ctx, rootDirMeta.UUID, path)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetUUIDByPath failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if uuid == "" {
		logger.Info("[-USER-]", PathNotFoundMessage)
		return "", errors.New(PathNotFoundMessage)
	}
	// Authorization. Note: Since we only
	// know if the user has permission to view this UUID
	// after retrieving the UUID, the authorization is executed
	// at last.
	if isDir {
		isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, uuid, author.ViewDirAction)
		if err != nil {
			logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
			return "", errors.New(InternalErrorMessage)
		}
		if !isAuthorized {
			logger.Info(ForbiddenOperationErrorMessage)
			return "", errors.New(ForbiddenOperationErrorMessage)
		}
		return uuid, nil
	}
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, uuid, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}
	return uuid, nil
}

// CreateNewFile creates a new empty file with a given filename in a specific user storage space.
func (m *MultiOSLocalFManServiceHandler) CreateNewFile(ctx context.Context, userUUID, filename, parentDirUUID string) (string, error) {
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
	// Create a new ReadCloser
	contentReadCloser := ioutil.NopCloser(bytes.NewReader(emptyBytes))
	return m.UploadFile(ctx, userUUID, filename, parentDirUUID, contentReadCloser)
}

func (m *MultiOSLocalFManServiceHandler) UploadFile(ctx context.Context, userUUID, filename, parentDirUUID string, fileReadCloser io.ReadCloser) (string, error) {
	// Make sure the fileReadCloser is closed at the end.
	defer fileReadCloser.Close()
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !fileUtils.IsFilenameOk(filename) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return "", errors.New(InvalidFileNameErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, parentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}

	// Check if the file already exists in a desired location in the db.
	isFilenameExist, err := m.localFManDBRepo.IsNameExist(ctx, filename, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isFilenameExist {
		logger.Infof("[-USER-] %s already exists in the desired location", filename)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	// Generate a new UUID.
	newFileUUID := m.uuidTool.NewUUID()
	logger = logger.WithField("newFileUUID", newFileUUID)
	// Save file to the local disk.
	relFilePath := filepath.Join(userUUID, newFileUUID)
	size, absPathOD, err := m.fileOps.SaveCloseFile(relFilePath, fileReadCloser)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveCloseFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	newFileMetadata := models.FileMetadata{
		UUID:          newFileUUID,
		Filename:      filename,
		AbsPathOnDisk: absPathOD,
		ParentUUID:    parentDirUUID,
		Size:          size,
		OwnerUUID:     userUUID,
	}
	// Insert new file metadata to the DB.
	if err := m.localFManDBRepo.InsertFileMetadata(ctx, newFileMetadata); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s due to InsertFileMetadata error %s", newFileUUID, err.Error())
			err = m.fileOps.RemoveFile(newFileUUID)
			if err != nil {
				logger.Errorf("[-INTERNAL-] RemoveFile failed with error %s", err.Error())
			}
		}()
		logger.Errorf("[-INTERNAL-] InsertFileRecord failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	return newFileUUID, nil
}

func (m *MultiOSLocalFManServiceHandler) CopyFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return "", errors.New(InvalidFileUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dstParentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}
	isAuthorized, err = m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.CopyFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}
	return m.copyFile(ctx, logger, userUUID, fileUUID, dstParentDirUUID)
}

func (m *MultiOSLocalFManServiceHandler) copyFile(ctx context.Context, logger *log.Entry, userUUID, fileUUID, dstParentDirUUID string) (string, error) {
	// Get the source file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	// Get source file reader to read its content.
	srcFReadCloser, err := m.fileOps.ReadFile(srcFile.AbsPathOnDisk)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}

	// Generate a new UUID for the destination file.
	newFileUUID := m.uuidTool.NewUUID()
	logger = logger.WithField("fileUUID", newFileUUID)
	// Save the dst file to the disk.
	relDstFilePath := filepath.Join(userUUID, fileUUID)
	size, realPath, err := m.fileOps.SaveCloseFile(relDstFilePath, srcFReadCloser)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveCloseFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	newFileMetadata := models.FileMetadata{
		UUID:          newFileUUID,
		Filename:      srcFile.Filename,
		AbsPathOnDisk: realPath,
		ParentUUID:    dstParentDirUUID,
		Size:          size,
		OwnerUUID:     userUUID, // other user can copy the owner's file (if permited)
	}
	// Insert a new file metadata to the DB.
	if err = m.localFManDBRepo.InsertFileMetadata(ctx, newFileMetadata); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s due to InsertFileMetadata error %s", newFileUUID, err.Error())
			err = m.fileOps.RemoveFile(newFileUUID)
			if err != nil {
				logger.Errorf("[-INTERNAL-] RemoveFile failed with error %s", err.Error())
			}
		}()
		logger.Errorf("[-INTERNAL-] InsertFileRecord failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	return newFileUUID, nil
}

func (m *MultiOSLocalFManServiceHandler) MoveFile(ctx context.Context, userUUID, fileUUID, dstParentDirUUID string) (string, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return "", errors.New(InvalidFileUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, dstParentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}
	isAuthorized, err = m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.CopyFileAction, author.RemoveFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}

	// Get the file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	srcFile.ParentUUID = dstParentDirUUID
	err = m.localFManDBRepo.UpdateFileMetadata(ctx, srcFile)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateFileMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	return fileUUID, nil
}

func (m *MultiOSLocalFManServiceHandler) GetFile(ctx context.Context, userUUID, fileUUID string) (models.File, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.File{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return models.File{}, errors.New(InvalidFileUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.File{}, errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return models.File{}, errors.New(ForbiddenOperationErrorMessage)
	}

	// Get the file metadata.
	srcFile, err := m.localFManDBRepo.GetFile(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFile failed with error %s", err.Error())
		return models.File{}, errors.New(InternalErrorMessage)
	}
	return srcFile, nil
}

func (m *MultiOSLocalFManServiceHandler) DownloadFile(ctx context.Context, userUUID, fileUUID string) (models.FilePayload, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.FilePayload{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return models.FilePayload{}, errors.New(InvalidFileUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, fileUUID, author.ViewFileAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
		return models.FilePayload{}, errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return models.FilePayload{}, errors.New(ForbiddenOperationErrorMessage)
	}

	// Get the file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return models.FilePayload{}, errors.New(InternalErrorMessage)
	}
	fileRCloser, err := m.fileOps.ReadFile(srcFile.AbsPathOnDisk)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		return models.FilePayload{}, errors.New(InternalErrorMessage)
	}
	payload := models.FilePayload{
		Filename:      srcFile.Filename,
		ContentLength: srcFile.Size,
		ReadCloser:    fileRCloser,
	}
	return payload, nil
}

func (m *MultiOSLocalFManServiceHandler) DownloadFileBatch(ctx context.Context, userUUID string, fileUUIDs []string) (models.TmpFilePayload, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.TmpFilePayload{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	for _, uuid := range fileUUIDs {
		if !m.uuidTool.ValidateUUID(uuid) {
			logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
			return models.TmpFilePayload{}, errors.New(InvalidFileUUIDErrorMessage)
		}
	}
	// Authorization.
	for _, uuid := range fileUUIDs {
		isAuthorized, err := m.author.AuthorizeActionsOnFile(ctx, userUUID, uuid, author.ViewFileAction)
		if err != nil {
			logger.Errorf("[-INTERNAL-] AuthorizeActionsOnFile failed with error %s", err.Error())
			return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
		}
		if !isAuthorized {
			logger.Info(ForbiddenOperationErrorMessage)
			return models.TmpFilePayload{}, errors.New(ForbiddenOperationErrorMessage)
		}
	}
	// Get the files' metadata.
	srcFiles, err := m.localFManDBRepo.GetFileMetadataBatch(ctx, fileUUIDs)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataBatch failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	inZipPaths := make(map[string]string)
	for _, srcFile := range srcFiles {
		inZipPaths[srcFile.Filename] = srcFile.AbsPathOnDisk
	}
	tmpFile, err := m.fileCompress.CompressFiles(inZipPaths)
	if err != nil {
		logger.Errorf("[-INTERNAL-] CompressFiles failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	// Get zipped file size
	tmpFStat, err := tmpFile.Stat()
	if err != nil {
		logger.Errorf("[-INTERNAL-] Failed to read file stat %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	payload := models.TmpFilePayload{
		Filename:      tmpFile.Name(),
		ContentLength: tmpFStat.Size(),
		File:          tmpFile,
	}
	return payload, nil
}

func (m *MultiOSLocalFManServiceHandler) SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.File, []models.Directory, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return nil, nil, errors.New(InvalidUserUUIDErrorMessage)
	}
	if !fileUtils.IsFilenameOk(filename) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return nil, nil, errors.New(InvalidFileNameErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return nil, nil, errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, parentDirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return nil, nil, errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return nil, nil, errors.New(ForbiddenOperationErrorMessage)
	}

	files, dirs, err := m.localFManDBRepo.SearchByName(ctx, userUUID, filename, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetUUIDByPath failed with error %s", err.Error())
		return nil, nil, errors.New(InternalErrorMessage)
	}
	return files, dirs, nil
}

func (m *MultiOSLocalFManServiceHandler) CreateNewDirectory(ctx context.Context, userUUID, dirname, parentDirUUID string) (string, error) {
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
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !fileUtils.IsFilenameOk(dirname) {
		logger.Info("[-USER-]", InvalidDirNameErrorMessage)
		return "", errors.New(InvalidDirNameErrorMessage)
	}
	if m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidDirNameErrorMessage)
		return "", errors.New(InvalidDirNameErrorMessage)
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, userUUID, parentDirUUID, author.UploadToDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isAuthorized {
		logger.Info(ForbiddenOperationErrorMessage)
		return "", errors.New(ForbiddenOperationErrorMessage)
	}

	return m.createNewDirectory(ctx, logger, userUUID, dirname, parentDirUUID)
}

func (m *MultiOSLocalFManServiceHandler) createNewDirectory(ctx context.Context, logger *log.Entry, userUUID, dirname, parentDirUUID string) (string, error) {
	// Check if the directory already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, dirname, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirname)
		return "", errors.New(NameAlreadyExistErrorMessage)
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
	err = m.localFManDBRepo.InsertDirectoryMetadata(ctx, newDirMetadata)
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertDirRecord failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	return newDirUUID, nil
}