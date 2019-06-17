package report

import (
	"os"
	"os/exec"
	"text/template"

	dep "github.com/tomeryakir/gdau/parsers"
)

const (
	reportFile = "report.html"
)

type reportData struct {
	UptodatePackages int
	OutdatedPackages int
	SkippedPackages  int
	ProblemPackages  int
	Entries          []*dep.Entry
}

func GenerateReportFile(entries []*dep.Entry) error {
	data := reportData{0, 0, 0, 0, entries}
	for _, entry := range entries {
		if entry.IsUpdated {
			data.UptodatePackages++
		} else {
			data.OutdatedPackages++
		}
		if entry.IsSkipped {
			data.SkippedPackages++
		}
		if entry.IsProblem {
			data.ProblemPackages++
		}
	}

	tmpl := template.Must(template.New("dependencies").Parse(string(GetHtmlTemplateBinData())))
	f, err := os.Create(reportFile)
	if err != nil {
		return err
	}
	tmpl.Execute(f, data)
	return nil

}

func OpenReportFile() {
	exec.Command("open", reportFile).Run()
}
