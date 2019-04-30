package main

import (
	"os"
	"os/exec"
	"text/template"
)

const (
	reportTemplate = "reportTemplate.html"
	reportFile     = "report.html"
)

type reportData struct {
	UptodatePackages int
	OutdatedPackages int
	SkippedPackages  int
	ProblemPackages  int
	Entries          []*GodepsEntry
}

func generateReportFile(entries []*GodepsEntry) error {
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
	tmpl := template.Must(template.ParseFiles(reportTemplate))
	f, err := os.Create(reportFile)
	if err != nil {
		return err
	}
	tmpl.Execute(f, data)
	return nil

}

func openReportFile() {
	exec.Command("open", reportFile).Run()
}
