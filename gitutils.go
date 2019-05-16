package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

func getCommitDiffSummary(gitpath string, oldcommit, newcommit string) (string, error) {
	// git diff --shortstat oldcommit newcommit
	logDebug("getting diff summary for %s", gitpath)
	cmd := exec.Command("git", "-C", gitpath, "diff", "--shortstat", oldcommit, newcommit)
	logDebug("running command %v", *cmd)
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

func getLatestGitCommit(gitpath string) (string, string, error) {
	// git --no-pager log --pretty=format:"%H,%cd,%cr"
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "log", "--pretty=format:\"%H;%cd;%cr\"", "-n", "1")
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("failed to get git log for %s. err: %v", gitpath, err))
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("Failed to get git log output for %s", gitpath)
	}
	tokens := strings.Split(lines[0], ";")
	return clearQuotes(tokens[0]), fmt.Sprintf("%s (%s)", tokens[1], clearQuotes(tokens[2])), nil
}

func getLatestGitCommitByTag(gitpath string) (string, string, string, error) {
	// git --no-pager tag --format='%(creatordate:iso);%(creatordate:relative);%(refname:strip=2)' --sort=tag
	ignorableTags := []string{"rc", "night", "unstable"}
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "tag", "--format=\"%(creatordate:iso);%(creatordate:relative);%(refname:strip=2)\"", "--sort=tag")
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("failed to get git log tag %s. err: %v", gitpath, err)) // TODO - replace panic
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
	commit, err := getCommitByTag(gitpath, latestTag)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get commit for tag %s of package %s", latestTag, gitpath)
	}
	return clearQuotes(commit), clearQuotes(latestTag), fmt.Sprintf("%s (%s)", clearQuotes(latestTagDate), clearQuotes(latestTagRelDate)), nil
}

func getCommitByTag(gitpath, tag string) (string, error) {
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "log", "--pretty=format:\"%H\"", "-1", clearQuotes(tag))
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("failed to get git log tag %s. err: %v", gitpath, err)) // TODO - replace panic
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Failed to get git tag output for %s", gitpath)
	}
	return clearQuotes(lines[0]), nil
}

func stringContains(s []string, v string) bool {
	for _, sv := range s {
		if strings.Contains(v, sv) {
			return true
		}
	}
	return false
}

func getGitRemoteUrl(gitpath string) (string, error) {
	cmd := exec.Command("git", "-C", gitpath, "config", "--get", "remote.origin.url")
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git remote url for %s. err: %v", gitpath, err)
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", fmt.Errorf("Failed to get git remote url output for %s", gitpath)
	}
	tokens := strings.Split(lines[0], ",")
	return clearQuotes(tokens[0]), nil
}

func getGitRoot(godepsPath string) string {
	cmd := exec.Command("git", "-C", path.Dir(godepsPath), "rev-parse", "--show-toplevel")
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("failed to get git root. err: %v", err))
	}
	return strings.Trim(string(out), "\n")
}

func goget(gopath, gogetpath, packagePath, gitremote string) {
	logDebug("getting package %s", gogetpath)
	cmd := exec.Command("go", "get", "-d", gogetpath)
	cmd.Dir = path.Join(gopath, "src")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", gopath))
	logDebug("running command %v", *cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "no Go files in") {
			panic(fmt.Sprintf("failed to run go get for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err))
		}
	}
	addRemote(gogetpath, gitremote, packagePath)
}

func addRemote(gogetpath, gitremote, packagePath string) {
	if gitremote != "" {
		logDebug("adding remote %s to %s", gitremote, packagePath)
		// git remote add downstream ""
		cmd := exec.Command("git", "-C", packagePath, "remote", "add", "downstream", gitremote)
		logDebug("running command %v", *cmd)
		out, err := cmd.CombinedOutput()
		if err != nil {
			if !strings.Contains(string(out), "remote downstream already exists") {
				panicWithMessage("failed to run git remote add for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err)
			}
		}
		// git fetch downstream
		cmd = exec.Command("git", "-C", packagePath, "fetch", "downstream")
		logDebug("running command %v", *cmd)
		out, err = cmd.CombinedOutput()
		if err != nil {
			panicWithMessage("failed to run git fetch downstream for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err)
		}
	}
}

func gitpull(packagePath string) {
	cmd := exec.Command("git", "-C", packagePath, "checkout", "master")
	logDebug("running command %v", *cmd)
	out, err := cmd.CombinedOutput()
	if err != nil {
		logInfo("failed to run git pull for package %s.\nout: %v\nerr: %v", packagePath, string(out), err)
	}

	cmd = exec.Command("git", "-C", packagePath, "pull")
	out, err = cmd.CombinedOutput()
	if err != nil {
		logInfo("failed to run git pull for package %s.\nout: %v\nerr: %v", packagePath, string(out), err)
	}
}
