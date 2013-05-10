package deprlib

import (
	"errors"
	"fmt"
	. "github.com/araddon/gou"
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
	// generally we are going to force clean on all directories unless overridden
	if !allowNonClean {
		if !d.CheckClean() {
			return errors.New("THERE ARE UNCLEAN DIRS")
		}
	}
	d.load()
	return nil
}
func (d *Dependencies) init() {
	for _, dep := range *d {
		dep.setup()
	}
}
func (d Dependencies) load() {
	var wg sync.WaitGroup
	for _, dep := range d {
		wg.Add(1)
		go func(depIn *Dep) {
			if !depIn.Load() {
				Logf(ERROR, "FAILED, not loaded  %v", depIn)
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
	CheckClean(*Dep) (bool, error)
	Checkout(*Dep) (bool, error)
}

// The dependency struct, provides the data for dependeny info
// * Each Dep represents one package/dependency
type Dep struct {
	exists  bool   // does this path exist?
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
		d.exists = false
		return true
	}
	if fi != nil && fi.IsDir() {
		d.exists = true
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
//  - if we have a source control provider we will use that
func (d *Dep) Load() bool {
	if len(d.As) > 0 || d.NeedsCheckout() {
		if didCheckout, err := d.control.Checkout(d); didCheckout && err == nil {
			return true
		}
	} else {
		// go get -u leaves git in detached head state
		// so we can't get pull in future, so don't use it if we have a choice
		if d.control != nil && d.exists {
			if didCheckout, err := d.control.Checkout(d); didCheckout && err == nil {
				return true
			} else {
				Logf(ERROR, "%v", err)
				return false
			}
		}
		// use Go Get?  Should we specify?  How do we do a go get -u?
		Debugf("go get -u '%s'", d.Src)
		_, err := exec.Command(GoCmdPath, "get", "-u", d.Src).Output()
		quitIfErr(err)
		if d.control != nil {
			// Try to checkout master, to prevent detached head and non updating
			if didCheckout, err := d.control.Checkout(d); didCheckout && err == nil {
				return true
			} else {
				Logf(ERROR, "%v", err)
				return false
			}
		}
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

func quitIfErr(err error) {
	if err != nil {
		LogD(4, ERROR, "Error: ", err)
		os.Exit(1)
	}
}
