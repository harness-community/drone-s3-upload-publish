package main

import (
	"errors"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sirupsen/logrus"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

func GetGlobArgsList(includeFilesGlob string) []string {

	patterns := strings.Split(includeFilesGlob, ",")
	for i, pattern := range patterns {
		patterns[i] = strings.TrimSpace(pattern)
	}
	return patterns
}

func GetMatchedFiles(source, pattern string) ([]string, error) {

	sourceDir := os.DirFS(source)
	matchedFiles, err := doublestar.Glob(sourceDir, pattern)
	if err != nil {
		logrus.Println("matchedDirs Error: ", err.Error())
		return []string{}, errors.New("Failed to match files")
	}
	return matchedFiles, nil
}

func BatchCopyFiles(source string, sourceFilesList []string, targetPath, s3Bucket, defaultRegion string,
	batchSize uint64) ([]string, error) {

	urlsList := []string{}

	numFiles := len(sourceFilesList)
	if numFiles == 0 {
		return urlsList, errors.New("No files to copy")
	}

	for i := 0; i < numFiles; i += int(batchSize) {
		end := i + int(batchSize)
		if end > numFiles {
			end = numFiles
		}

		batch := sourceFilesList[i:end]
		var wg sync.WaitGroup

		for _, sourceFile := range batch {
			wg.Add(1)
			go func(file string) {
				defer wg.Done()
				prefixedSrcPath := source + "/" + file
				dstPath := filepath.Base(prefixedSrcPath)
				topLevel := filepath.Base(source)
				dstPath = topLevel + "/" + replacePrefix(prefixedSrcPath, source, "")
				urls, err := CopyToS3(prefixedSrcPath, targetPath, dstPath, s3Bucket, defaultRegion)
				if err != nil {
					logrus.Printf("Failed to upload %s: %v\n", file, err)
				}
				urlsList = append(urlsList, urls)
			}(sourceFile)
		}
		wg.Wait()
	}

	return urlsList, nil
}

func replacePrefix(path, prefix, newPrefix string) string {
	prefix += "/"
	if strings.HasPrefix(path, prefix) {
		return newPrefix + strings.TrimPrefix(path, prefix)
	}
	return path
}

const CopyBatchSize = 5

func GetCopyBatchSize() uint64 {
	return CopyBatchSize
}

//
