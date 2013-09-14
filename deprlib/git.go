package deprlib

import (
	"fmt"
	u "github.com/araddon/gou"
	"os/exec"
	"strings"
)

var (
	BRANCHES    = "master,develop,gh-pages"
	BASE_BRANCH = "develop"
)

// Implementation of the git interface for managing checkouts
type Git struct{}

// Check to see if this dependency is clean or not
func (s *Git) CheckClean(d *Dep) error {
	//git diff --exit-code
	cmd := exec.Command("git", "diff", "--exit-code")
	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if len(out) == 0 || err == nil {
		return nil
	}
	u.Debugf("out? %v", string(out))
	return fmt.Errorf("GIT NOT CLEAN: %s", d.AsPath())
}

// Initial Creation of this repo
func (s *Git) Clone(d *Dep) error {
	if !d.exists {
		// new, initial clone?
		// git@github.com:lytics/cache.git
		parts := strings.Split(d.Src, "/")
		// 0: github.com  1:lytics   2:cache
		if len(parts) < 2 {
			return fmt.Errorf("Invalid src?  %s", d.Src)
		}
		gitPath := fmt.Sprintf("git@%s:%s/%s.git", parts[0], parts[1], parts[2])
		u.Warnf("cloning src? %s", gitPath)
		cmdgit := exec.Command("git", "clone", gitPath)
		cmdgit.Dir = d.ParentDir()
		out, err := cmdgit.Output()
		u.Debug(string(out), err)
		return err
	}
	return nil
}

func (s *Git) Pull(d *Dep) error {
	// need to refactor order of these
	return nil
}

func (s *Git) getRightBranch(d *Dep) error {
	var cmd *exec.Cmd

	//git checkout hash
	if len(d.Hash) > 0 {
		cmd = exec.Command("git", "checkout", d.Hash)
		u.Debugf("hash checkout (%v) path:%s", cmd.Args, d.AsPath())
	} else if len(d.Branch) > 0 {
		u.Debugf("git checkout3 %s  as:%s", d.Branch, d.AsPath())
		cmd = exec.Command("git", "checkout", d.Branch)
	} else {
		//u.Debugf("git checkout4 master  as:%s", d.AsPath())
		cmd = exec.Command("git", "checkout", "master")
	}

	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if err != nil && len(out) > 0 {
		u.Errorf("out='%s'  err=%v  cmd=%v", out, err, cmd)
		return err
	}
	return nil
}

func (s *Git) doPull(d *Dep) error {
	var cmd *exec.Cmd

	//now fetch or pull?
	cmd = exec.Command("git", "pull")
	cmd.Dir = d.AsPath()
	out, err := cmd.Output()
	if err != nil && len(out) > 0 {
		u.Errorf("GIT PULL ERR out='%s' %s err=%v  cmd=%v", out, d.AsPath(), err, cmd)
		if err := s.fixDetachedHead(d); err == nil {
			return nil
		}
		return err
	}
	return nil
}

func (s *Git) fixDetachedHead(d *Dep) error {
	// if has d.Hash and the hash is not a known branch (ie:  develop,master,gh-pages,etc)
	if len(d.Hash) > 0 && d.exists && !strings.Contains(BRANCHES, d.Hash) {
		// we are possibly in in detached head mode at the moment
		// lets try checking out master and repeating, to onto a branch
		cmd := exec.Command("git", "checkout", BASE_BRANCH)
		cmd.Dir = d.AsPath()
		out, err := cmd.Output()
		if err != nil && len(out) > 0 {
			u.Errorf("GIT PULL ERR out='%s'  err=%v  cmd=%v", out, err)
			return err
		}
		u.Debugf("hash checkout %s (hash=%v) path:%s  out='%s'",
			BASE_BRANCH, d.Hash, d.AsPath(), string(out))
	} else {
		u.Debugf("on master, just pull: %s", d.AsPath())
	}
	return nil
}

// Checkout appropriate branch if any
func (s *Git) Checkout(d *Dep) error {
	if err := s.getRightBranch(d); err != nil {
		return err
	}
	if err := s.doPull(d); err != nil {
		return err
	}

	return nil
}
