package gitutils

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/tomeryakir/gdau/utils"
)

// GetCommitDiffSummary - getting diff summary between two commits
func GetCommitDiffSummary(gitpath string, oldcommit, newcommit string, logger *utils.Logger) (string, error) {
	// git diff --shortstat oldcommit newcommit
	logger.LogDebug("getting diff summary for %s", gitpath)
	cmd := exec.Command("git", "-C", gitpath, "diff", "--shortstat", oldcommit, newcommit)
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to get git diff for %s. err: %v. out: %v", gitpath, err, string(out))
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Failed to get git diff output for %s", gitpath)
	}
	return lines[0], nil
}

// GetLatestGitCommit - getting latest git commit
func GetLatestGitCommit(gitpath string, logger *utils.Logger) (string, string, error) {
	// git --no-pager log --pretty=format:"%H,%cd,%cr"
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "log", "--pretty=format:\"%H;%cd;%cr\"", "-n", "1")
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		logger.PanicWithMessage("failed to get git log for %s. err: %v", gitpath, err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("Failed to get git log output for %s", gitpath)
	}
	tokens := strings.Split(lines[0], ";")
	return utils.ClearQuotes(tokens[0]), fmt.Sprintf("%s (%s)", tokens[1], utils.ClearQuotes(tokens[2])), nil
}

// GetLatestGitCommitByTag - getting latest git commit by tag
func GetLatestGitCommitByTag(gitpath string, logger *utils.Logger) (string, string, string, error) {
	// git --no-pager tag --format='%(creatordate:iso);%(creatordate:relative);%(refname:strip=2)' --sort=tag
	ignorableTags := []string{"rc", "night", "unstable"}
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "tag", "--format=\"%(creatordate:iso);%(creatordate:relative);%(refname:strip=2)\"", "--sort=tag")
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		logger.PanicWithMessage(fmt.Sprintf("failed to get git log tag %s. err: %v", gitpath, err)) // TODO - replace panic
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", "", "", fmt.Errorf("Failed to get git tag output for %s", gitpath)
	}
	var latestTag string
	var latestTagDate string
	var latestTagRelDate string
	for _, line := range lines {
		tokens := strings.Split(line, ";")
		if len(tokens) < 2 {
			continue
		}
		if !stringContains(ignorableTags, tokens[2]) {
			latestTagDate = tokens[0]
			latestTagRelDate = tokens[1]
			latestTag = tokens[2]
		}
	}
	commit, err := GetCommitByTag(gitpath, latestTag, logger)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get commit for tag %s of package %s", latestTag, gitpath)
	}
	return utils.ClearQuotes(commit), utils.ClearQuotes(latestTag), fmt.Sprintf("%s (%s)", utils.ClearQuotes(latestTagDate), utils.ClearQuotes(latestTagRelDate)), nil
}

// GetCommitByTag - getting commit for tag
func GetCommitByTag(gitpath, tag string, logger *utils.Logger) (string, error) {
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "log", "--pretty=format:\"%H\"", "-1", utils.ClearQuotes(tag))
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		logger.PanicWithMessage("failed to get git log tag %s. err: %v", gitpath, err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Failed to get git tag output for %s", gitpath)
	}
	return utils.ClearQuotes(lines[0]), nil
}

func stringContains(s []string, v string) bool {
	for _, sv := range s {
		if strings.Contains(v, sv) {
			return true
		}
	}
	return false
}

// GetGitRemoteURL - get remote origin url
func GetGitRemoteURL(gitpath string, logger *utils.Logger) (string, error) {
	var err error
	var out []byte
	cmd := exec.Command("git", "-C", gitpath, "config", "--get", "remote.origin.url")
	logger.LogDebug("running command %v", *cmd)
	out, err = cmd.Output()
	if err != nil {
		// try with remote.downstream.url
		logger.LogDebug("trying with fallback method")
		cmd = exec.Command("git", "-C", gitpath, "config", "--get", "remote.downstream.url")
		logger.LogDebug("running (fallback) command %v", *cmd)
		out, err = cmd.Output()
		if err != nil {
			return "", fmt.Errorf("failed to get git remote url for %s. err: %v", gitpath, err)
		}
	}

	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Failed to get git remote url output for %s", gitpath)
	}
	tokens := strings.Split(lines[0], ",")
	return utils.ClearQuotes(tokens[0]), nil
}

// GetGitRoot - get git root
func GetGitRoot(godepsPath string, logger *utils.Logger) string {
	cmd := exec.Command("git", "-C", path.Dir(godepsPath), "rev-parse", "--show-toplevel")
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		logger.PanicWithMessage("failed to get git root. err: %v", err)
	}
	return strings.Trim(string(out), "\n")
}

// Goget - get go package
func Goget(gopath, gogetpath, packagePath, gitremote string, logger *utils.Logger) error {
	logger.LogDebug("getting package %s", gogetpath)
	cmd := exec.Command("go", "get", gogetpath)
	cmd.Dir = path.Join(gopath, "src")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", gopath))
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "no Go files in") {
			return fmt.Errorf("failed to run go get for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err)
		}
	}
	return AddRemote(gogetpath, gitremote, packagePath, logger)
}

// AddRemote - add remote repo to path
func AddRemote(gogetpath, gitremote, packagePath string, logger *utils.Logger) error {
	if gitremote != "" {
		logger.LogDebug("adding remote %s to %s", gitremote, packagePath)
		// git remote add downstream ""
		cmd := exec.Command("git", "-C", packagePath, "remote", "add", "downstream", gitremote)
		logger.LogDebug("running command %v", *cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if !strings.Contains(string(out), "remote downstream already exists") {
				logger.PanicWithMessage("failed to run git remote add for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err)
			}
		}
		// git fetch downstream
		cmd = exec.Command("git", "-C", packagePath, "fetch", "downstream")
		logger.LogDebug("running command %v", *cmd)
		out, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to run git fetch downstream for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err)
		}
	}
	return nil
}

// Gitpull - git pull
func Gitpull(packagePath string, logger *utils.Logger) {
	cmd := exec.Command("git", "-C", packagePath, "checkout", "master")
	logger.LogDebug("running command %v", *cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logger.LogInfo("failed to run git pull for package %s.\nout: %v\nerr: %v", packagePath, string(out), err)
	}

	cmd = exec.Command("git", "-C", packagePath, "pull")
	out, err = cmd.CombinedOutput()
	if err != nil {
		logger.LogInfo("failed to run git pull for package %s.\nout: %v\nerr: %v", packagePath, string(out), err)
	}
}
