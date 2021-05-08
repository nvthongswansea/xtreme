package svhander

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"

	"github.com/nvthongswansea/xtreme/internal/file-manager/local"
	"github.com/nvthongswansea/xtreme/internal/models"
	fileUtils "github.com/nvthongswansea/xtreme/pkg/file-utils"
	uuidUtils "github.com/nvthongswansea/xtreme/pkg/uuid-utils"
	log "github.com/sirupsen/logrus"
)

const (
	MissingFileNameErrorMessage      = "Missing filename"
	MissingPathErrorMessage          = "Missing path"
	MissingDirNameErrorMessage       = "Missing directory name"
	MissingFileUUIDErrorMessage      = "Missing file UUID"
	MissingParentDirUUIDErrorMessage = "Missing parent directory UUID"

	InvalidUserUUIDErrorMessage      = "user UUID is not valid"
	InvalidParentDirUUIDErrorMessage = "parent directory UUID is not valid"
	InvalidFileUUIDErrorMessage      = "file UUID is not valid"
	InvalidPathErrorMessage          = "path is not valid"

	NameAlreadyExistErrorMessage = "Name already exists in desired location"

	InternalErrorMessage = "Internal error"
)

// MultiOSLocalFManServiceHandler implements interface LocalFManServiceHandler.
// This implementation of LocalFManServiceHandler support multiple Operating Systems.
type MultiOSLocalFManServiceHandler struct {
	locaFManDBRepo local.LocalFManDBRepo
	uuidTool       uuidUtils.UUIDGenerateValidator
	fileOps        fileUtils.FileSaveReadRemover
	fileCompress   fileUtils.FileCompressor
}

