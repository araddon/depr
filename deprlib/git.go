package deprlib

import (
	"fmt"
	u "github.com/araddon/gou"
	"os/exec"
	"strings"
)

var (
	BRANCHES = "master,develop,gh-pages"
)

// Implementation of the git interface for managing checkouts
type Git struct{}

// Check to see if this dependency is clean or not
func (s *Git) CheckClean(d *Dep) error {
	//git diff --exit-code
	cmd := exec.Command("git", "diff", "--exit-code")
	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if len(out) == 0 && err == nil {
		return nil
	}
	return fmt.Errorf("GIT NOT CLEAN: %s", d.AsPath())
}

// Initial Creation of this repo
func (s *Git) Clone(d *Dep) error {
	if !d.exists {
		// new, initial clone?
		u.Warnf("cloning src? %s", d.AsDir())
		cmdgit := exec.Command("git", "clone", d.As)
		cmdgit.Dir = d.AsDir()
		out, err := cmdgit.Output()
		u.Debug(out, err)
		return err
	}
	return nil
}

// Initial Pull
func (s *Git) Pull(d *Dep) error {
	var cmd *exec.Cmd
	if len(d.Hash) > 0 && d.exists && !strings.Contains(BRANCHES, d.Hash) {
		// we are in detached head mode at the moment most likely, get onto a branch
		cmd = exec.Command("git", "checkout", "master")
		cmd.Dir = d.AsPath()
		out, err := cmd.Output()
		if err != nil {
			u.Errorf("GIT PULL ERR out='%s'  err=%v  cmd=%v", out, err)
			return err
		}
		u.Debugf("hash checkout master (hash=%v) path:%s  out='%s'", d.Hash, d.AsPath(), string(out))
	}
	//now do a git pull after ensuring we are on a branch?
	cmd = exec.Command("git", "pull")
	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if err != nil {
		u.Errorf("GIT PULL ERR out='%s'  err=%v  cmd=%v", out, err, cmd)
		return err
	}
	u.Debugf("Git pull? %s   %s", d.AsPath(), chompnl(out))
	return nil

}

// Checkout appropriate branch if any
func (s *Git) Checkout(d *Dep) error {
	var cmd *exec.Cmd

	//git checkout hash
	if len(d.Hash) > 0 {
		cmd = exec.Command("git", "checkout", d.Hash)
		u.Debugf("hash checkout (%v) path:%s", cmd.Args, d.AsPath())
	} else if len(d.Branch) > 0 {
		u.Debugf("git checkout3 %s  as:%s", d.Branch, d.AsPath())
		cmd = exec.Command("git", "checkout", d.Branch)
	} else {
		u.Debugf("git checkout4 master  as:%s", d.AsPath())
		cmd = exec.Command("git", "checkout", "master")
	}

	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if err != nil {
		u.Errorf("out='%s'  err=%v  cmd=%v", out, err, cmd)
		return err
	}

	return nil
}
