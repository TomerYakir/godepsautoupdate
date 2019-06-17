package parsers

import (
	"strings"

	"github.com/tomeryakir/gdau/utils"
)

type GopkgParser struct {
	gitRoot string
	depPath string
	logger  *utils.Logger
}

func NewGopkgParser(gitRoot, depPath string, logger *utils.Logger) *GopkgParser {
	return &GopkgParser{gitRoot, depPath, logger}
}

func (p *GopkgParser) GitRoot() string {
	return p.gitRoot
}

func (p *GopkgParser) DepPath() string {
	return p.depPath
}

func (p *GopkgParser) ReadFile(gitRoot, godepsPath string) ([]*Entry, string, map[string]string, map[string]*Entry) {
	entries := make([]*Entry, 0)
	contents := utils.ReadFileContents(godepsPath, p.logger)
	p.logger.LogDebug("got file contents %s", contents)
	m := make(map[string]string)
	me := make(map[string]*Entry)
	lines := strings.Split(contents, "\n")
	var currentEntry = &Entry{
		IsUpdated: true,
	}
	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) < 3 {
			continue
		}
		if tokens[0] == "name" {
			if currentEntry.Path != "" {
				entries = append(entries, currentEntry)
			}
			currentEntry = &Entry{IsUpdated: true}
			currentEntry.Path = utils.ClearQuotes(tokens[2])
			me[utils.ClearQuotes(tokens[2])] = currentEntry
		}
		if tokens[0] == "source" {
			if strings.HasPrefix(utils.ClearQuotes(tokens[2]), "git@") {
				p.logger.LogInfo("packages with @ in their paths aren't supported (yet). line: %s", line)
				currentEntry.IsSkipped = true
				currentEntry.Summary = "packages with @ in their paths aren't supported (yet)"
			} else {
				currentEntry.GitRemote = utils.ClearQuotes(tokens[2])
			}
		}
		if tokens[0] == "revision" {
			currentEntry.GitType = Commit
			currentEntry.CommitVersion = utils.ClearQuotes(tokens[2])
		}
		if tokens[0] == "version" {
			currentEntry.GitType = BranchVersion
			currentEntry.CommitVersion = utils.ClearQuotes(tokens[2])
		}
	}
	me[currentEntry.Path] = currentEntry
	entries = append(entries, currentEntry)
	return entries, contents, m, me
}

func (p *GopkgParser) UpdateFile(entries []*Entry, content string, contentMap map[string]string, entryMap map[string]*Entry) {
	needUpdate := false
	shouldUpdateEntry := false
	newContent := ""
	var currentEntry *Entry
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		if strings.HasPrefix(line, "#") {
			newContent += line + "\n"
			continue
		}
		tokens := strings.Fields(line)
		if len(tokens) < 3 {
			newContent += line + "\n"
			continue
		}
		if tokens[0] == "name" {
			if entry, ok := entryMap[utils.ClearQuotes(tokens[2])]; ok {
				if !entry.IsUpdated && !entry.IsProblem && !entry.IsSkipped {
					shouldUpdateEntry = true
					currentEntry = entry
					needUpdate = true
				}
			}
			newContent += line + "\n"
		} else if (tokens[0] == "version" || tokens[0] == "revision") && shouldUpdateEntry {
			newLine := strings.Replace(line, currentEntry.CommitVersion, currentEntry.NewCommitVersion, -1)
			newContent += newLine + "\n"
			shouldUpdateEntry = false
		} else {
			newContent += line + "\n"
		}
	}
	if needUpdate {
		p.logger.LogDebug("content is now:ֿ\n%s", newContent)
		p.logger.LogInfo("Updating file")
		utils.WriteFile(p.DepPath(), newContent, p.logger)
	} else {
		p.logger.LogInfo("File already updated")
	}
}
