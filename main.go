package main

import (
	"flag"
	"fmt"
	. "github.com/araddon/gou"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"os"
	"os/exec"
	"strings"
)

var (
	config        string
	goCmdPath     string
	gopath        string
	allowNonClean bool
)

func init() {
	gopath = os.Getenv("GOPATH")
	flag.StringVar(&config, "config", ".depr.yaml", "Yaml config file with dependencies to resolve")
	flag.StringVar(&goCmdPath, "gopath", "/usr/bin/go", "Yaml config file with dependencies to resolve")
	flag.BoolVar(&allowNonClean, "no-clean", false, "Allow dirty branches?  (uncommited changes)")
}

func main() {
	flag.Parse()
	SetLogger(log.New(os.Stderr, "", log.Ltime|log.Lshortfile), "debug")

	yamlBytes, err := ioutil.ReadFile(config)
	quitIfErr(err)

	var d Dependencies
	err = goyaml.Unmarshal(yamlBytes, &d)
	quitIfErr(err)

	d.Run()
}

// List of dependencies describing the specific packages versions etc 
type Dependencies []*Dep

// Run the dependency resolution, first check cleanliness on all branches
// before proceeding
func (d *Dependencies) Run() {
	d.init()
	if !d.CheckClean() {
		Log(ERROR, "THERE ARE UNCLEAN DIRS")
		return
	}
	d.load()
}
func (d *Dependencies) init() {
	for _, dep := range *d {
		dep.setup()
	}
}
func (d Dependencies) load() {
	for _, dep := range d {
		if !dep.Load() {
			Debugf("FAILED, not loaded  %v", dep)
			return
		}
	}
}

// A sourcecontrol interface allow different implementations of git, hg, etc
type SourceControl interface {
	// Check if this folder/path is clean to determine if there are changes
	// that are uncommited
	CheckClean(*Dep) (bool, error)
	Checkout(*Dep) (bool, error)
}

// Implementation of the git interface for managing checkouts
type Git struct{}

// Check to see if this dependency is clean or not 
func (s *Git) CheckClean(d *Dep) (bool, error) {
	//git diff --exit-code
	cmd := exec.Command("git", "diff", "--exit-code")
	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if len(out) == 0 && err == nil {
		return true, nil
	}
	Debug("GIT NOT CLEAN: ", d.AsPath())
	return false, nil
}

// Checkout appropriate branch if any
func (s *Git) Checkout(d *Dep) (bool, error) {
	var cmd *exec.Cmd
	Debugf("checkout %s:%s", d.As, d.Src)
	if len(d.As) > 0 {
		// Check to see if we already have it?
		_, err := os.Stat(d.AsPath())
		if err != nil && strings.Contains(err.Error(), "no such file or directory") {
			//TODO:  os.Mkdir(name, perm)
			Logf(WARN, "Creating dir? %s", d.AsDir())
			cmdDir := exec.Command("mkdir", "-p", d.AsDir())
			cmdDir.Dir = gopath
			cmdDir.Output()

			// Clone 
			Logf(WARN, "cloning src? %s", d.AsDir())
			cmdgit := exec.Command("git", "clone", d.As)
			cmdgit.Dir = d.AsDir()
			out, err := cmdgit.Output()
			Debug(out, err)
		}
		// GIT UPDATE!!!!
		Logf(WARN, "git pull? %s", d.As)
		cmdgit := exec.Command("git", "pull")
		cmdgit.Dir = d.AsPath()
		out, err := cmdgit.Output()
		Debug(string(out), err)
	}
	//git checkout hash
	if len(d.Hash) > 0 {
		Debugf("git checkout %s    # %s", d.Hash, d.SrcPath())
		cmd = exec.Command("git", "checkout", d.Hash)
	} else if len(d.Branch) > 0 {
		cmd = exec.Command("git", "checkout", d.Branch)
	} else {
		//??   git pull?
		cmd = exec.Command("git", "pull")
	}

	cmd.Dir = d.SrcPath()
	out, err := cmd.Output()
	if err != nil {
		Logf(ERROR, "out='%s'  err=%v  cmd=%v", out, err, cmd)
		return false, err
	}

	return true, nil
}

// The dependency struct, provides the data for dependeny info
// * Each Dep represents one package/dependency
type Dep struct {
	Src     string // The path to source:   github.com/araddon/gou, launchpad.net/goyaml, etc
	As      string // the Path to emulate if getting from different Path
	Hash    string // the hash to checkout to, if nil, not used
	Branch  string // the branch to checkout if not supplied, uses default
	control SourceControl
}

func (d *Dep) setup() {
	parts := strings.Split(d.Src, "#")
	//Debugf("parts=%v len(parts)=%d   hashlen=%d", parts, len(parts), len(d.Hash))
	if len(parts) > 1 && len(d.Hash) == 0 {
		d.Hash = parts[1]
		d.Src = parts[0]
		Debugf("Setting hash to %s: %s", d.Src, d.Hash)
	}
	// now setup our source control provider
	if strings.Contains(d.Src, "github.com") {
		d.control = &Git{}
		//Debugf("Setting src to github for %s : %v", d.Src, d.control)
	}
}

// The source path 
func (d *Dep) SrcPath() string {
	return fmt.Sprintf("%s/src/%s", gopath, d.Src)
}

// The local disk path
func (d *Dep) AsPath() string {
	if len(d.As) > 0 {
		return fmt.Sprintf("%s/src/%s", gopath, d.As)
	}
	return fmt.Sprintf("%s/src/%s", gopath, d.Src)
}

// The local disk directory
func (d *Dep) AsDir() string {
	parts := strings.Split(d.As, "/")
	if len(parts) < 2 {
		Logf(ERROR, "Missing as?   %s", d.As)
		return ""
	}
	return strings.Replace(d.As, "/"+parts[len(parts)-1], "", -1)
}

// Does this dependency need a checkout?   Ie, is not able to use *go get*
func (d *Dep) NeedsCheckout() bool {
	return (len(d.Branch) > 0 && d.Branch != "master") || len(d.Hash) > 0
}

// Check if this folder/path is clean to determine if there are changes
// that are uncommited
func (d *Dep) Clean() bool {
	// if the directory doesn't exist it is clean
	fi, err := os.Stat(d.AsPath())
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		return true
	}
	if fi != nil && fi.IsDir() {
		if clean, err := d.control.CheckClean(d); clean && err == nil {
			return true
		}
	}

	return false
}

// Load the source for this dependency
//  - Check to see if it uses "As" to alias source if so, doesn't use go get
func (d *Dep) Load() bool {
	if len(d.As) > 0 {
		if didCheckout, err := d.control.Checkout(d); didCheckout && err == nil {
			return true
		}
	} else {
		//Debugf("go getting:  %v", d)
		_, err := exec.Command(goCmdPath, "get", d.Src).Output()
		//Debugf("go get: '%s'   err=%v\n", out, err)
		quitIfErr(err)
		if d.NeedsCheckout() {
			Debugf("Needs checkout? %s hash=%s branch=%s", d.Src, d.Hash, d.Branch)
			if didCheckout, err := d.control.Checkout(d); didCheckout && err == nil {
				return true
			}
		} else {
			return true
		}
	}

	return false
}

// Check all the dependencies and make sure they are clean, no uncommited changes
// if so we are going to fail now
func (d Dependencies) CheckClean() bool {
	clean := true
	for _, dep := range d {
		if !dep.Clean() {
			clean = false
		}
	}
	return clean
}

func quitIfErr(err error) {
	if err != nil {
		Log(ERROR, "Error: %v", err)
		os.Exit(1)
	}
}
