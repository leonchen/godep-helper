package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Dep struct {
	ImportPath string
	Rev        string
	Comment    string
}

type GoDep struct {
	ImportPath string
	GoVersion  string
	Packages   []string
	Deps       []*Dep
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: %s update [package]\n", os.Args[0])
	flag.PrintDefaults()
	os.Exit(2)
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if len(os.Args) != 3 {
		usage()
	}

	if os.Args[1] == "update" {
		if pkg := os.Args[2]; len(pkg) > 0 {
			update(pkg)
		} else {
			usage()
		}
	}
}

// update:
// 1. scan Godeps/Godeps.json
// 2. go get package
// 3. copy package from GOPATH to Godeps/_workspace/
// 4. update Godeps/Godeps.json
func update(pkg string) {
	var err error
	pwd, _ := os.Getwd()

	godepFile := pwd + "/Godeps/Godeps.json"
	godep, err := parseGodep(godepFile)
	if err != nil {
		fmt.Println("No Godeps found in current directory")
		os.Exit(1)
	}

	err = getPackage(pkg)
	if err != nil {
		errored(err)
	}

	// looking for the downloaded package from the first dir in GOPATH
	pkgPath := strings.Split(os.Getenv("GOPATH"), ":")[0] + "/src/" + pkg
	rev, err := getPackageRev(pkgPath)
	if err != nil {
		errored(err)
	}

	// verify the target package definition in Godeps.json
	for idx, dep := range godep.Deps {
		if dep.ImportPath == pkg {
			if dep.Rev == rev {
				fmt.Println("No need to update Godep package", pkg)
				os.Exit(0)
			} else {
				godep.Deps = append(godep.Deps[:idx], godep.Deps[idx+1:]...)
				break
			}
		}
	}

	// copy the downloaded package to Godeps..
	err = importPackage(pkg, pkgPath, pwd)
	// ..and update Godeps.json
	if err == nil {
		newPkg := &Dep{
			ImportPath: pkg,
			Rev:        rev,
		}

		godep.Deps = append(godep.Deps, newPkg)
		updated, _ := json.MarshalIndent(godep, "", "\t")
		err = ioutil.WriteFile(godepFile, updated, os.ModePerm)
	}
	if err != nil {
		errored(err)
	}

	fmt.Println("updated", pkg)
}

func parseGodep(godepFile string) (godep GoDep, err error) {
	godepJson, err := ioutil.ReadFile(godepFile)
	if err != nil {
		return
	}

	err = json.Unmarshal(godepJson, &godep)
	return
}

func errored(err error) {
	fmt.Println("Error:", err)
	os.Exit(1)
}

func getPackage(pkg string) (err error) {
	fmt.Println("getting", pkg)
	cmd := exec.Command("go", "get", "-u", pkg)
	std, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(std))
	}
	return
}

func getPackageRev(pkgPath string) (rev string, err error) {
	os.Chdir(pkgPath)
	cmd := exec.Command("git", "show-ref", "--hash", "HEAD")
	std, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(errors.New("cannot get HEAD sha in " + pkgPath))
	} else {
		rev = strings.TrimSpace(string(std))
	}
	return
}

func importPackage(pkg string, pkgPath string, rootPath string) (err error) {
	godepPkgPath := rootPath + "/Godeps/_workspace/" + "src/" + pkg
	clean := exec.Command("rm", "-rf", godepPkgPath)
	err = clean.Run()
	if err == nil {
		exec.Command("mkdir", "-p", godepPkgPath).Run()
		paths := strings.Split(godepPkgPath, "/")
		targetParentPath := strings.Join(paths[:len(paths)-1], "/")
		cp := exec.Command("cp", "-rf", pkgPath, targetParentPath)
		std, err := cp.CombinedOutput()
		if err != nil {
			fmt.Println(string(std))
			errored(err)
		}
	}
	return
}
