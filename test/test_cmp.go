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
	checkJakarta()
	checkApacheCvs()
	checkAllTestsDir()
}

func checkTest(pattern string) {
	fmt.Println("\n\n")
	baseDirs := []string{
		"/tmp/test-glob",
	}
	args := Args{
		TargetDir: baseDirs[0],
		Filter:    pattern,
		Excludes:  "",
	}

	fmt.Println("================= double star =================")
	doubleStar(baseDirs[0], pattern)
	fmt.Println("================= ApplyFilter1 =================")
	ApplyFilter0(args)
	fmt.Println("================= ApplyFilter2 =================")
	ApplyFilter2(baseDirs[0], pattern, "")
}

func checkCVS() {
	checkTest("**/CVS/*")
}

func checkJakarta() {
	checkTest("org/apache/jakarta/**")
}

func checkApacheCvs() {
	checkTest("org/apache/**/CVS/*")
}

func checkAllTestsDir() {
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

func ApplyFilter0(args Args) ([]FileInfo, error) {
	var files []FileInfo
	m := antpathmatcher.NewAntPathMatcher()

	if args.TargetDir == "" {
		args.TargetDir = "."
	}

	err := filepath.WalkDir(args.TargetDir, func(path string, d os.DirEntry, e error) error {

		if m.Match(args.Filter, path) {
			if m.Match(args.Excludes, path) {
				fmt.Println("eeee. ", path)
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

func ApplyFilter2(targetDir, filter, excludes string) ([]FileInfo, error) {
	var files []FileInfo
	m := antpathmatcher.NewAntPathMatcher()

	if targetDir == "" {
		targetDir = "."
	}

	err := filepath.WalkDir(
		targetDir,
		func(path string, d os.DirEntry, e error) error {
			//fmt.Println("path ", path, " filter ", filter)

			relativePath, _ := filepath.Rel(targetDir, path)
			// relativePath = path

			if m.Match(filter, relativePath) {
				fmt.Println("wwww. ", relativePath)
				if m.Match(excludes, path) {
					fmt.Printf("path %s match exclude criteria %s", path, excludes)
				} else {
					file, err := getFileInfo(path)
					if err != nil {
						fmt.Printf("error to get file info of path %s", path)
						return errors.New("Bad") // logError(logger, fmt.Sprintf("error to get file info of path %s", path), err)
					}
					files = append(files, file)
				}
			} else {
				//fmt.Println("not matched ", path, " filter ", filter)
			}
			return nil
		})

	if err != nil {
		return []FileInfo{}, err
	}
	fmt.Println("====================")
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

func ApplyFilter1(targetDir, filter, excludes string) ([]FileInfo, error) {
	var files []FileInfo
	m := antpathmatcher.NewAntPathMatcher()

	if targetDir == "" {
		targetDir = "."
	}

	err := filepath.WalkDir(
		targetDir,
		func(path string, d os.DirEntry, e error) error {
			fmt.Println("path ", path, " filter ", filter)
			if m.Match(filter, path) {
				fmt.Println("***************0 matched ", path, " filter ", filter)
				if m.Match(excludes, path) {
					fmt.Printf("path %s match exclude criteria %s", path, excludes)
				} else {
					file, err := getFileInfo(path)
					if err != nil {
						fmt.Printf("error to get file info of path %s", path)
						return errors.New("Bad")
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
