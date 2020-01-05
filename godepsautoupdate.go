package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	git "github.com/tomeryakir/gdau/gitutils"
	dep "github.com/tomeryakir/gdau/parsers"
	"github.com/tomeryakir/gdau/report"
	"github.com/tomeryakir/gdau/utils"
)

func main() {
	var depsPath string
	var gopath string
	var tipe string
	var updateFile bool
	var debug bool

	flag.StringVar(&depsPath, "path", "", "path to dependency file")
	flag.StringVar(&gopath, "gopath", "", "path to packages root")
	flag.StringVar(&tipe, "deptype", "gpm", "type of dependency file (can be gpm, dep, module)")
	flag.BoolVar(&debug, "debug", false, "turn on debug")
	flag.BoolVar(&updateFile, "updateFile", false, "update the dependency file")
	flag.Parse()

	logger := utils.NewLogger(debug)

	if depsPath == "" {
		flag.Usage()
		panic("dependency path wasn't specified")
	}
	if gopath == "" {
		flag.Usage()
		panic("Gopath wasn't specified")
	}
	gitRoot := git.GetGitRoot(depsPath, logger)
	logger.LogDebug("got git root %s", gitRoot)

	var parser dep.Parser
	switch tipe {
	case "gpm":
		parser = dep.NewGPMParser(gitRoot, depsPath, logger)
	case "dep":
		parser = dep.NewGopkgParser(gitRoot, depsPath, logger)
	default:
		logger.PanicWithMessage("unsupported dependency format %s", tipe)
	}

	entries, content, contentMap, entryMap := dep.ReadDependencyFile(parser)
	logger.LogDebug("got entries %+v", entries)

	analyzeEntries(entries, gopath, logger)

	err := report.GenerateReportFile(entries)
	if err != nil {
		logger.PanicWithMessage("failed to generate the report file. error: %v", err)
	}
	report.OpenReportFile()

	if updateFile {
		dep.UpdateDependencyFile(parser, entries, content, contentMap, entryMap)
	}

}

func runCommand(cmd, cmdDir, cmdArgs string, logger *utils.Logger) (string, error) {
	cmdArgsSplit := strings.Fields(cmdArgs)
	c := exec.Command(cmd, cmdArgsSplit...)
	c.Dir = cmdDir
	c.Env = os.Environ()
	logger.LogDebug("running command %v", *c)
	out, err := c.CombinedOutput()
	logger.LogDebug("got output %s", string(out))
	return string(out), err
}

func analyzeEntry(entry *dep.Entry, gopath string, logger *utils.Logger) {
	logger.LogInfo("analyzing package %s", entry.Path)
	packagePath := path.Join(gopath, "src", entry.Path)
	if utils.DirExists(packagePath) && entry.GitRemote == "" {
		url, err := git.GetGitRemoteURL(packagePath, logger)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		entry.RemoteURL = url
	}
	if !utils.DirExists(packagePath) {
		if err := git.Goget(gopath, entry.Path, packagePath, entry.GitRemote, logger); err != nil {
			entry.Summary = err.Error()
			entry.IsProblem = true
			return
		}
	} else {
		if err := git.AddRemote(entry.Path, entry.GitRemote, packagePath, logger); err != nil {
			entry.Summary = err.Error()
			entry.IsProblem = true
			return
		}
		git.Gitpull(packagePath, logger)
	}

	entry.ReleasesURL = fmt.Sprintf("%s/releases", strings.TrimSuffix(entry.RemoteURL, ".git"))
	if entry.GitType == dep.Commit {
		// get commits
		commit, dateSummary, err := git.GetLatestGitCommit(packagePath, logger)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		entry.NewCommitDateSummary = dateSummary
		entry.NewCommitVersion = commit
		if entry.CommitVersion != entry.NewCommitVersion {
			entry.IsUpdated = false
			summary, err := git.GetCommitDiffSummary(packagePath, entry.CommitVersion, commit, logger)
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
		oldcommit, err := git.GetCommitByTag(packagePath, entry.CommitVersion, logger)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		commit, tag, dateSummary, err := git.GetLatestGitCommitByTag(packagePath, logger)
		if err != nil {
			entry.IsProblem = true
			entry.Summary = err.Error()
			return
		}
		entry.NewCommitDateSummary = dateSummary
		entry.NewCommitVersion = tag
		if entry.CommitVersion != entry.NewCommitVersion {
			entry.IsUpdated = false
			summary, err := git.GetCommitDiffSummary(packagePath, oldcommit, commit, logger)
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

func analyzeEntries(entries []*dep.Entry, gopath string, logger *utils.Logger) {
	srcPath := path.Join(gopath, "src")
	if !utils.DirExists(srcPath) {
		err := os.Mkdir(srcPath, 0777)
		if err != nil {
			logger.PanicWithMessage("failed to create dir %s. error: %v", srcPath, err)
		}
	}
	for _, entry := range entries {
		if entry.IsSkipped {
			continue
		}
		// not parallelising this for now as there may be multiple packages that use the same path
		logger.LogDebug("analysing entry %v", *entry)
		analyzeEntry(entry, gopath, logger)
		logger.LogDebug("** package %s - data: %v", entry.Path, *entry)
	}
}
