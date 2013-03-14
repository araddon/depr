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
	//Debug(string(yamlBytes))
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
	// generally we are going to force clean on all directories unless overridden
	if !allowNonClean {
		if !d.CheckClean() {
			Log(ERROR, "THERE ARE UNCLEAN DIRS")
			return
		}
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
		//Debugf("was clean: %s", d.AsPath())
		return true, nil
	}
	Logf(WARN, "GIT NOT CLEAN: %s", d.AsPath())
	return false, nil
}

// Checkout appropriate branch if any
func (s *Git) Checkout(d *Dep) (bool, error) {
	var cmd *exec.Cmd
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
		} else {
			// make sure we are not in detached state
			Logf(WARN, "git checkout? src:%s  as:%s", d.Src, d.AsPath())
			branch := "master"
			if len(d.Branch) > 0 {
				branch = d.Branch
			}
			cmdgit := exec.Command("git", "checkout", branch)
			cmdgit.Dir = d.AsPath()
			out, err := cmdgit.Output()
			if err != nil {
				Logf(ERROR, "ERROR on git checkout?  %v    %s", err, out)
				return false, err
			}
		}
		// GIT UPDATE!!!!
		Logf(WARN, "git pull? src:%s  as:%s", d.Src, d.AsPath())
		cmdgit := exec.Command("git", "pull")
		cmdgit.Dir = d.AsPath()
		out, err := cmdgit.Output()
		if err != nil {
			Logf(ERROR, "ERROR on git pull?  %v    %s", err, out)
			return false, err
		}
		return true, nil
	}
	//git checkout hash
	if len(d.Hash) > 0 {
		Debugf("git checkout %s    # %s", d.Hash, d.AsPath())
		Debugf("git checkout %s   # hash", d.Hash)
		cmd = exec.Command("git", "checkout", d.Hash)
	} else if len(d.Branch) > 0 {
		Debugf("git checkout %s  as:%s", d.Branch, d.AsPath())
		cmd = exec.Command("git", "checkout", d.Branch)
	} else {
		//??   git pull?
		Debugf("Git pull? %s", d.AsPath())
		cmd = exec.Command("git", "pull")
	}

	cmd.Dir = d.AsPath()
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
	if len(parts) > 1 && len(d.Hash) == 0 {
		d.Hash = parts[1]
		d.Src = parts[0]
	}
	// now setup our source control provider
	if strings.Contains(d.Src, "github.com") {
		d.control = &Git{}
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

// The local disk directory, the path that will be used
// for importing into go projects, may not be same as source path
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
	return len(d.Branch) > 0 || len(d.Hash) > 0
}

// Check if this folder/path is clean to determine if there are changes
// that are uncommited
func (d *Dep) Clean() bool {
	// if the directory doesn't exist it is clean
	//Debugf("Check clean:  %s", d.AsPath())
	fi, err := os.Stat(d.AsPath())
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		return true
	}
	if fi != nil && fi.IsDir() {
		if d.control == nil {
			return true
		}
		if clean, err := d.control.CheckClean(d); clean && err == nil {
			return true
		}
	}

	return false
}

// Load the source for this dependency
//  - Check to see if it uses "As" to alias source if so, doesn't use go get
func (d *Dep) Load() bool {
	if len(d.As) > 0 || d.NeedsCheckout() {
		if didCheckout, err := d.control.Checkout(d); didCheckout && err == nil {
			return true
		}
	} else {
		// use Go Get?  Should we specify?  How do we do a go get -u?
		Debugf("go get -u '%s'", d.Src)
		_, err := exec.Command(goCmdPath, "get", "-u", d.Src).Output()
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
		LogD(4, ERROR, "Error: ", err)
		os.Exit(1)
	}
}
