package parsers

import (
	"encoding/hex"
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
	// ReadFile - get gitroot and dep path; return slice of Entries, file content and map of entries with lines
	ReadFile(string, string) ([]*Entry, string, map[string]string, map[string]*Entry)

	// UpdateFile - get slice of processed entries, raw content, map of entries with lines, map of entry path with Entry
	UpdateFile([]*Entry, string, map[string]string, map[string]*Entry)
	GitRoot() string
	DepPath() string
}

func ReadDependencyFile(p Parser) ([]*Entry, string, map[string]string, map[string]*Entry) {
	return p.ReadFile(p.GitRoot(), p.DepPath())
}

func UpdateDependencyFile(p Parser, entries []*Entry, content string, contentMap map[string]string, entryMap map[string]*Entry) {
	p.UpdateFile(entries, content, contentMap, entryMap)
}
