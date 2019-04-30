package main

import (
	"flag"
	"fmt"
	"os"
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
	RemoteURL            string
	NewCommitVersion     string
	NewCommitDateSummary string
	DiffURL              string
	Summary              string
}

type EntryStatus int

const (
	Uptodate EntryStatus = 0
	Outdated EntryStatus = 1
	Skipped  EntryStatus = 2
	Problem  EntryStatus = 3
)

type EntryType int

const (
	Commit        EntryType = 0
	BranchVersion EntryType = 1
)

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
	if g.GitRemote != "" {
		g.RemoteURL = g.GitRemote
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

func analyzeEntry(entry *GodepsEntry, gopath string) {

	packagePath := path.Join(gopath, "src", entry.Path)
	if !dirExists(packagePath) {
		goget(gopath, entry.Path, packagePath, entry.GitRemote)
	}
	if entry.GitType == Commit {
		// get commits
		commit, dateSummary, err := getLatestGitCommit(packagePath)
		if err != nil {
			entry.Status = Problem
			entry.Summary = err.Error()
			return
		}
		entry.NewCommitDateSummary = dateSummary
		entry.NewCommitVersion = commit
		if entry.CommitVersion != entry.NewCommitVersion {
			entry.Status = Outdated
			summary, err := getCommitDiffSummary(packagePath, entry.CommitVersion, commit)
			if err != nil {
				entry.Status = Problem
				entry.Summary = err.Error()
				return
			}
			if entry.GitRemote != "" {
				url, err := getGitRemoteUrl(packagePath)
				if err != nil {
					entry.Status = Problem
					entry.Summary = err.Error()
					return
				}
				entry.RemoteURL = url
			}
			entry.Summary = summary
			entry.DiffURL = fmt.Sprintf("%s/compare/%s...%s", entry.RemoteURL, entry.CommitVersion, entry.NewCommitVersion)
		}
	} else {
		// tags or branches
		oldcommit, err := getCommitByTag(packagePath, entry.CommitVersion)
		if err != nil {
			entry.Status = Problem
			entry.Summary = err.Error()
			return
		}
		commit, tag, dateSummary, err := getLatestGitCommitByTag(packagePath)
		if err != nil {
			entry.Status = Problem
			entry.Summary = err.Error()
			return
		}
		entry.NewCommitDateSummary = dateSummary
		entry.NewCommitVersion = tag
		if entry.CommitVersion != entry.NewCommitVersion {
			entry.Status = Outdated
			summary, err := getCommitDiffSummary(packagePath, oldcommit, commit)
			if err != nil {
				entry.Status = Problem
				entry.Summary = err.Error()
				return
			}
			if entry.GitRemote != "" {
				url, err := getGitRemoteUrl(packagePath)
				if err != nil {
					entry.Status = Problem
					entry.Summary = err.Error()
					return
				}
				entry.RemoteURL = url
			}
			entry.Summary = summary
			entry.DiffURL = fmt.Sprintf("%s/compare/%s...%s", entry.RemoteURL, oldcommit, commit)
		}
	}
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
		if entry.Status == Skipped {
			continue
		}
		// not parallelising this for now as there may be multiple packages that use the same path
		logDebug("analysing entry %v", *entry)
		analyzeEntry(entry, gopath)
		logInfo("** package %s - data: %v", entry.Path, *entry)
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
		tokens := strings.Fields(line)
		if len(tokens) < 2 {
			continue
		}
		var gitRemote string
		if len(tokens) > 2 && strings.HasPrefix(tokens[2], "git.remote") {
			gitRemote = strings.Replace(tokens[2], "git.remote=", "", -1)
		}
		entry := NewGoDepsEntry(tokens[0], tokens[1], gitRemote)
		if strings.HasPrefix(line, "git@") {
			logInfo("packages with @ in their paths aren't supported (yet). line: %s", line)
			entry.Status = Skipped
			entry.Summary = "packages with @ in their paths aren't supported (yet)"
		}
		entries = append(entries, entry)
	}
	return entries
}