// MultiOSLocalFManServiceHandler creates a new MultiOSLocalFManServiceHandler.
func NewMultiOSLocalFManServiceHandler(locaFManDBRepo local.LocalFManDBRepo, uuidTool uuidUtils.UUIDGenerateValidator,
	fileOps fileUtils.FileSaveReadRemover, fileCompress fileUtils.FileCompressor) *MultiOSLocalFManServiceHandler {
	return &MultiOSLocalFManServiceHandler{
		locaFManDBRepo,
		uuidTool,
		fileOps,
		fileCompress,
	}
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
	if path == "" {
		logger.Info("[-USER-]", MissingPathErrorMessage)
		return "", errors.New(MissingPathErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	uuid, err := m.locaFManDBRepo.GetUUIDByPath(ctx, userUUID, path)
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
	contentReadCloser := ioutil.NopCloser((bytes.NewReader(emptyBytes)))
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
	if filename == "" {
		logger.Info("[-USER-]", MissingFileNameErrorMessage)
		return "", errors.New(MissingFileNameErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	// Validate parent UUID.
	isParentDirExist, err := m.locaFManDBRepo.IsParentDirExist(ctx, userUUID, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentUUIDExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", parentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isFilenameExist, err := m.locaFManDBRepo.IsNameExist(ctx, userUUID, filename, parentDirUUID)
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
	// Insert new file record to the DB.
	if err := m.locaFManDBRepo.InsertFileRecord(ctx, userUUID, newFileUUID, filename, parentDirUUID, realAbsPath, size); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s due to InsertFileRecord error %s", newFileUUID, err.Error())
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
	isParentDirExist, err := m.locaFManDBRepo.IsParentDirExist(ctx, userUUID, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentDirExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", dstParentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Get the source file record.
	srcFile, err := m.locaFManDBRepo.GetFileRecord(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileRecord failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", dstParentDirUUID)
		return "", errors.New(InvalidFileUUIDErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := m.locaFManDBRepo.IsNameExist(ctx, userUUID, srcFile.Filename, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsNameExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", srcFile.Filename)
		return "", errors.New(NameAlreadyExistErrorMessage)
	}
	// Get source file pointer to read its content.
	srcFReadCloser, err := m.fileOps.ReadFile(srcFile.RealPath)
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
	// Insert a new file record to the DB.
	if err = m.locaFManDBRepo.InsertFileRecord(ctx, userUUID, newFileUUID, srcFile.Filename, dstParentDirUUID, realPath, size); err != nil {
		// If error presents while inserting a new record,
		// remove the file from the storage.
		defer func() {
			logger.Debugf("Removing file %s due to InsertFileRecord error %s", newFileUUID, err.Error())
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
	isParentDirExist, err := m.locaFManDBRepo.IsParentDirExist(ctx, userUUID, dstParentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentDirExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", dstParentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Get the file record.
	srcFile, err := m.locaFManDBRepo.GetFileRecord(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileRecord failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return "", errors.New(InvalidFileUUIDErrorMessage)
	}
	// Check if the file already exists in a desired location in the db.
	isExist, err := m.locaFManDBRepo.IsNameExist(ctx, userUUID, srcFile.Filename, dstParentDirUUID)
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
		err = m.locaFManDBRepo.UpdateFileRecord(ctx, userUUID, srcFile.Filename, dstParentDirUUID)
		if err != nil {
			logger.Errorf("[-INTERNAL-] InsertFileRecord failed with error %s", err.Error())
			return "", errors.New(InternalErrorMessage)
		}
	}
	return fileUUID, nil
}

func (m *MultiOSLocalFManServiceHandler) GetFileMeta(ctx context.Context, userUUID, fileUUID string) (models.FileMetadata, error) {
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetFileMeta",
		"userUUID":  userUUID,
		"fileUUID":  fileUUID,
	})
	logger.Debug("Start retrieving file metadata")
	defer logger.Debug("Finish retrieving file metadata")
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return models.FileMetadata{}, errors.New(InvalidUserUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(fileUUID) {
		logger.Info("[-USER-]", InvalidFileUUIDErrorMessage)
		return models.FileMetadata{}, errors.New(InvalidFileUUIDErrorMessage)
	}
	// Get the file record.
	srcFile, err := m.locaFManDBRepo.GetFileRecord(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileRecord failed with error %s", err.Error())
		return models.FileMetadata{}, errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return models.FileMetadata{}, errors.New(InvalidFileUUIDErrorMessage)
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
	// Get the file record.
	srcFile, err := m.locaFManDBRepo.GetFileRecord(ctx, userUUID, fileUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileRecord failed with error %s", err.Error())
		return models.FilePayload{}, errors.New(InternalErrorMessage)
	}
	if (srcFile == models.FileMetadata{}) {
		logger.Infof("[-USER-] file UUID (%s) does not exist", fileUUID)
		return models.FilePayload{}, errors.New(InvalidFileUUIDErrorMessage)
	}
	fileRCloser, err := m.fileOps.ReadFile(srcFile.RealPath)
	if err != nil {
		logger.Errorf("[-INTERNAL-] ReadFile failed with error %s", err.Error())
		return models.FilePayload{}, errors.New(InternalErrorMessage)
	}
	payload := models.FilePayload{
		Filename:      srcFile.Filename,
		ContentLength: srcFile.FileSize,
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
	// Get the file records.
	srcFiles, err := m.locaFManDBRepo.GetFileRecordBatch(ctx, userUUID, fileUUIDs)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileRecord failed with error %s", err.Error())
		return models.TmpFilePayload{}, errors.New(InternalErrorMessage)
	}
	if len(srcFiles) == 0 {
		logger.Infof("[-USER-] all file UUIDs (%v) do not exist", fileUUIDs)
		return models.TmpFilePayload{}, errors.New(InvalidFileUUIDErrorMessage)
	}
	inZipPaths := make(map[string]string)
	for _, srcFile := range srcFiles {
		inZipPaths[srcFile.Filename] = srcFile.RealPath
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

func (m *MultiOSLocalFManServiceHandler) SearchByName(ctx context.Context, userUUID, filename, parentDirUUID string) ([]models.FileMetadata, []models.DirectoryMetadata, error) {
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
	if !m.uuidTool.ValidateUUID(parentDirUUID) {
		logger.Info("[-USER-]", InvalidParentDirUUIDErrorMessage)
		return nil, nil, errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Validate parent UUID.
	isParentDirExist, err := m.locaFManDBRepo.IsParentDirExist(ctx, userUUID, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentUUIDExist failed with error %s", err.Error())
		return nil, nil, errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", parentDirUUID)
		return nil, nil, errors.New(InvalidParentDirUUIDErrorMessage)
	}
	files, dirs, err := m.locaFManDBRepo.SearchByName(ctx, userUUID, filename, parentDirUUID)
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
	if dirname == "" {
		logger.Info("[-USER-]", MissingDirNameErrorMessage)
		return "", errors.New(MissingDirNameErrorMessage)
	}
	if parentDirUUID == "" {
		logger.Info("[-USER-]", MissingParentDirUUIDErrorMessage)
		return "", errors.New(MissingParentDirUUIDErrorMessage)
	}
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", InvalidUserUUIDErrorMessage)
		return "", errors.New(InvalidUserUUIDErrorMessage)
	}
	// Validate parent UUID.
	isParentDirExist, err := m.locaFManDBRepo.IsParentDirExist(ctx, userUUID, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsParentUUIDExist failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	if !isParentDirExist {
		logger.Infof("[-USER-] parent directory UUID (%s) does not exist", parentDirUUID)
		return "", errors.New(InvalidParentDirUUIDErrorMessage)
	}
	// Check if the directory already exists in a desired location in the db.
	isExist, err := m.locaFManDBRepo.IsNameExist(ctx, userUUID, dirname, parentDirUUID)
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
	err = m.locaFManDBRepo.InsertDirRecord(ctx, userUUID, newDirUUID, dirname, parentDirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertDirRecord failed with error %s", err.Error())
		return "", errors.New(InternalErrorMessage)
	}
	return newDirUUID, nil
}
