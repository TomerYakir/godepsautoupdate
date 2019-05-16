package main

import (
	"flag"
	"fmt"
	"io/ioutil"
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
	IsUpdated            bool
	IsSkipped            bool
	IsProblem            bool
	RemoteURL            string
	ReleasesURL          string
	NewCommitVersion     string
	NewCommitDateSummary string
	DiffURL              string
	Summary              string
}

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
	g.IsUpdated = true
	g.IsSkipped = false
	g.IsProblem = false
	return g
}

func main() {
	var godepsPath string
	var gopath string
	var updateFile bool
	var installDepsCommand string
	var installDepsCommandDir string
	var installDepsCommandArgs string
	var buildCommand string
	var buildCommandDir string
	var buildCommandArgs string

	flag.StringVar(&godepsPath, "path", "", "path to godeps")
	flag.StringVar(&gopath, "gopath", "", "path to packages root")
	flag.BoolVar(&debug, "debug", false, "turn on debug")
	flag.BoolVar(&updateFile, "updateFile", false, "update the Godeps file")
	flag.StringVar(&installDepsCommand, "installDepsCommand", "", "path to install deps command")
	flag.StringVar(&installDepsCommandDir, "installDepsCommandDir", "", "path to install deps working dir")
	flag.StringVar(&installDepsCommandArgs, "installDepsCommandArgs", "", "args for install deps command")
	flag.StringVar(&buildCommand, "buildCommand", "", "path to build command")
	flag.StringVar(&buildCommandDir, "buildCommandDir", "", "path to build source")
	flag.StringVar(&buildCommandArgs, "buildCommandArgs", "", "args for build command")
	flag.Parse()

	if godepsPath == "" {
		flag.Usage()
		panic("Godeps path wasn't specified")
	}
	if gopath == "" {
		flag.Usage()
		panic("Gopath wasn't specified")
	}
	gitRoot := getGitRoot(godepsPath)
	logDebug("got git root %s", gitRoot)

	entries, content, contentMap := readGodepsFile(gitRoot, godepsPath)
	logDebug("got entries %+v", entries)

	analyzeEntries(entries, gopath)

	err := generateReportFile(entries)
	if err != nil {
		panicWithMessage("failed to generate the report file. error: %v", err)
	}
	openReportFile()

	if updateFile {
		updateGodepsFile(entries, godepsPath, content, contentMap)
	}

	if buildCommand != "" {
		build(installDepsCommand, installDepsCommandDir, installDepsCommandArgs, buildCommand, buildCommandDir, buildCommandArgs)
	}

}

func build(installDepsCommand, installDepsCommandDir, installDepsCommandArgs, buildCommand, buildCommandDir, buildCommandArgs string) bool {
	// install deps first
	logInfo("installing dependencies... (this may take a while)")
	out, err := runCommand(installDepsCommand, installDepsCommandDir, installDepsCommandArgs)
	if err != nil {
		logInfo("Failed to run the install deps command. Error was: %v\nOutput: %v", err, out)
		return false
	}
	// try build
	logInfo("building... (this may take a while)")
	out, err = runCommand(buildCommand, buildCommandDir, buildCommandArgs)
	if err != nil {
		logInfo("Failed to run the build command. Error was: %v\nOutput: %v", err, out)
		return false
	}
	return true
}

func runCommand(cmd, cmdDir, cmdArgs string) (string, error) {
	cmdArgsSplit := strings.Fields(cmdArgs)
	c := exec.Command(cmd, cmdArgsSplit...)
	c.Dir = cmdDir
	c.Env = os.Environ()
	logDebug("running command %v", *c)
	out, err := c.CombinedOutput()
	logDebug("got output %s", string(out))
	return string(out), err
}

func updateGodepsFile(entries []*GodepsEntry, godepsPath, content string, contentMap map[string]string) {
	needUpdate := false
	for _, entry := range entries {
		if !entry.IsUpdated {
			logDebug("entry %v is outdated", entry)
			old := contentMap[entry.Path]
			new := strings.Replace(old, entry.CommitVersion, entry.NewCommitVersion, 1)
			content = strings.Replace(content, old, new, 1)
			needUpdate = true
		}
	}
	if needUpdate {
		logDebug("content is now:ֿ\n%s", content)
		if err := ioutil.WriteFile(godepsPath, []byte(content), 0644); err != nil {
			panicWithMessage("failed to update godeps file. Error: %v", err)
		}
	} else {
		logInfo("File already updated")
	}
}

func analyzeEntry(entry *GodepsEntry, gopath string) {
	logInfo("analyzing package %s", entry.Path)
	packagePath := path.Join(gopath, "src", entry.Path)
	if !dirExists(packagePath) {
		goget(gopath, entry.Path, packagePath, entry.GitRemote)
	} else {
		addRemote(entry.Path, entry.GitRemote, packagePath)
		gitpull(packagePath)
	}
	if entry.GitRemote == "" {
		url, err := getGitRemoteUrl(packagePath)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		entry.RemoteURL = url
	}
	entry.ReleasesURL = fmt.Sprintf("%s/releases", strings.TrimSuffix(entry.RemoteURL, ".git"))
	if entry.GitType == Commit {
		// get commits
		commit, dateSummary, err := getLatestGitCommit(packagePath)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		entry.NewCommitDateSummary = dateSummary
		entry.NewCommitVersion = commit
		if entry.CommitVersion != entry.NewCommitVersion {
			entry.IsUpdated = false
			summary, err := getCommitDiffSummary(packagePath, entry.CommitVersion, commit)
			if err != nil {
				entry.IsProblem = true
				entry.Summary = err.Error()
				return
			}
			entry.Summary = summary
			entry.DiffURL = fmt.Sprintf("%s/compare/%s...%s", entry.RemoteURL, entry.CommitVersion, entry.NewCommitVersion)
		}
	} else {
		// tags or branches
		oldcommit, err := getCommitByTag(packagePath, entry.CommitVersion)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		commit, tag, dateSummary, err := getLatestGitCommitByTag(packagePath)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		entry.NewCommitDateSummary = dateSummary
		entry.NewCommitVersion = tag
		if entry.CommitVersion != entry.NewCommitVersion {
			entry.IsUpdated = false
			summary, err := getCommitDiffSummary(packagePath, oldcommit, commit)
			if err != nil {
				entry.IsProblem = true
				entry.Summary = err.Error()
				return
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
		if entry.IsSkipped {
			continue
		}
		// not parallelising this for now as there may be multiple packages that use the same path
		logDebug("analysing entry %v", *entry)
		analyzeEntry(entry, gopath)
		logDebug("** package %s - data: %v", entry.Path, *entry)
	}
}

func readGodepsFile(gitRoot, godepsPath string) ([]*GodepsEntry, string, map[string]string) {
	entries := make([]*GodepsEntry, 0)
	contents := readFileContents(godepsPath)
	logDebug("got file contents %s", contents)
	m := make(map[string]string)
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
			entry.IsSkipped = true
			entry.Summary = "packages with @ in their paths aren't supported (yet)"
		}
		m[tokens[0]] = line
		entries = append(entries, entry)
	}
	return entries, contents, m
}
