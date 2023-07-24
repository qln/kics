package model

import (
	"github.com/Checkmarx/kics/internal/constants"
	"github.com/Checkmarx/kics/pkg/model"
)

// severitySonarQubeEquivalence maps the severity of the KICS to the SonarQube equivalent
var severitySonarQubeEquivalence = map[model.Severity]string{
	"INFO":   "INFO",
	"LOW":    "MINOR",
	"MEDIUM": "MAJOR",
	"HIGH":   "CRITICAL",
}

// categorySonarQubeEquivalence maps the category to the SonarQube equivalent
var categorySonarQubeEquivalence = map[string]string{
	"Access Control":          "VULNERABILITY",
	"Availability":            "VULNERABILITY",
	"Backup":                  "VULNERABILITY",
	"Best Practices":          "CODE_SMELL",
	"Build Process":           "VULNERABILITY",
	"Encryption":              "VULNERABILITY",
	"Insecure Configurations": "CODE_SMELL",
	"Insecure Defaults":       "CODE_SMELL",
	"Networking and Firewall": "VULNERABILITY",
	"Observability":           "VULNERABILITY",
	"Resource Management":     "VULNERABILITY",
	"Secret Management":       "VULNERABILITY",
	"Supply-Chain":            "VULNERABILITY",
	"Structure and Semantics": "CODE_SMELL",
}

// SonarQubeReportBuilder is the builder for the SonarQubeReport struct
type SonarQubeReportBuilder struct {
	version string
	report  *SonarQubeReport
}

// SonarQubeReport is a list of issues for SonarQube Report
type SonarQubeReport struct {
	Issues []Issue `json:"issues"`
}

// Issue is a single issue for SonarQube Report
type Issue struct {
	EngineID           string             `json:"engineId"`
	RuleID             string             `json:"ruleId"`
	Severity           string             `json:"severity"`
	Type               string             `json:"type"`
	PrimaryLocation    *Location          `json:"primaryLocation"`
	SecondaryLocations []*LocationWithRef `json:"secondaryLocations,omitempty"`
}

// Location is the location for the vulnerability in the SonarQube Report
type Location struct {
	Message   string `json:"message"`
	FilePath  string `json:"filePath"`
	TextRange *Range `json:"textRange"`
}

// LocationWithRef is the location for the vulnerability in the SonarQube Report with the Ref nfo
type LocationWithRef struct {
	Message            string `json:"message"`
	FilePath           string `json:"filePath"`
	TextRange          *Range `json:"textRange"`
	ReferenceFilePath  string `json:"referenceFilePath"`
	ReferenceTextRange *Range `json:"referenceTextRange"`
}

// Range is the range for the vulnerability in the SonarQube Report
type Range struct {
	StartLine int `json:"startLine"`
}

// NewSonarQubeRepory creates a new SonarQubeReportBuilder instance
func NewSonarQubeRepory() *SonarQubeReportBuilder {
	return &SonarQubeReportBuilder{
		version: "KICS " + constants.Version,
		report: &SonarQubeReport{
			Issues: make([]Issue, 0),
		},
	}
}

// BuildReport builds the SonarQubeReport from the given QueryResults
func (s *SonarQubeReportBuilder) BuildReport(summary *model.Summary) *SonarQubeReport {
	for i := range summary.Queries {
		s.buildIssue(&summary.Queries[i])
	}
	return s.report
}

// buildIssue builds the issue from the given QueryResult and adds it to the SonarQubeReport
func (s *SonarQubeReportBuilder) buildIssue(query *model.QueryResult) {
	issue := Issue{
		EngineID:           s.version,
		RuleID:             query.QueryID,
		Severity:           severitySonarQubeEquivalence[query.Severity],
		Type:               categorySonarQubeEquivalence[query.Category],
		PrimaryLocation:    buildLocation(0, query),
		SecondaryLocations: buildSecondaryLocation(query),
	}
	s.report.Issues = append(s.report.Issues, issue)
}

// buildSecondaryLocation builds the secondary location for the SonarQube Report
func buildSecondaryLocation(query *model.QueryResult) []*LocationWithRef {
	locations := make([]*LocationWithRef, 0)
	for i := range query.Files[1:] {
		locations = append(locations, buildLocationWithRef(i+1, query))
	}
	return locations
}

// buildLocationWithRef builds the location for the SonarQube Report
func buildLocationWithRef(index int, query *model.QueryResult) *LocationWithRef {
	message := query.Description
	if query.CISDescriptionID != "" {
		message = query.CISDescriptionID
	}
	location := &LocationWithRef{
		Message:  message,
		FilePath: query.Files[index].FileName,
		TextRange: &Range{
			StartLine: query.Files[index].Line,
		},
	}

	if query.Files[index].ReferenceLine != -1 {
		location.ReferenceFilePath = query.Files[index].ReferenceFileName
		location.ReferenceTextRange = &Range{
			StartLine: query.Files[index].ReferenceLine,
		}
	}
	return location
}

// buildLocation builds the location for the SonarQube Report
func buildLocation(index int, query *model.QueryResult) *Location {
	message := query.Description
	if query.CISDescriptionID != "" {
		message = query.CISDescriptionID
	}
	return &Location{
		Message:  message,
		FilePath: query.Files[index].FileName,
		TextRange: &Range{
			StartLine: query.Files[index].Line,
		},
	}
}
