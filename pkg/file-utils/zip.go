package fileUtils

import (
	"archive/zip"
	"bufio"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

type FileCompressor interface {
	CompressFiles(inZipPaths map[string]string) (*os.File, error)
}

type FileZipper struct {
	basePath string
}

// CreateNewFileZipper create a new FileZipper.
func CreateNewFileZipper(basePath, tmpFilePath string) *FileZipper {
	return &FileZipper{
		basePath,
	}
}

func (z *FileZipper) CompressFiles(inZipPaths map[string]string) (*os.File, error) {
	tmpfile, err := ioutil.TempFile("", "compress_temp_*")
	if err != nil {
		log.Fatal(err)
	}
	// Create a buffer to write a file.
	bufFWriter := bufio.NewWriter(tmpfile)
	defer bufFWriter.Flush()
	// Create a new zip writer (which writes to temp file)
	zipWriter := zip.NewWriter(bufFWriter)
	defer zipWriter.Close()
	for pathInZip, relFilePathOD := range inZipPaths {
		// Create a new path inside the zip file
		f, err := zipWriter.Create(pathInZip)
		if err != nil {
			return nil, err
		}
		absFilePath := filepath.Join(z.basePath, relFilePathOD)
		// Read content of file which needs to be zipped.
		fileToZip, err := os.Open(absFilePath)
		if err != nil {
			return nil, err
		}
		// Copy the content from real file to the file(in zip).
		_, err = io.Copy(f, fileToZip)
		if err != nil {
			return nil, err
		}
		// Close the real file.
		fileToZip.Close()
	}
	return tmpfile, nil
}
