package deprlib

import (
	"errors"
	"fmt"
	u "github.com/araddon/gou"
	"os"
	"os/exec"
	"strings"
	"sync"
)

var (
	GoPath    = ""
	GoCmdPath = "/usr/bin/go"
)

func init() {
	GoPath = os.Getenv("GOPATH")
}

// List of dependencies describing the specific packages versions etc
type Dependencies []*Dep

// Run the dependency resolution, first check cleanliness on all branches
// before proceeding
func (d *Dependencies) Run(allowNonClean bool) error {
	d.init()
	if d.checkClean(allowNonClean) {
		return errors.New("Unclean Directories")
	}
	d.load()
	return nil
}
func (d *Dependencies) init() {
	for _, dep := range *d {
		dep.setup()
	}
}
func (d Dependencies) checkClean(allowNonClean bool) bool {
	var wg sync.WaitGroup
	hasErrors := false
	for _, dep := range d {
		wg.Add(1)
		go func(depIn *Dep) {
			depIn.createPath()
			// generally we are going to force clean on all directories unless overridden
			if !allowNonClean {
				if !depIn.Clean() {
					u.Debug(depIn)
					hasErrors = true
				}
			}
			wg.Done()
		}(dep)
	}
	wg.Wait()
	return hasErrors
}

// Check all the dependencies and make sure they are clean, no uncommited changes
// if so we are going to fail now
func (d Dependencies) XXXCheckClean() bool {
	clean := true
	var wg sync.WaitGroup
	for _, dep := range d {
		wg.Add(1)
		go func(depIn *Dep) {
			if !depIn.Clean() {
				clean = false
			}
			wg.Done()
		}(dep)

	}
	wg.Wait()
	return clean
}

func (d Dependencies) load() {
	var wg sync.WaitGroup
	for _, dep := range d {
		wg.Add(1)
		go func(depIn *Dep) {
			if !depIn.Load() {
				u.Errorf("FAILED, not loaded  %v", depIn)
			}
			if depIn.Build {
				depIn.Buildr()
			}
			wg.Done()
		}(dep)
	}
	wg.Wait()
}

// A sourcecontrol interface allow different implementations of git, hg, etc
type SourceControl interface {
	// Check if this folder/path is clean to determine if there are changes
	// that are uncommited
	CheckClean(*Dep) error
	Checkout(*Dep) error
	Clone(*Dep) error
	Pull(*Dep) error
}

// The dependency struct, provides the data for dependeny info
// * Each Dep represents one package/dependency
type Dep struct {
	exists  bool   // does this path exist?
	Src     string // The path to source:   github.com/araddon/gou, launchpad.net/goyaml, etc
	As      string // the Path to emulate if getting from different Path
	Hash    string // the hash to checkout to, if nil, not used
	Branch  string // the branch to checkout if not supplied, uses default
	Build   bool   // should we build it?
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
	} else {
		d.control = &GoGet{}
	}
}

// The source path
func (d *Dep) SrcPath() string {
	return fmt.Sprintf("%s/src/%s", GoPath, d.Src)
}

// The local disk path
func (d *Dep) AsPath() string {
	if len(d.As) > 0 {
		return fmt.Sprintf("%s/src/%s", GoPath, d.As)
	}
	return fmt.Sprintf("%s/src/%s", GoPath, d.Src)
}

// The local disk directory, the path that will be used
// for importing into go projects, may not be same as source path
func (d *Dep) AsDir() string {
	parts := strings.Split(d.As, "/")
	if len(parts) < 2 {
		u.Errorf("Missing as?   %s", d.As)
		return ""
	}
	return strings.Replace(d.As, "/"+parts[len(parts)-1], "", -1)
}

// Does this dependency need a checkout?   Ie, is not able to use *go get*
func (d *Dep) NeedsCheckout() bool {
	return len(d.Branch) > 0 || len(d.Hash) > 0
}

// ensure this path exists
func (d *Dep) createPath() error {
	fi, err := os.Stat(d.AsPath())
	if err != nil && strings.Contains(err.Error(), "no such file or directory") {
		d.exists = false
		u.Debugf("Creating dir %s", d.AsPath())
		if err := os.MkdirAll(d.AsPath(), os.ModeDir|0700); err != nil {
			u.Error(err)
			return err
		}
		return nil
	}
	if fi != nil && fi.IsDir() {
		d.exists = true
	}

	return nil
}

// Check if this folder/path is clean to determine if there are changes
// that are uncommited
func (d *Dep) Clean() bool {
	if err := d.control.CheckClean(d); err != nil {
		u.Error(err)
		return false
	}
	return true
}

// Load the source for this dependency
//  - Check to see if it uses "As" to alias source if so, doesn't use go get
//  - if we have a source control provider we will use that
func (d *Dep) Load() bool {
	if err := d.control.Pull(d); err != nil {
		u.Errorf("FAILED, not loaded  %v", err)
		return false
	}

	if err := d.control.Checkout(d); err != nil {
		return false
	}

	return true
}

func (d *Dep) Buildr() {
	// new, initial clone?
	u.Warnf("building %s", d.AsDir())
	cmd := exec.Command("go", "clean")
	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	cmd = exec.Command("go", "install")
	out, err = cmd.Output()
	if err != nil {
		u.Error(string(out), err)
	}
}

func quitIfErr(err error) {
	if err != nil {
		u.LogD(4, u.ERROR, "Error: ", err)
		os.Exit(1)
	}
}
