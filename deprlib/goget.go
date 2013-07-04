package deprlib

import (
	u "github.com/araddon/gou"
	"os/exec"
	"strings"
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
	out, err := exec.Command(GoCmdPath, "get", "-u", d.Src).Output()
	//no Go source files in /home/ubuntu/Dropbox/go/root/src/code.google.com/p/goprotobuf
	if err != nil {
		if strings.Contains(string(out), "no Go source files") {
			return nil
		} else if len(out) == 0 {
			return nil
		}
	}
	return err
}

// Checkout appropriate branch if any
func (s *GoGet) Checkout(d *Dep) error {

	return nil
}
