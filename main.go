package main

import (
	"flag"
	"github.com/araddon/depr/deprlib"
	u "github.com/araddon/gou"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"os"
)

var (
	config        string
	goCmdPath     string
	allowNonClean bool
)

func init() {
	flag.StringVar(&config, "config", ".depr.yaml", "Yaml config file with dependencies to resolve")
	flag.StringVar(&goCmdPath, "gopath", "/usr/bin/go", "Yaml config file with dependencies to resolve")
	flag.BoolVar(&allowNonClean, "no-clean", false, "Allow dirty branches?  (uncommited changes)")
}

func main() {
	flag.Parse()
	u.SetLogger(log.New(os.Stderr, "", log.Ltime|log.Lshortfile), "debug")

	yamlBytes, err := ioutil.ReadFile(config)
	//Debug(string(yamlBytes))
	quitIfErr(err)

	deprlib.GoCmdPath = goCmdPath

	var d deprlib.Dependencies
	err = goyaml.Unmarshal(yamlBytes, &d)
	quitIfErr(err)

	d.Run(allowNonClean)
}

func quitIfErr(err error) {
	if err != nil {
		u.LogD(4, u.ERROR, "Error: ", err)
		os.Exit(1)
	}
}
