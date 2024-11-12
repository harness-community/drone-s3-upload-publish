package main

import (
	"errors"
	"fmt"
	"github.com/bmatcuk/doublestar/v4"
	"github.com/georgeJobs/go-antpathmatcher"
	"os"
	"path/filepath"
	"time"
)

func main() {
	checkCVS()
	fmt.Println("\n\n")
	checkJakarta()
	fmt.Println("\n\n")
	checkApacheCvs()
	fmt.Println("\n\n")
	checkAllTestsDir()
}

func checkTest(pattern string) {
	baseDirs := []string{
		"./test-glob",
		//"/tmp/test-glob",
	}
	args := Args{
		TargetDir: baseDirs[0],
		Filter:    pattern,
		Excludes:  "",
	}

	fmt.Println("================= double star =================")
	doubleStar(baseDirs[0], pattern)
	fmt.Println("================= ApplyFilterFindFileDefault =================")
	ApplyFilterFindFileDefault(args)
	//fmt.Println("================= ApplyFilterFindFileDefaultFixed =================")
	//ApplyFilterFindFileDefaultFixed(baseDirs[0], pattern, "")
}

func checkCVS() {
	fmt.Println("********************* CVS test ***********************")
	checkTest("**/CVS/*")
}

func checkJakarta() {
	fmt.Println("********************* Jakarta test ***********************")

	checkTest("org/apache/jakarta/**")
}

func checkApacheCvs() {
	fmt.Println("********************* Apache CVS test ***********************")

	checkTest("org/apache/**/CVS/*")
}

func checkAllTestsDir() {
	fmt.Println("********************* All test Dirs ***********************")
	checkTest("**/test/**")
}

func doubleStar(sourceDirStr, pattern string) {

	sourceDir := os.DirFS(sourceDirStr)
	matchedFiles, err := doublestar.Glob(sourceDir, pattern)
	if err != nil {
		fmt.Println("matchedDirs Error: ", err.Error())
	}
	for _, match := range matchedFiles {
		fmt.Println("qqqq> ", match)
	}
}

type FileInfo struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IsDirectory  bool   `json:"isDirectory"`
	Length       int64  `json:"length"`
	LastModified string `json:"lastModified"`
}

type Args struct {
	TargetDir string
	Filter    string
	Excludes  string
}

func ApplyFilterFindFileDefault(args Args) ([]FileInfo, error) {
	var files []FileInfo
	m := antpathmatcher.NewAntPathMatcher()

	if args.TargetDir == "" {
		args.TargetDir = "."
	}

	err := filepath.WalkDir(args.TargetDir, func(path string, d os.DirEntry, e error) error {
		//fmt.Println("args.Filter ", args.Filter, " path ", path)
		if m.Match(args.Filter, path) {
			fmt.Println("eeee. ", path)
			if m.Match(args.Excludes, path) {
				//fmt.Printf("path %s match exclude criteria %s", path, args.Excludes)
			} else {
				file, err := getFileInfo(path)
				if err != nil {
					return errors.New("Bad") //
				}

				files = append(files, file)
			}
		}

		return nil
	})
	if err != nil {
		return []FileInfo{}, err
	}

	return files, nil
}

func getFileInfo(path string) (FileInfo, error) {
	fi, err := os.Lstat(path)

	if err != nil {
		return FileInfo{}, err
	}
	return FileInfo{
		Name:         fi.Name(),
		Path:         path,
		IsDirectory:  fi.IsDir(),
		Length:       fi.Size(),
		LastModified: fi.ModTime().Format(time.RFC3339),
	}, nil
}
