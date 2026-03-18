package ui

import (
	"fmt"
	"io"
	"strings"
)

// Status represents the result of a check.
type Status int

const (
	StatusPass Status = iota
	StatusWarn
	StatusInfo
)

// Check is a single item in a report section.
type Check struct {
	Status Status
	Label  string
	Detail string
}

// Section is a titled group of checks.
type Section struct {
	Title  string
	Checks []Check
}

// Report collects sections and renders them as a formatted report.
type Report struct {
	title    string
	sections []Section
	noColor  bool
}

// NewReport creates a new report with the given title.
func NewReport(title string, noColor bool) *Report {
	return &Report{title: title, noColor: noColor}
}

// AddSection appends a section to the report.
func (r *Report) AddSection(s Section) {
	r.sections = append(r.sections, s)
}

// Render writes the formatted report to w.
func (r *Report) Render(w io.Writer) {
	green := "\033[32m"
	yellow := "\033[33m"
	dim := "\033[2m"
	bold := "\033[1m"
	reset := "\033[0m"
	if r.noColor {
		green = ""
		yellow = ""
		dim = ""
		bold = ""
		reset = ""
	}

	// Title
	_, _ = fmt.Fprintf(w, "\n%s%s%s\n", bold, r.title, reset)
	_, _ = fmt.Fprintf(w, "%s%s%s\n", dim, strings.Repeat("─", 40), reset)

	var passed, warned int

	for _, sec := range r.sections {
		_, _ = fmt.Fprintf(w, "%s%s%s\n", bold, sec.Title, reset)
		for _, c := range sec.Checks {
			var icon string
			switch c.Status {
			case StatusPass:
				icon = green + "✓" + reset
				passed++
			case StatusWarn:
				icon = yellow + "~" + reset
				warned++
			case StatusInfo:
				icon = dim + "-" + reset
			}

			detail := ""
			if c.Detail != "" {
				detail = " " + dim + c.Detail + reset
			}

			_, _ = fmt.Fprintf(w, "  %s %s%s\n", icon, c.Label, detail)
		}
		_, _ = fmt.Fprintln(w)
	}

	// Summary
	_, _ = fmt.Fprintf(w, "%s%d passed%s  %s%d warnings%s\n",
		green, passed, reset,
		yellow, warned, reset,
	)
}
