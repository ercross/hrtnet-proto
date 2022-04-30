package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"github.com/Hrtnet/social-activities/internal/db"
	"github.com/pkg/errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

// unzipAndSave
// Returns save directory.
func unzipAndSave(source multipart.File, header *multipart.FileHeader, destination string) (*[]string, db.ErrorType, error) {

	// save file in temp dir
	// Custom dir name to avoid directory name collision if
	// multiple go routines invoke function at a time
	tempSavePath := fmt.Sprintf("*./temp-%s", header.Filename)
	err := os.MkdirAll(tempSavePath, 0700)
	if err != nil {
		return nil, db.InternalError, errors.Wrap(err, "error creating temp save path")
	}
	defer os.RemoveAll(tempSavePath) // clean up tempSavePath

	url, err := saveFile(source, header.Filename, tempSavePath)
	if err == errFileTooLarge {
		return nil, db.ValidationError, err
	}
	if err != nil {
		return nil, db.InternalError, errors.Wrap(err, "error: temporarily saving zipped file failed")
	}

	// open source zipped file
	reader, err := zip.OpenReader(url)
	if err != nil {
		return nil, db.ValidationError, errors.Wrap(err, "invalid zipped file")
	}
	defer reader.Close()

	// convert destination to absolute path
	destination, err = filepath.Abs(destination)
	if err != nil {
		return nil, db.InternalError, errors.Wrap(err, "error obtaining destination's absolute path: unzipAndSave")
	}

	// unpack each file inside zipped file to destination
	var savePaths []string
	for _, f := range reader.File {
		savedPath, err := unpackFile(f, destination)
		if err != nil {
			return nil, db.ValidationError, errors.Wrap(err, "unable to unpack file: invalid zip file")
		}
		savePaths = append(savePaths, savedPath)
	}

	return &savePaths, db.None, nil
}

// unpackFile unpacks a single zip file into destination
func unpackFile(file *zip.File, destination string) (string, error) {

	// check that file paths are not vulnerable to Zip Slip attack
	// described at https://snyk.io/research/zip-slip-vulnerability
	filePath := filepath.Join(destination, file.Name)
	if !strings.HasPrefix(filePath, filepath.Clean(destination)+string(os.PathSeparator)) {
		return "", fmt.Errorf("invalid file path: %s", filePath)
	}

	// create directory tree corresponding to file path into destination path
	if file.FileInfo().IsDir() {
		if err := os.MkdirAll(filePath, 0700); err != nil {
			return "", err
		}
		return "", nil
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0700); err != nil {
		return "", err
	}

	// create destination file for unzipped file content
	destinationFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return "", err
	}
	defer destinationFile.Close()

	// unzip the content of file and copy it to destination file
	zippedFile, err := file.Open()
	if err != nil {
		return "", err
	}
	defer zippedFile.Close()

	if _, err := io.Copy(destinationFile, zippedFile); err != nil {
		return "", err
	}
	savedPath := fmt.Sprintf("%s/%s", destination, file.Name)
	return savedPath, nil
}

// saveFile renames file to saveAs and saves file into savePath.
// If file is bigger than 5mb, errFileTooLarge is returned.
// Any other error returned is a server error.
// Note that saveAs must not end with a trailing slash.
// saveFile closes file before returning
func saveFile(file multipart.File, saveAs, savePath string) (fileUrl string, err error) {
	defer file.Close()
	buffer := &bytes.Buffer{}
	fileSize, err := buffer.ReadFrom(file)

	if err != nil {
		return "", errors.Wrap(err, "error reading file")
	}

	const fiveMB = 5 << 20
	if fileSize > fiveMB {
		return "", errFileTooLarge
	}

	fileUrl = savePath + "/" + saveAs
	out, err := os.Create(fileUrl)
	defer out.Close()
	if err != nil {

		return "", errors.Wrap(err, "error creating temp file")
	}

	_, err = io.Copy(out, buffer)
	if err != nil {
		return "", errors.Wrap(err, "error saving file")
	}
	return fileUrl, nil
}
