package parsers

import (
	"strings"

	"github.com/tomeryakir/gdau/utils"
)

type GPMParser struct {
	gitRoot string
	depPath string
	logger  *utils.Logger
}

func NewGPMParser(gitRoot, depPath string, logger *utils.Logger) *GPMParser {
	return &GPMParser{gitRoot, depPath, logger}
}

func (p *GPMParser) GitRoot() string {
	return p.gitRoot
}

func (p *GPMParser) DepPath() string {
	return p.depPath
}

func (p *GPMParser) ReadFile(gitRoot, godepsPath string) ([]*Entry, string, map[string]string, map[string]*Entry) {
	entries := make([]*Entry, 0)
	contents := utils.ReadFileContents(godepsPath, p.logger)
	p.logger.LogDebug("got file contents %s", contents)
	m := make(map[string]string)
	me := make(map[string]*Entry)
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
		entry := NewEntry(tokens[0], tokens[1], gitRemote)
		if strings.HasPrefix(line, "git@") {
			p.logger.LogInfo("packages with @ in their paths aren't supported (yet). line: %s", line)
			entry.IsSkipped = true
			entry.Summary = "packages with @ in their paths aren't supported (yet)"
		}
		m[tokens[0]] = line
		me[tokens[0]] = entry
		entries = append(entries, entry)
	}
	return entries, contents, m, me
}

func (p *GPMParser) UpdateFile(entries []*Entry, content string, contentMap map[string]string, entryMap map[string]*Entry) {
	needUpdate := false
	for _, entry := range entries {
		if !entry.IsUpdated {
			p.logger.LogInfo("updating entry %v", entry)
			old := contentMap[entry.Path]
			new := strings.Replace(old, entry.CommitVersion, entry.NewCommitVersion, 1)
			content = strings.Replace(content, old, new, 1)
			needUpdate = true
		}
	}
	if needUpdate {
		p.logger.LogDebug("content is now:ֿ\n%s", content)
		p.logger.LogInfo("Updating file")
		utils.WriteFile(p.DepPath(), content, p.logger)
	} else {
		p.logger.LogInfo("File already updated")
	}
}
