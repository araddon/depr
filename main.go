package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"os/exec"
)

var (
	config string
	gopath string
)

func init() {
	flag.StringVar(&config, "config", ".depr.yaml", "Yaml config file with dependencies to resolve")
	flag.StringVar(&gopath, "gopath", "/usr/bin/go", "Yaml config file with dependencies to resolve")
}

func main() {
	flag.Parse()
	baseDir, err := os.Getwd()
	quitIfErr(err)

	print("config = ", config)
	yamlBytes, err := ioutil.ReadFile(config)
	//print(string(yamlBytes))
	quitIfErr(err)

	var d []Dep
	err = goyaml.Unmarshal(yamlBytes, &d)
	//print(d)
	quitIfErr(err)

	loadDependencies(baseDir, d)
}

// The dependency struct, provides the data for dependeny info
type Dep struct {
	Path   string // The path to source:   github.com/araddon/gou, launchpad.net/goyaml, etc
	Hash   string // the hash to checkout to, if nil, not used
	Branch string // the branch to checkout if not supplied, uses default
}

func loadDependencies(dirName string, d []Dep) {
	//infos, err := ioutil.ReadDir(dirName)
	//print(infos)
	//quitIfErr(err)
	for _, dep := range d {
		print(dep)
		out, err := exec.Command(gopath, "get", dep.Path).Output()
		quitIfErr(err)
		printf("result is %s\n", out)
	}
}

func quitIfErr(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}
func print(args ...interface{}) {
	fmt.Println("depr: " + fmt.Sprint(args...))
}
func printf(fmtStr string, args ...interface{}) {
	fmt.Printf("depr: "+fmtStr+"\n", args...)
}
