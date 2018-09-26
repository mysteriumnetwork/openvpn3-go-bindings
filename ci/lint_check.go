// +build mage

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
	"github.com/mysteriumnetwork/go-openvpn/ci/util"
)

// Checks if golint exists, if not installs it
func GetLint() error {
	path, _ := util.GetGoBinaryPath("golint")
	if path != "" {
		fmt.Println("Tool 'golint' already installed")
		return nil
	}
	err := sh.RunV("go", "get", "-u", "golang.org/x/lint/golint")
	if err != nil {
		fmt.Println("Could not go get golint")
		return err
	}
	return nil
}

var packageRegexp = regexp.MustCompile(`\.\./(.*)\/.*\.go`)

func getPackageFromGoLintOutput(line string) string {
	results := packageRegexp.FindAllStringSubmatch(line, -1)
	for i := range results {
		return results[i][1]
	}
	return ""
}

func beautiflyPrintGoLintOutput(rawGolint string) {
	packageErrorMap := make(map[string][]string, 0)
	separateLines := strings.Split(rawGolint, "\n")

	for i := range separateLines {
		pkg := getPackageFromGoLintOutput(separateLines[i])
		if val, ok := packageErrorMap[pkg]; ok {
			packageErrorMap[pkg] = append(val, separateLines[i])
		} else {
			lines := []string{separateLines[i]}
			packageErrorMap[pkg] = lines
		}
	}

	fmt.Println()
	for k := range packageErrorMap {
		fmt.Println("PACKAGE: ", k)
		fmt.Println()
		for _, v := range packageErrorMap[k] {
			fmt.Println(v)
		}
		fmt.Println()
	}
}

func GoLint() error {
	mg.Deps(GetLint)
	path, err := util.GetGoBinaryPath("golint")
	if err != nil {
		return err
	}
	var files []string
	var excludedDirs = []string{".git", "vendor"}
	err = filepath.Walk("../", func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			for _, exclude := range excludedDirs {
				if strings.Contains(path, "/"+exclude) {
					return nil
				}
			}
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return err
	}
	args := []string{"--set_exit_status", "--min_confidence=1"}
	args = append(args, files...)
	output, err := sh.Output(path, args...)
	exitStatus := sh.ExitStatus(err)
	if exitStatus == 0 {
		fmt.Println("No linting errors")
		return nil
	}

	beautiflyPrintGoLintOutput(output)
	fmt.Println("Linting failed!")
	return err
}
