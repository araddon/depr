package deprlib

import (
	u "github.com/araddon/gou"
	//"os"
	"os/exec"
)

// Implementation of using Go Get, for non github
type GoGet struct{}

// Check to see if this dependency is clean or not
func (s *GoGet) CheckClean(d *Dep) error {
	return nil
}

// Initial Creation of this repo
func (s *GoGet) Clone(d *Dep) error {
	if !d.exists {
		// use Go Get?  Should we specify?  How do we do a go get -u?
		u.Debugf("go get -u '%s'", d.Src)
		_, err := exec.Command(GoCmdPath, "get", "-u", d.Src).Output()
		return err
	}
	return nil
}

// Initial Pull
func (s *GoGet) Pull(d *Dep) error {
	u.Debugf("go get -u '%s'", d.Src)
	_, err := exec.Command(GoCmdPath, "get", "-u", d.Src).Output()
	return err
}

// Checkout appropriate branch if any
func (s *GoGet) Checkout(d *Dep) error {

	return nil
}
