package main

import (
	"flag"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli"
	"os"
	"os/exec"
	"testing"
)

var capturedArgs []string

func mockExecCommand(command string, args ...string) *exec.Cmd {
	capturedArgs = append([]string{command}, args...)
	cs := []string{"-test.run=TestHelperProcess", "--", command}
	cs = append(cs, args...)
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	return cmd
}

func TestHelperProcess(*testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}
	os.Exit(0)
}

func TestRun_GeneratesCorrectArgs(t *testing.T) {

	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	defer func() { capturedArgs = []string{} }()

	app := cli.NewApp()
	app.Action = run
	set := flag.NewFlagSet("test", 0)
	set.String("aws-access-key", "mock-access-key", "AWS Access Key ID")
	set.String("aws-secret-key", "mock-secret-key", "AWS Secret Access Key")
	set.String("aws-default-region", "ap-south-1", "AWS Default Region")
	set.String("aws-bucket", "bfw-hns-test-bucket", "AWS S3 Bucket")
	set.String("source", "./test", "Source")
	set.String("target-path", "test-target", "Target path")
	set.String("artifact-file", "artifact.txt", "Artifact file")
	set.String("glob", "", "Include patterns")

	context := cli.NewContext(app, set, nil)

	err := run(context)
	assert.NoError(t, err)

	expectedArgs := []string{
		"aws", "s3", "cp", "./test",
		"s3://bfw-hns-test-bucket/test-target/test",
		"--region", "ap-south-1", "--recursive",
	}

	assert.Equal(t, expectedArgs, capturedArgs)
}

func TestCopyFilesToS3WithGlobIncludes(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	defer func() { capturedArgs = []string{} }()

	globArgsList := GetGlobArgsList("**/*.html, **/*.css")

	var allMatchedFiles []string

	for _, pattern := range globArgsList {
		tmpFilesList, err := GetMatchedFiles("./test", pattern)
		if err != nil {
			t.Error("Failed to get matched files")
		}
		allMatchedFiles = append(allMatchedFiles, tmpFilesList...)
	}

	expectedFilesMap := map[string]bool{
		"s3-copy-test-files/project_root/level1/level2/styles/contact.css": true,
		"s3-copy-test-files/project_root/level1/styles/about.css":          true,
		"s3-copy-test-files/project_root/styles/style.css":                 true,
		"s3-copy-test-files/project_root/index.html":                       true,
		"s3-copy-test-files/project_root/level1/about.html":                true,
		"s3-copy-test-files/project_root/level1/level2/contact.html":       true,
	}

	gotFilesMap := make(map[string]bool)

	for _, file := range allMatchedFiles {
		gotFilesMap[file] = true
	}

	for key := range expectedFilesMap {
		assert.True(t, gotFilesMap[key])
	}
}

func TestExecArgs(t *testing.T) {
	execCommand = mockExecCommand
	defer func() { execCommand = exec.Command }()
	defer func() { capturedArgs = []string{} }()

	app := cli.NewApp()
	app.Action = run
	set := flag.NewFlagSet("test", 0)
	set.String("aws-access-key", "mock-access-key", "AWS Access Key ID")
	set.String("aws-secret-key", "mock-secret-key", "AWS Secret Access Key")
	set.String("aws-default-region", "ap-south-1", "AWS Default Region")
	set.String("aws-bucket", "bfw-hns-test-bucket", "AWS S3 Bucket")
	set.String("source", "./test", "Source")
	set.String("target-path", "test-target", "Target path")
	set.String("artifact-file", "artifact.txt", "Artifact file")
	set.String("glob", "**/*.xhtml", "Include patterns")

	context := cli.NewContext(app, set, nil)

	err := run(context)
	assert.NoError(t, err)

	expectedArgs := []string{"aws", "s3", "cp", "./test/s3-copy-test-files/project_root/level1/level2/contact.xhtml",
		"s3://bfw-hns-test-bucket/test-target/test/s3-copy-test-files/project_root/level1/level2/contact.xhtml",
		"--region", "ap-south-1"}
	assert.Equal(t, expectedArgs, capturedArgs)

}
