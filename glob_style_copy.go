package main

import (
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"os"
	"os/exec"
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
}

func NewS3GlobCopyConfig(c *cli.Context) *S3GlobStyleCopy {
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
	}
}

func CopyFilesToS3WithGlobIncludes(c *cli.Context) error {

	copyConfig := NewS3GlobCopyConfig(c)
	globArgsList := copyConfig.GetGlobArgsList()

	fmt.Println("Glob args list: ", globArgsList)

	if globArgsList == nil {
		return errors.New("Invalid glob pattern")
	}
	if len(globArgsList) < 1 {
		return errors.New("No files found")
	}

	for _, pattern := range globArgsList {
		err := copyConfig.CopyFilesToS3(pattern)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *S3GlobStyleCopy) GetGlobArgsList() []string {
	patterns := strings.Split(c.IncludeFilesGlob, ",")
	for i, pattern := range patterns {
		patterns[i] = strings.TrimSpace(pattern)
	}
	return patterns
}

func (c *S3GlobStyleCopy) CopyFilesToS3(pattern string) error {
	logrus.Println("Copying files to S3 for pattern ", pattern)

	sourceDir := os.DirFS(c.Source)

	matchedFiles, err := doublestar.Glob(sourceDir, pattern)
	if err != nil {
		fmt.Println("matchedDirs Error: ", err.Error())
		return errors.New("Failed to match files")
	}

	return c.CopyFiles(matchedFiles, 5)
}

func (c *S3GlobStyleCopy) CopyFiles(sourceFilesList []string, batchSize uint64) error {

	numFiles := len(sourceFilesList)

	if numFiles == 0 {
		return fmt.Errorf("no files to copy")
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
					fmt.Printf("Failed to upload %s: %v\n", file, err)
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
		s3Path = fmt.Sprintf("s3://%s/%s/%s/%s", c.AwsBucket, c.TargetPath, newFolder, filepath.Base(sourceFile))
	} else {
		s3Path = fmt.Sprintf("s3://%s/%s/%s", c.AwsBucket, newFolder, filepath.Base(sourceFile))
	}

	absoluteSourceFilePath := filepath.Join(c.Source, sourceFile)
	argsList := []string{
		"s3", "cp", absoluteSourceFilePath, s3Path,
		"--region", c.AwsDefaultRegion,
	}

	fmt.Println("Executing command: aws", argsList)
	fmt.Println("aws ", argsList)

	// Execute the command
	cmd := exec.Command("aws", argsList...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Error copying %s to S3: %s\n", absoluteSourceFilePath, string(output))
		return err
	}

	fmt.Printf("Successfully uploaded %s to %s\n", absoluteSourceFilePath, s3Path)
	return nil
}
