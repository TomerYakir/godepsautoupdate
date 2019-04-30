package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
)

var debug bool

type GodepsEntry struct {
	Path                 string
	CommitVersion        string
	GitRemote            string
	GitType              EntryType
	Status               EntryStatus
	NewCommitVersion     string
	NewCommitDateSummary string
	NewCommitVersionURL  string
	Summary              string
}

type EntryStatus int

const (
	Uptodate EntryStatus = 0
	Outdated EntryStatus = 1
)

type EntryType int

const (
	Commit        EntryType = 0
	BranchVersion EntryType = 1
)

type analyzeResult struct {
	path                 string
	Status               EntryStatus
	NewCommitVersion     string
	NewCommitDateSummary string
	NewCommitVersionURL  string
	Summary              string
}

func NewGoDepsEntry(path, commitVersion, gitRemote string) *GodepsEntry {
	g := &GodepsEntry{}
	g.Path = path
	g.CommitVersion = commitVersion
	g.GitRemote = gitRemote
	if isHexString(commitVersion) {
		g.GitType = Commit
	} else {
		g.GitType = BranchVersion
	}
	return g
}

func main() {
	var godepsPath string
	var gopath string
	var report bool
	var update bool

	flag.StringVar(&godepsPath, "path", "", "path to godeps")
	flag.StringVar(&gopath, "gopath", "", "path to packages root")
	flag.BoolVar(&report, "report", false, "generate an HTML report")
	flag.BoolVar(&update, "update", false, "update the godeps file")
	flag.BoolVar(&debug, "debug", false, "turn on debug")

	flag.Parse()

	if godepsPath == "" {
		panic("Godeps path wasn't specified")
	}
	if gopath == "" {
		panic("Gopath wasn't specified")
	}
	gitRoot := getGitRoot(godepsPath)
	logDebug("got git root %s", gitRoot)

	entries := readGodepsFile(gitRoot, godepsPath)
	logDebug("got entries %v", entries)

	analyzeEntries(entries, gopath)

}

func goget(gopath, packagePath string) {
	logDebug("getting package %s", packagePath)
	cmd := exec.Command("go", "get", packagePath)
	cmd.Dir = path.Join(gopath, "src")
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, fmt.Sprintf("GOPATH=%s", gopath))
	logDebug("running command %v", *cmd)
	out, err := cmd.CombinedOutput()

	if err != nil {
		panic(fmt.Sprintf("failed to run go get for package %s.\nout: %v\nerr: %v", packagePath, string(out), err))
	}
}

func analyzeEntry(entry *GodepsEntry, gopath string) analyzeResult {

	r := analyzeResult{entry.Path, Uptodate, "", "", "", ""}
	packagePath := path.Join(gopath, "src", entry.Path)
	if !dirExists(packagePath) {
		goget(gopath, entry.Path)
	}
	if entry.GitType == Commit {
		// get commits
		commit, dateSummary := getLatestGitCommit(packagePath)
		r.NewCommitDateSummary = dateSummary
		r.NewCommitVersion = commit
		if entry.CommitVersion != r.NewCommitVersion {
			r.Status = Outdated
		}
	} else {
		// tags or branches
	}
	return r
}

func analyzeEntries(entries []*GodepsEntry, gopath string) {
	srcPath := path.Join(gopath, "src")
	if !dirExists(srcPath) {
		err := os.Mkdir(srcPath, 0777)
		if err != nil {
			panicWithMessage("failed to create dir %s. error: %v", srcPath, err)
		}
	}
	for _, entry := range entries {
		logDebug("analysing entry %v", *entry)
		result := analyzeEntry(entry, gopath)
		logInfo("package %s - got commit %s, committed at %s", entry.Path, result.NewCommitVersion, result.NewCommitDateSummary)
	}
}

func readGodepsFile(gitRoot, godepsPath string) []*GodepsEntry {
	entries := make([]*GodepsEntry, 0)
	contents := readFileContents(godepsPath)
	logDebug("got file contents %s", contents)
	lines := strings.Split(contents, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "git@") {
			logInfo("packages with @ in their paths aren't supported (yet). line: %s", line)
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) < 2 {
			continue
		}
		var gitRemote string
		if len(tokens) > 2 && strings.HasPrefix(tokens[2], "git.remote") {
			gitRemote = tokens[2]
		}
		entries = append(entries, NewGoDepsEntry(tokens[0], tokens[1], gitRemote))
	}
	return entries
}
