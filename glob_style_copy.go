package main

import (
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"log"
	"os"
	"path/filepath"
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

func NewS3GlobCopyConfig(c *cli.Context, copyBatchSize uint64) *S3GlobStyleCopy {
	source := c.String("source")

	return &S3GlobStyleCopy{
		AwsAccessKey:     c.String("aws-access-key"),
		AwsSecretKey:     c.String("aws-secret-key"),
		AwsDefaultRegion: c.String("aws-default-region"),
		AwsBucket:        c.String("aws-bucket"),
		Source:           source,
		TargetPath:       c.String("target-path"),
		NewFolder:        filepath.Base(source),
		ArtifactFilePath: c.String("artifact-file"),
		IncludeFilesGlob: c.String("include"),
		CopyBatchSize:    copyBatchSize,
	}
}

func (c *S3GlobStyleCopy) GetGlobArgsList() []string {

	patterns := strings.Split(c.IncludeFilesGlob, ",")
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
				err := c.uploadToS3(file)
				if err != nil {
					logrus.Printf("Failed to upload %s: %v\n", file, err)
				}
			}(sourceFile)
		}
		wg.Wait()
	}

	return nil
}

func (c *S3GlobStyleCopy) uploadToS3(sourceFile string) error {
	var s3Path string
	newFolder := c.NewFolder

	if c.TargetPath != "" {
		s3Path = fmt.Sprintf("s3://%s/%s/%s/%s", c.AwsBucket, c.TargetPath, newFolder, sourceFile)
	} else {
		s3Path = fmt.Sprintf("s3://%s/%s/%s", c.AwsBucket, newFolder, sourceFile)
	}

	absoluteSourcePath := filepath.Join(c.Source, sourceFile)
	fileType, err := os.Stat(absoluteSourcePath)
	if err != nil {
		log.Fatal(err)
	}
	cmd := CopyToS3(absoluteSourcePath, s3Path, c.AwsDefaultRegion, fileType.IsDir())
	output, err := cmd.CombinedOutput()
	if err != nil {
		logrus.Printf("Error copying %s to S3: %s\n", absoluteSourcePath, string(output))
		return err
	}

	return nil
}

const CopyBatchSize = 5

func GetCopyBatchSize() uint64 {
	return CopyBatchSize
}

//
