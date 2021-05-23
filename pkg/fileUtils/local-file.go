package fileUtils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const filenameInvalidChars string = "\\/.?%*:|\"<>,;= "

//const maxCopyFileResultChanSize = 100

// FileSaveReadCpRmer provides an interface to save/copy/read/remove files to/to/from/from a source.
type FileSaveReadCpRmer interface {
	// SaveFile saves content from a an io.Reader to a file in a location.
	SaveFile(relFilePathOD string, contentReader io.Reader) (int64, error)

	// ReadFile returns an instance of FileReadCloser. Data can be read from the instance via
	// Read() function. NOTE: Remember to Close() after reading the content.
	ReadFile(relFilePathOD string) (FileReadCloser, error)

	// GetFileReadCloserRmer returns an instance of FileReadCloseRmer.
	// FileReadCloseRmer allows reading content of the tmp file, closing tmp file,
	// and removing the tmp file after use.
	GetFileReadCloserRmer(absTmpFilePath string) (FileReadCloseRmer, error)

	//// InitAsyncCopyFileBatch
	//InitAsyncCopyFileBatch(ctx context.Context, copyJobChan chan CopyFileJob) chan CopyFileResult

	// RemoveFile removes a file from a source.
	RemoveFile(relFilePathOD string) error
}

type FileReadCloser interface {
	io.ReadCloser
	GetSize() (int64, error)
}

type File struct {
	*os.File
}

func (f *File) GetSize() (int64, error) {
	tmpFStat, err := f.File.Stat()
	if err != nil {
		return 0, err
	}
	return tmpFStat.Size(), nil
}

type FileReadCloseRmer interface {
	FileReadCloser
	Remove() error
}

type TmpFile struct {
	*File
}

func (tmp *TmpFile) Remove() error {
	return os.Remove(tmp.Name())
}

type LocalFileOperator struct {
	basePath string
}

// CreateNewLocalFileOperator create a new LocalFileOperator
func CreateNewLocalFileOperator(basePath string) LocalFileOperator {
	return LocalFileOperator{
		basePath,
	}
}

// SaveFile saves a file from a reader to the local disk.
// Return the number of bytes saved on the local disk.
// If the filename already exists, return error.
func (fs LocalFileOperator) SaveFile(relFilePathOD string, contentReader io.Reader) (int64, error) {
	// absolute filepath on disk.
	absFilePathOD := filepath.Join(fs.basePath, relFilePathOD)
	// Check if the file already exists.
	if _, err := os.Stat(absFilePathOD); err == nil {
		// Return error if the file already exists.
		return 0, fmt.Errorf("TmpFile %s already exist", absFilePathOD)
	}
	// Create a new empty dst file.
	dstF, err := os.Create(absFilePathOD)
	if err != nil {
		return 0, err
	}
	defer dstF.Close()
	// Copy content to the dst file.
	size, err := io.Copy(dstF, contentReader)
	return size, err
}

// ReadFile returns an os.File pointer with a given filename,
// which can be only used for reading the file content from the
// local storage.
func (fs LocalFileOperator) ReadFile(relFilePathOD string) (FileReadCloser, error) {
	absFilePathOD := filepath.Join(fs.basePath, relFilePathOD)
	file, err := os.Open(absFilePathOD)
	if err != nil {
		return nil, err
	}
	return &File{file}, nil
}

func (fs LocalFileOperator) GetFileReadCloserRmer(absTmpFilePath string) (FileReadCloseRmer, error) {
	tmpFile, err := os.Open(absTmpFilePath)
	if err != nil {
		return nil, err
	}
	return &TmpFile{&File{tmpFile}}, nil
}

//
//type CopyFileJob struct {
//	NewFileUUID    string
//	RelSrcFilePath string
//	RelDstFilePath string
//}
//
//type CopyFileResult struct {
//	CopyFileJob
//	err error
//}
//
//func (fs *LocalFileOperator) InitAsyncCopyFileBatch(ctx context.Context, copyJobChan chan CopyFileJob) chan CopyFileResult {
//	resultChan := make(chan CopyFileResult, maxCopyFileResultChanSize)
//	maxNWorkers := runtime.NumCPU() * 2
//	var wg sync.WaitGroup
//	for i := 0; i < maxNWorkers; i++ {
//		wg.Add(1)
//		go fs.copyFileWorker(i, ctx, copyJobChan, resultChan, wg.Done)
//	}
//	// init a goroutine to close the result channel
//	// when all copy jobs are done.
//	go func() {
//		wg.Wait()
//		close(resultChan)
//	}()
//	return resultChan
//}
//
//func (fs *LocalFileOperator) copyFileWorker(id int, ctx context.Context, jobs <-chan CopyFileJob, resultChan chan CopyFileResult, onExit func()) {
//	defer onExit()
//	for {
//		select {
//		case <-ctx.Done():
//			// If context is done (timeout, cancel, etc.)
//			// stop the goroutine.
//			return
//		case job, ok := <-jobs:
//			// Stop the goroutine if the channel is closed.
//			if !ok {
//				return
//			}
//			absSrcFilePath := filepath.Join(fs.basePath, job.RelSrcFilePath)
//			// Open source file.
//			srcFile, err := os.Open(absSrcFilePath)
//			if err != nil {
//				resultChan <- CopyFileResult{
//					CopyFileJob: job,
//					err:         err,
//				}
//				// do another job
//				continue
//			}
//			absDstFilePath := filepath.Join(fs.basePath, job.RelDstFilePath)
//			// Check if the file already exists.
//			if _, err := os.Stat(absDstFilePath); err == nil {
//				resultChan <- CopyFileResult{
//					CopyFileJob: job,
//					err:         err,
//				}
//				// do another job
//				continue
//			}
//			// Create a new empty dst file.
//			dstFile, err := os.Create(absDstFilePath)
//			if err != nil {
//				resultChan <- CopyFileResult{
//					CopyFileJob: job,
//					err:         err,
//				}
//				// do another job
//				continue
//			}
//			dstFile.Close()
//			// Copy content to the dst file.
//			_, err = io.Copy(dstFile, srcFile)
//			if err != nil {
//				resultChan <- CopyFileResult{
//					CopyFileJob: job,
//					err:         err,
//				}
//				// do another job
//				continue
//			}
//			resultChan <- CopyFileResult{
//				CopyFileJob: job,
//				err:         nil,
//			}
//		}
//	}
//}

// RemoveFile removes a file from the local disk.
func (fs LocalFileOperator) RemoveFile(relFilePathOD string) error {
	absFilePathOD := filepath.Join(fs.basePath, relFilePathOD)
	return os.Remove(absFilePathOD)
}

// IsFilenameOk checks if filename is valid. If the filename doesn't
// contain any prohibited characters, return true; otherwise
// return false
func IsFilenameOk(filename string) bool {
	if filename == "" || strings.ContainsAny(filename, filenameInvalidChars) {
		return false
	}
	return true
}
