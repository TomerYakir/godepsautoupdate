package main

import (
	"fmt"
	"os/exec"
	"path"
	"strings"
)

func getLatestGitCommit(gitpath string) (string, string) {
	// git --no-pager log --pretty=format:"%H,%cd,%cr"
	cmd := exec.Command("git", "--no-pager", "-C", gitpath, "log", "--pretty=format:\"%H,%cd,%cr\"", "-n", "1")
	logDebug("running command %v", *cmd)
	out, err := cmd.Output()
	if err != nil {
		panic(fmt.Sprintf("failed to get git log for %s. err: %v", gitpath, err))
	}
	lines := strings.Split(string(out), "\n")
	if len(lines) == 0 {
		// TODO - error
	}
	tokens := strings.Split(lines[0], ",")
	return tokens[0], fmt.Sprintf("%s (%s)", tokens[1], tokens[2])
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
