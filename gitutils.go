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
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "log", "--pretty=format:\"%H,%cd,%cr\"", "-n", "1")
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("failed to get git log for %s. err: %v", gitpath, err))
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		return "", "", fmt.Errorf("Failed to get git log output for %s", gitpath)
	}
	tokens := strings.Split(lines[0], ",")
	return clearQuotes(tokens[0]), fmt.Sprintf("%s (%s)", tokens[1], clearQuotes(tokens[2])), nil
}

func getGitRemoteUrl(gitpath string) (string, error) {
	// git config --get remote.origin.url
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
	logInfo("getting package %s", gogetpath)
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

	if gitremote != "" {
		logDebug("adding remote %s to %s", gitremote, packagePath)
		// git remote add downstream ""
		cmd = exec.Command("git", "-C", packagePath, "remote", "add", "downstream", gitremote)
		logDebug("running command %v", *cmd)
		out, err = cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf("failed to run git remote add for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err))
		}
		// git fetch downstream
		cmd = exec.Command("git", "-C", packagePath, "fetch", "downstream")
		logDebug("running command %v", *cmd)
		out, err = cmd.CombinedOutput()
		if err != nil {
			panic(fmt.Sprintf("failed to run git remote add for package %s.\nout: %v\nerr: %v", gogetpath, string(out), err))
		}
	}
}
