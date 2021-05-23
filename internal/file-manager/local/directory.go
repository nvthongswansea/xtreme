package local

import (
	"context"
	"github.com/nvthongswansea/xtreme/internal/author"
	"github.com/nvthongswansea/xtreme/internal/models"
	"github.com/nvthongswansea/xtreme/internal/repository/transaction"
	"github.com/nvthongswansea/xtreme/pkg/fileUtils"
	log "github.com/sirupsen/logrus"
	"path/filepath"
)

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
	if !fileUtils.IsFilenameOk(dirname) {
		logger.Info("[-USER-]", invalidDirNameErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirNameErrorMessage,
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
	return m.createNewDirectory(ctx, tx, logger, userUUID, dirname, parentDirUUID)
}

func (m *MultiOSFileManager) createNewDirectory(ctx context.Context, tx transaction.RollbackCommitter, logger *log.Entry, userUUID, dirname, parentDirUUID string) (string, error) {
	// Check if the directory already exists in a desired location in the db.
	var isExist bool
	var err error
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, parentDirUUID, dirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, parentDirUUID, dirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirname)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return "", err
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

	newDirUUID, err := m.dirRepo.InsertDirectory(ctx, tx, models.Directory{
		Metadata: models.DirectoryMetadata{
			Dirname:    dirname,
			ParentUUID: parentDirUUID,
			OwnerUUID:  userUUID,
			Path:       filepath.Join(parentDirMeta.Path, dirname),
		},
	})
	if err != nil {
		logger.Errorf("[-INTERNAL-] InsertDirectory failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	return newDirUUID, nil
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
	tx, err := m.txRepo.StartTransaction(ctx)
	if err != nil {
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.Directory{}, err
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
		return models.Directory{}, err
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
		return models.Directory{}, err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, dirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.Directory{}, err
	}
	if !isAuthorized {
		logger.Info(forbiddenOperationErrorMessage)
		err = models.XtremeError{
			Code:    models.ForbiddenOperationErrorCode,
			Message: forbiddenOperationErrorMessage,
		}
		return models.Directory{}, err
	}

	directory, err := m.dirRepo.GetDirectory(ctx, tx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirectory failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.Directory{}, err
	}
	return directory, nil
}

func (m *MultiOSFileManager) GetDirUUIDByPath(ctx context.Context, userUUID, path string) (string, error) {
	// Init logger header
	logger := log.WithFields(log.Fields{
		"Loc":       "local-service_handler-multi_os",
		"Operation": "GetDirUUIDByPath",
		"userUUID":  userUUID,
		"path":      path,
	})
	logger.Debug("Start retrieving UserUUID via path")
	defer logger.Debug("Finish retrieving UserUUID via path")
	var err error
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return "", err
	}
	path = filepath.Clean(path)
	if path == "" {
		logger.Info("[-USER-]", invalidPathErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidPathErrorMessage,
		}
		return "", err
	}
	uuid, err := m.dirRepo.GetDirUUIDByPath(ctx, nil, path, userUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirUUIDByPath failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if uuid == "" {
		logger.Info("[-USER-]", pathNotFoundMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: pathNotFoundMessage,
		}
		return "", err
	}
	return uuid, nil
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
	var err error
	// Pre-validate inputs
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return models.Directory{}, err
	}
	rooDir, err := m.dirRepo.GetRootDirectoryByUserUUID(ctx, nil, userUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetRootDirectoryByUserUUID failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.Directory{}, err
	}
	return rooDir, nil
}

func (m *MultiOSFileManager) CopyDirectory(ctx context.Context, userUUID, dirUUID, dstParentDirUUID string) (string, error) {
	panic("implement me")
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
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
		return err
	}
	if !fileUtils.IsFilenameOk(newDirname) {
		logger.Info("[-USER-]", invalidFileNameErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidFileNameErrorMessage,
		}
		return err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, dirUUID, author.UpdateDirAction)
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

	// Get the source directory metadata.
	dirMeta, err := m.dirRepo.GetDirMetadata(ctx, tx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	// Check if the name already exists in a current location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, dirMeta.ParentUUID, newDirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, dirMeta.ParentUUID, newDirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", newDirname)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return err
	}
	err = m.dirRepo.UpdateDirname(ctx, tx, newDirname, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateDirname failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return err
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
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
		return "", err
	}
	if !m.uuidTool.ValidateUUID(dstParentDirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
		return "", err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, dirUUID, author.CopyDirAction, author.RemoveDirAction)
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

	// Get the directory metadata.
	dirMeta, err := m.dirRepo.GetDirMetadata(ctx, tx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	// Check if the directory name already exists in a desired location in the db.
	var isExist bool
	isExist, err = m.fileRepo.IsFilenameExist(ctx, tx, dirMeta.ParentUUID, dirMeta.Dirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsFilenameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	isExist, err = m.dirRepo.IsDirNameExist(ctx, tx, dirMeta.ParentUUID, dirMeta.Dirname)
	if err != nil {
		logger.Errorf("[-INTERNAL-] IsDirNameExist failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
	}
	if isExist {
		logger.Infof("[-USER-] %s already exists in the desired location", dirMeta.Dirname)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: nameAlreadyExistErrorMessage,
		}
		return "", err
	}
	err = m.dirRepo.UpdateParentDirUUID(ctx, tx, dstParentDirUUID, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] UpdateParentDirUUID failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return "", err
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
	if !m.uuidTool.ValidateUUID(userUUID) {
		logger.Info("[-USER-]", invalidUserUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidUserUUIDErrorMessage,
		}
		return models.TmpFilePayload{}, err
	}
	if !m.uuidTool.ValidateUUID(dirUUID) {
		logger.Info("[-USER-]", invalidDirUUIDErrorMessage)
		err = models.XtremeError{
			Code:    models.BadInputErrorCode,
			Message: invalidDirUUIDErrorMessage,
		}
		return models.TmpFilePayload{}, err
	}
	// Authorization.
	isAuthorized, err := m.author.AuthorizeActionsOnDir(ctx, tx, userUUID, dirUUID, author.ViewDirAction)
	if err != nil {
		logger.Errorf("[-INTERNAL-] AuthorizeActionsOnDir failed with error %s", err.Error())
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

	// Get dir metadata.
	dirMeta, err := m.dirRepo.GetDirMetadata(ctx, tx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetDirMetadata failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.TmpFilePayload{}, err
	}

	// Get all child-files metadata.
	fileMetadataList, err := m.fileRepo.GetFileMetadataListByDir(ctx, tx, dirUUID)
	if err != nil {
		logger.Errorf("[-INTERNAL-] GetFileMetadataListByDir failed with error %s", err.Error())
		err = models.XtremeError{
			Code:    models.InternalServerErrorCode,
			Message: err.Error(),
		}
		return models.TmpFilePayload{}, err
	}

	inZipPaths := make(map[string]string)
	for _, childFileMeta := range fileMetadataList {
		relPath, err := filepath.Rel(dirMeta.Path, childFileMeta.Path)
		if err != nil {
			logger.Errorf("[-INTERNAL-] filepath.Rel failed with error %s", err.Error())
			err = models.XtremeError{
				Code:    models.InternalServerErrorCode,
				Message: err.Error(),
			}
			return models.TmpFilePayload{}, err
		}
		inZipPaths[relPath] = childFileMeta.RelPathOnDisk
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

// TODO: reimplement SoftRemoveDir
func (m *MultiOSFileManager) SoftRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
	panic("implement me")
}

// TODO: reimplement HardRemoveDir
func (m *MultiOSFileManager) HardRemoveDir(ctx context.Context, userUUID, dirUUID string) error {
	panic("implement me")
}
