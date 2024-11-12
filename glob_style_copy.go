package main

import (
	"errors"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

type S3GlobStyleCopy struct {
	AwsAccessKey     string
	AwsSecretKey     string
	AwsDefaultRegion string
	AwsBucket        string
	Source           string
	TargetPath       string
	NewFolder        string
	ArtifactFilePath string
	IncludeFilesGlob string
	CopyBatchSize    uint64
}

func GetGlobArgsList(includeFilesGlob string) []string {

	patterns := strings.Split(includeFilesGlob, ",")
	for i, pattern := range patterns {
		patterns[i] = strings.TrimSpace(pattern)
	}
	return patterns
}

func (c *S3GlobStyleCopy) GetMatchedFiles(pattern string) ([]string, error) {

	sourceDir := os.DirFS(c.Source)

	matchedFiles, err := doublestar.Glob(sourceDir, pattern)
	if err != nil {
		logrus.Println("matchedDirs Error: ", err.Error())
		return []string{}, errors.New("Failed to match files")
	}
	return matchedFiles, nil
}

func (c *S3GlobStyleCopy) CopyFiles(sourceFilesList []string, batchSize uint64) error {

	numFiles := len(sourceFilesList)

	if numFiles == 0 {
		return errors.New("No files to copy")
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
				urls, err := CopyToS3(file, c.TargetPath, c.NewFolder, c.AwsBucket, c.AwsDefaultRegion, false)
				if err != nil {
					logrus.Printf("Failed to upload %s: %v\n", file, err)
				}
				_ = urls
			}(sourceFile)
		}
		wg.Wait()
	}

	return nil
}

const CopyBatchSize = 5

func GetCopyBatchSize() uint64 {
	return CopyBatchSize
}

//
