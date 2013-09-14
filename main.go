package main

import (
	"flag"
	"github.com/araddon/depr/deprlib"
	u "github.com/araddon/gou"
	"io/ioutil"
	"launchpad.net/goyaml"
	"os"
	"runtime"
)

var (
	config        string
	goCmdPath     string
	allowNonClean bool
	baseBranch    string
)

func init() {
	flag.StringVar(&config, "config", ".depr.yaml", "Yaml config file with dependencies to resolve")
	flag.StringVar(&goCmdPath, "gopath", "/usr/local/go/bin/go", "Path to go binary")
	flag.BoolVar(&allowNonClean, "no-clean", false, "Allow dirty branches?  (uncommited changes)")
	flag.StringVar(&baseBranch, "basebranch", "develop", "Branch to use if none specified")
}

func main() {
	flag.Parse()
	u.SetupLogging("debug")
	u.SetColorIfTerminal()
	deprlib.BASE_BRANCH = baseBranch

	yamlBytes, err := ioutil.ReadFile(config)
	//Debug(string(yamlBytes))
	quitIfErr(err)
	goroot := runtime.GOROOT()
	goCmdPath = goroot + "/bin/go"
	deprlib.GoCmdPath = goCmdPath

	var d deprlib.Dependencies
	err = goyaml.Unmarshal(yamlBytes, &d)

	err = d.Run(allowNonClean)
	quitIfErr(err)
}

func quitIfErr(err error) {
	if err != nil {
		u.LogD(4, u.ERROR, "Error: ", err)
		os.Exit(1)
	}
}
