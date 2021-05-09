package svhander

import (
	"bytes"
	"context"
	"errors"
	"github.com/nvthongswansea/xtreme/internal/database"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/nvthongswansea/xtreme/internal/models"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
	log "github.com/sirupsen/logrus"
)

const (
	InvalidDirNameErrorMessage       = "directory name is invalid"
	InvalidFileNameErrorMessage      = "filename is invalid"
	InvalidUserUUIDErrorMessage      = "user UUID is not valid"
	InvalidParentDirUUIDErrorMessage = "parent directory UUID is not valid"
	InvalidFileUUIDErrorMessage      = "file UUID is not valid"
	InvalidPathErrorMessage          = "path is not valid"

	NameAlreadyExistErrorMessage = "name already exists in desired location"

	InternalErrorMessage = "internal error"
)

// MultiOSLocalFManServiceHandler implements interface LocalFManServiceHandler.
// This implementation of LocalFManServiceHandler support multiple Operating Systems.
type MultiOSLocalFManServiceHandler struct {
	localFManDBRepo  database.LocalFManRepository
	uuidTool         uuidUtils.UUIDGenerateValidator
	fileOps          fileUtils.FileSaveReadRemover
	fileCompress     fileUtils.FileCompressor
	validateFilename fileUtils.FilenameValidator
}

// MultiOSLocalFManServiceHandler creates a new MultiOSLocalFManServiceHandler.
func NewMultiOSLocalFManServiceHandler(localFManDBRepo database.LocalFManRepository, uuidTool uuidUtils.UUIDGenerateValidator,
	fileOps fileUtils.FileSaveReadRemover, fileCompress fileUtils.FileCompressor, validateFilename fileUtils.FilenameValidator) *MultiOSLocalFManServiceHandler {
	return &MultiOSLocalFManServiceHandler{
		localFManDBRepo,
		uuidTool,
		fileOps,
		fileCompress,
		validateFilename,
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
	if !m.validateFilename(newFileName) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return errors.New(InvalidFileNameErrorMessage)
	}
	// Get the source file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return  errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return errors.New(InvalidFileUUIDErrorMessage)
	}
	// Check if the file already exists in a current location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, userUUID, srcFile.Filename, srcFile.ParentUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return errors.New(NameAlreadyExistErrorMessage)
	}
	// If the filename is changed, update the filename field.
	if srcFile.Filename != newFileName {
		srcFile.Filename = newFileName
		err = m.localFManDBRepo.UpdateFileMetadata(ctx, srcFile)
		if err != nil {
			logger.Errorf("[-INTERNAL-] UpdateFileRecord failed with error %s", err.Error())
			return errors.New(InternalErrorMessage)
		}
	}
	return nil
}

func (m *MultiOSLocalFManServiceHandler) SoftRemoveFile(ctx context.Context, userUUID, fileUUID string) error {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) HardRemoveFile(ctx context.Context, userUUID, fileUUID string) error {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) GetDirectoryMeta(ctx context.Context, userUUID, dirUUID string) (models.Directory, error) {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) CopyDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) RenameDirectory(ctx context.Context, userUUID, dirUUID, newDirName string) error {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) MoveDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) DownloadDirectory(ctx context.Context, userUUID, dirUUID string) (models.TmpFilePayload, error) {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) SoftRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
	panic("implement me")
}

func (m *MultiOSLocalFManServiceHandler) HardRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
	panic("implement me")
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
	if path == "" {
		logger.Info("[-USER-]", InvalidPathErrorMessage)
		return "", errors.New(InvalidPathErrorMessage)
	}
	uuid, err := m.localFManDBRepo.GetUUIDByPath(ctx, userUUID, path)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetUUIDByPath failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if uuid == "" {
		logger.Info("[-USER-]", InvalidPathErrorMessage)
		return "", errors.New(InvalidPathErrorMessage)
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
	if !m.validateFilename(filename) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return "", errors.New(InvalidFileNameErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Validate parent UUID.
	isParentDirExist, err := m.localFManDBRepo.IsDirExist(ctx, userUUID, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", parentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isFilenameExist, err := m.localFManDBRepo.IsNameExist(ctx, userUUID, filename, parentDirUUID)
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
	size, realAbsPath, err := m.fileOps.SaveCloseFile(relFilePath, fileReadCloser)
	if err != nil {
		logger.Errorf("[-INTERNAL-] SaveCloseFile failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	newFileMetadata := models.FileMetadata{
		UUID:          newFileUUID,
		Filename:      filename,
		AbsPathOnDisk: realAbsPath,
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
	// Validate dst parent UUID.
	isParentDirExist, err := m.localFManDBRepo.IsDirExist(ctx, userUUID, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", dstParentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Get the source file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return "", errors.New(InvalidFileUUIDErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, userUUID, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	// Get source file pointer to read its content.
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
		OwnerUUID:     userUUID,
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
	// Validate dst parent UUID.
	isParentDirExist, err := m.localFManDBRepo.IsDirExist(ctx, userUUID, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", dstParentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Get the file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return "", errors.New(InvalidFileUUIDErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, userUUID, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	// If the parentDirUUID is changed, update the parent UUID.
	if srcFile.ParentUUID != dstParentDirUUID {
		srcFile.ParentUUID = dstParentDirUUID
		err = m.localFManDBRepo.UpdateFileMetadata(ctx, srcFile)
		if err != nil {
			logger.Errorf("[-INTERNAL-] UpdateFileMetadata failed with error %s", err.Error())
			return "", errors.New(InternalErrorMessage)
		}
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
	// Get the file metadata.
	srcFile, err := m.localFManDBRepo.GetFile(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFile failed with error %s", err.Error())
		return models.File{}, errors.New(InternalErrorMessage)
	}
	if (srcFile == models.File{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return models.File{}, errors.New(InvalidFileUUIDErrorMessage)
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
	// Get the file metadata.
	srcFile, err := m.localFManDBRepo.GetFileMetadata(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadata failed with error %s", err.Error())
		return models.FilePayload{}, errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return models.FilePayload{}, errors.New(InvalidFileUUIDErrorMessage)
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
	// Get the files' metadata.
	srcFiles, err := m.localFManDBRepo.GetFileMetadataBatch(ctx, userUUID, fileUUIDs)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataBatch failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	if len(srcFiles) == 0 {
		logger.Infof("[-USER-] all file UUIDs (%v) do not exist", fileUUIDs)
		return models.TmpFilePayload{}, errors.New(InvalidFileUUIDErrorMessage)
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
	if !m.validateFilename(filename) {
		logger.Info("[-USER-]", InvalidFileNameErrorMessage)
		return nil, nil, errors.New(InvalidFileNameErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return nil, nil, errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Validate parent UUID.
	isParentDirExist, err := m.localFManDBRepo.IsDirExist(ctx, userUUID, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirExist failed with error %s", err.Error())
		return nil, nil, errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", parentDirUUID)
		return nil, nil, errors.New(InvalidParentDirUUIDErrorMessage)
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
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.validateFilename(dirname) {
		logger.Info("[-USER-]", InvalidDirNameErrorMessage)
		return "", errors.New(InvalidDirNameErrorMessage)
	}
	if m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidDirNameErrorMessage)
		return "", errors.New(InvalidDirNameErrorMessage)
	}
	// Validate parent UUID.
	isParentDirExist, err := m.localFManDBRepo.IsDirExist(ctx, userUUID, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", parentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Check if the directory already exists in a desired location in the db.
	isExist, err := m.localFManDBRepo.IsNameExist(ctx, userUUID, dirname, parentDirUUID)
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
