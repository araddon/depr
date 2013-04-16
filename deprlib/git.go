package deprlib

import (
	. "github.com/araddon/gou"
	"os"
	"os/exec"
	"strings"
)

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
			cmdDir.Dir = GoPath
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
