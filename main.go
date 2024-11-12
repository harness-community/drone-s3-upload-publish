package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	pluginVersion = "1.0.0"
)

func main() {
	app := cli.NewApp()
	app.Name = "drone-s3-upload-publish"
	app.Usage = "Drone plugin to upload file/directories to AWS S3 Bucket and display the bucket url under 'Executions > Artifacts' tab"
	app.Action = run
	app.Version = pluginVersion
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "aws-access-key",
			Usage:  "AWS Access Key ID",
			EnvVar: "PLUGIN_AWS_ACCESS_KEY_ID",
		},
		cli.StringFlag{
			Name:   "aws-secret-key",
			Usage:  "AWS Secret Access Key",
			EnvVar: "PLUGIN_AWS_SECRET_ACCESS_KEY",
		},
		cli.StringFlag{
			Name:   "aws-default-region",
			Usage:  "AWS Default Region",
			EnvVar: "PLUGIN_AWS_DEFAULT_REGION",
		},
		cli.StringFlag{
			Name:   "aws-bucket",
			Usage:  "AWS S3 Bucket",
			EnvVar: "PLUGIN_AWS_BUCKET",
		},
		cli.StringFlag{
			Name:   "source",
			Usage:  "Source",
			EnvVar: "PLUGIN_SOURCE",
		},
		cli.StringFlag{
			Name:   "target-path",
			Usage:  "target",
			EnvVar: "PLUGIN_TARGET",
		},
		cli.StringFlag{
			Name:   "artifact-file",
			Usage:  "Artifact file",
			EnvVar: "PLUGIN_ARTIFACT_FILE",
		},
		cli.StringFlag{
			Name:   "include",
			Usage:  "Include file patterns int ant style glob style",
			EnvVar: "PLUGIN_INCLUDE",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

var execCommand = exec.Command

func run(c *cli.Context) error {
	awsAccessKey := c.String("aws-access-key")
	awsSecretKey := c.String("aws-secret-key")
	awsDefaultRegion := c.String("aws-default-region")
	awsBucket := c.String("aws-bucket")
	source := c.String("source")
	target := c.String("target-path")
	newFolder := filepath.Base(source)
	artifactFilePath := c.String("artifact-file")
	includeFilesGlobStr := c.String("include")

	if strings.ContainsAny(source, "*") {
		log.Fatal("Glob pattern not allowed!")
	}

	// AWS config commands to set ACCESS_KEY_ID and SECRET_ACCESS_KEY
	execCommand("aws", "configure", "set", "aws_access_key_id", awsAccessKey).Run()
	execCommand("aws", "configure", "set", "aws_secret_access_key", awsSecretKey).Run()

	urls := ""
	urlsList := []string{}
	var err error
	var urlArtifactFiles []File

	switch {
	case includeFilesGlobStr != "": // Glob copy
		urlsList, err = CopyFilesToS3WithGlobIncludes(awsDefaultRegion, awsBucket, source, target, includeFilesGlobStr)
		if err != nil {
			log.Println("Error copying files to S3: ", err.Error())
			return err
		}
		for _, url := range urlsList {
			urlArtifactFiles = append(urlArtifactFiles, File{Name: artifactFilePath, URL: url})
		}

	default: // Single file or directory copy
		urls, err = CopyToS3(source, target, newFolder, awsBucket, awsDefaultRegion)
		if err != nil {
			log.Println("Error copying files to S3: ", err.Error())
			return err
		}
		urlArtifactFiles = append(urlArtifactFiles, File{Name: artifactFilePath, URL: urls})
	}

	return writeArtifactFile(urlArtifactFiles, artifactFilePath)
}

func CopyToS3(source, target, newFolder, awsBucket, awsDefaultRegion string) (string, error) {

	fileType, err := os.Stat(source)
	if err != nil {
		log.Fatal(err)
	}
	isDir := fileType.IsDir()

	s3Path, _, urls := GetPathsAndURLs(target, newFolder, awsBucket, awsDefaultRegion, isDir)

	UploadCmd := RunS3CliCopyCmd(source, s3Path, awsDefaultRegion, isDir)

	out, err := UploadCmd.Output()
	if err != nil {
		fmt.Println("Error uploading to S3: ", err.Error())
		return urls, err
	}
	fmt.Printf("Output: %s\n", out)
	// End of S3 upload operation
	return urls, nil
}

func GetPathsAndURLs(target, newFolder, awsBucket, awsDefaultRegion string, isDir bool) (string, string, string) {
	urls := ""
	prefixPath := awsBucket
	if target != "" {
		prefixPath += "/" + target
	}

	s3Path := "s3://" + prefixPath
	s3Path += "/" + newFolder

	if isDir {
		urls = baseURL + "buckets/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + prefixPath + "/" + newFolder + "/&showversions=false"
	} else {
		urls = baseURL + "object/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + prefixPath + "/" + newFolder
	}
	return s3Path, prefixPath, urls
}

func RunS3CliCopyCmd(source, s3Path, awsDefaultRegion string, isDir bool) *exec.Cmd {
	cliArgs := []string{"s3", "cp", source, s3Path, "--region", awsDefaultRegion}
	if isDir {
		cliArgs = append(cliArgs, "--recursive")
	}

	fmt.Println("aws ", strings.Join(cliArgs, " "))
	uploadCmd := execCommand("aws", cliArgs...)
	return uploadCmd
}

func CopyFilesToS3WithGlobIncludes(defaultRegion, s3Bucket, source, targetPath,
	includesGlob string) ([]string, error) {

	var allMatchedFiles []string

	globArgsList := GetGlobArgsList(includesGlob)

	if globArgsList == nil {
		return []string{}, errors.New("Invalid glob pattern")
	}
	if len(globArgsList) < 1 {
		return []string{}, errors.New("No files found")
	}

	for _, pattern := range globArgsList {
		tmpFilesList, err := GetMatchedFiles(source, pattern)
		if err != nil {
			return []string{}, err
		}
		allMatchedFiles = append(allMatchedFiles, tmpFilesList...)
	}

	return BatchCopyFiles(source, allMatchedFiles, targetPath, s3Bucket, defaultRegion, GetCopyBatchSize())
}

const baseURL = "https://s3.console.aws.amazon.com/s3/"

//
//
