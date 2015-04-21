package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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
// 1. scan Godeps.json
// 2. go get package
// 3. copy package to Godeps/_workspace/
// 4. update Godeps.json
// TODO get rid of the annoying error handlings
func update(pkg string) {
	var err error
	pwd, err := os.Getwd()
	if err != nil {
		errored(err)
	}

	godepFile, err := os.Open(pwd + "/Godeps/Godeps.json")
	if err != nil {
		fmt.Println("No Godeps found in current directory")
		os.Exit(1)
	}

	err = getPackage(pkg)
	if err != nil {
		errored(err)
	}

	godep, err := parseGodep(godepFile)
	if err != nil {
		errored(err)
	}

	rev, err := getPackageRev(pkg)
	if err != nil {
		errored(err)
	}

	for idx, dep := range godep.Deps {
		fmt.Println(dep.ImportPath)
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

	err = importPackage(pkg, pwd)
	if err != nil {
		errored(err)
	}

	err = updateGoDeps(pkg, rev, godepFile)
	if err != nil {
		errored(err)
	}
	fmt.Println("updated", pkg)
}

func parseGodep(godepFile *os.File) (godep GoDep, err error) {
	godepJson, err := ioutil.ReadAll(godepFile)
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

//TODO get the git sha of the package from $GOPATH
func getPackageRev(pkg string) (rev string, err error) {
	return
}

//TODO copy package from $GOPATH to Godeps/_workspace
func importPackage(pkg string, rootPath string) (err error) {
	return
}

//TODO update Godeps.json
func updateGoDeps(pkg string, rev string, godepFile *os.File) (err error) {
	return
}
