package parsers

import (
	"encoding/hex"
	"io/ioutil"
	"strings"

	"github.com/tomeryakir/gdau/utils"
)

type Entry struct {
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

func NewEntry(path, commitVersion, gitRemote string) *Entry {
	g := &Entry{}
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

func isHexString(s string) bool {
	_, err := hex.DecodeString(s)
	return err == nil
}

// Parser interface - for reading dependency files
type Parser interface {
	ReadFile(string, string) ([]*Entry, string, map[string]string)
	GitRoot() string
	DepPath() string
}

func UpdateDependencyFile(entries []*Entry, godepsPath, content string, contentMap map[string]string, logger *utils.Logger) {
	needUpdate := false
	for _, entry := range entries {
		if !entry.IsUpdated {
			logger.LogDebug("entry %v is outdated", entry)
			old := contentMap[entry.Path]
			new := strings.Replace(old, entry.CommitVersion, entry.NewCommitVersion, 1)
			content = strings.Replace(content, old, new, 1)
			needUpdate = true
		}
	}
	if needUpdate {
		logger.LogDebug("content is now:ֿ\n%s", content)
		if err := ioutil.WriteFile(godepsPath, []byte(content), 0644); err != nil {
			logger.PanicWithMessage("failed to update godeps file. Error: %v", err)
		}
	} else {
		logger.LogInfo("File already updated")
	}
}

func ReadDependencyFile(p Parser) ([]*Entry, string, map[string]string) {
	return p.ReadFile(p.GitRoot(), p.DepPath())
}
