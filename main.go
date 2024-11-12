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

	var urls string

	if strings.ContainsAny(source, "*") {
		log.Fatal("Glob pattern not allowed!")
	}

	// AWS config commands to set ACCESS_KEY_ID and SECRET_ACCESS_KEY
	execCommand("aws", "configure", "set", "aws_access_key_id", awsAccessKey).Run()
	execCommand("aws", "configure", "set", "aws_secret_access_key", awsSecretKey).Run()

	if includeFilesGlobStr != "" {
		err := CopyFilesToS3WithGlobIncludes(c)
		if err != nil {
			log.Println("Error copying files to S3: ", err.Error())
			return err
		}
		log.Println("All Files copied to S3 successfully!")
		return nil
	}

	fileType, err := os.Stat(source)
	if err != nil {
		log.Fatal(err)
	}

	var Uploadcmd *exec.Cmd

	prefixPath := awsBucket
	if target != "" {
		prefixPath += "/" + target
	}

	s3Path := "s3://" + prefixPath
	s3Path += "/" + newFolder

	if fileType.IsDir() {
		urls = baseURL + "buckets/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + prefixPath + "/" + newFolder + "/&showversions=false"
	} else {
		urls = baseURL + "object/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + prefixPath + "/" + newFolder
	}

	Uploadcmd = CopyToS3(source, s3Path, awsDefaultRegion, fileType.IsDir())

	out, err := Uploadcmd.Output()
	if err != nil {
		fmt.Println("Error uploading to S3: ", err.Error())
		return err
	}
	fmt.Printf("Output: %s\n", out)
	// End of S3 upload operation

	files := make([]File, 0)
	files = append(files, File{Name: artifactFilePath, URL: urls})

	return writeArtifactFile(files, artifactFilePath)
}

func CopyToS3(source, s3Path, awsDefaultRegion string, isDir bool) *exec.Cmd {
	cliArgs := []string{"s3", "cp", source, s3Path, "--region", awsDefaultRegion}
	if isDir {
		cliArgs = append(cliArgs, "--recursive")
	}

	uploadCmd := execCommand("aws", cliArgs...)
	return uploadCmd
}

func CopyFilesToS3WithGlobIncludes(c *cli.Context) error {

	var allMatchedFiles []string

	copyConfig := NewS3GlobCopyConfig(c, GetCopyBatchSize())
	globArgsList := copyConfig.GetGlobArgsList()

	if globArgsList == nil {
		return errors.New("Invalid glob pattern")
	}
	if len(globArgsList) < 1 {
		return errors.New("No files found")
	}

	for _, pattern := range globArgsList {
		tmpFilesList, err := copyConfig.GetMatchedFiles(pattern)
		if err != nil {
			return err
		}

		allMatchedFiles = append(allMatchedFiles, tmpFilesList...)
	}

	return copyConfig.CopyFiles(allMatchedFiles, 5)
}

const baseURL = "https://s3.console.aws.amazon.com/s3/"

//
//
