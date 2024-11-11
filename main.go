package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

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

	// AWS config commands to set ACCESS_KEY_ID and SECRET_ACCESS_KEY
	execCommand("aws", "configure", "set", "aws_access_key_id", awsAccessKey).Run()
	execCommand("aws", "configure", "set", "aws_secret_access_key", awsSecretKey).Run()

	var Uploadcmd *exec.Cmd
	if fileType.IsDir() {
		if target != "" {
			Uploadcmd = execCommand("aws", "s3", "cp", source, "s3://"+awsBucket+"/"+target+"/"+newFolder, "--region", awsDefaultRegion, "--recursive")
			urls = "https://s3.console.aws.amazon.com/s3/buckets/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + target + "/" + newFolder + "/&showversions=false"
		} else {
			Uploadcmd = execCommand("aws", "s3", "cp", source, "s3://"+awsBucket+"/"+newFolder+"/", "--region", awsDefaultRegion, "--recursive")
			urls = "https://s3.console.aws.amazon.com/s3/buckets/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + newFolder + "/&showversions=false"
		}
	} else {
		if target != "" {
			Uploadcmd = execCommand("aws", "s3", "cp", source, "s3://"+awsBucket+"/"+target+"/", "--region", awsDefaultRegion)
			urls = "https://s3.console.aws.amazon.com/s3/object/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + target + "/" + newFolder
		} else {
			Uploadcmd = execCommand("aws", "s3", "cp", source, "s3://"+awsBucket+"/", "--region", awsDefaultRegion)
			urls = "https://s3.console.aws.amazon.com/s3/object/" + awsBucket + "?region=" + awsDefaultRegion + "&prefix=" + newFolder
		}
	}

	out, err := Uploadcmd.Output()
	if err != nil {
		return err
	}
	fmt.Printf("Output: %s\n", out)
	// End of S3 upload operation

	files := make([]File, 0)
	files = append(files, File{Name: artifactFilePath, URL: urls})

	return writeArtifactFile(files, artifactFilePath)
}
