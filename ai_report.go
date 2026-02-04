package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"
)

// AIRedditReport represents a daily AI signals report from Reddit sources
type AIRedditReport struct {
	ID                int                  `json:"id"`
	ReportDate        string               `json:"report_date"`
	DayStart          string               `json:"day_start"`
	DayEnd            string               `json:"day_end"`
	ReportContent     string               `json:"report_content"`
	ReportData        AIRedditReportData   `json:"report_data"`
	SignalsProcessed  int                  `json:"signals_processed"`
	TopSeverity       int                  `json:"top_severity"`
	SourceAnalysisIds []int                `json:"source_analysis_ids"`
	SubredditsCovered []string             `json:"subreddits_covered"`
	GenerationCostUSD float64              `json:"generation_cost_usd"`
	CreatedAt         string               `json:"created_at"`
}

// AIRedditReportData contains the structured analysis
type AIRedditReportData struct {
	ByCountry        map[string][]AIRedditSignal `json:"by_country"`
	BySubcategory    map[string][]AIRedditSignal `json:"by_subcategory"`
	WeekLabel        string                      `json:"week_label"`
	GeneratedAt      string                      `json:"generated_at"`
	TopSeverity      int                         `json:"top_severity"`
	TotalSignals     int                         `json:"total_signals"`
	SubredditSources []string                    `json:"subreddit_sources"`
}

// AIRedditSignal represents an individual signal from Reddit
type AIRedditSignal struct {
	ID          int     `json:"id"`
	Title       string  `json:"title"`
	Region      string  `json:"region"`
	Country     string  `json:"country"`
	Evidence    *string `json:"evidence"`
	Severity    int     `json:"severity"`
	Subreddit   string  `json:"subreddit"`
	Description string  `json:"description"`
	Subcategory string  `json:"subcategory"`
}

// FetchAIRedditReport fetches the latest AI Reddit report from the API
func FetchAIRedditReport(url string) (*AIRedditReport, error) {
	if url == "" {
		return nil, fmt.Errorf("AI_REPORT_URL not configured in .env")
	}

	if verbose {
		log.Printf("[FetchAIRedditReport] Fetching from: %s", url)
	}

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch AI Reddit report: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI Reddit report API returned status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var report AIRedditReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, fmt.Errorf("failed to parse AI Reddit report response: %v", err)
	}

	if verbose {
		log.Printf("[FetchAIRedditReport] Retrieved report dated %s with %d signals",
			report.ReportDate, report.ReportData.TotalSignals)
	}

	return &report, nil
}

// FormatAIRedditReportForScenario converts a report into scenario text for the simulation
func FormatAIRedditReportForScenario(report *AIRedditReport, minSeverity int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Current AI signals briefing as of %s:\n\n", report.ReportDate))

	// Collect and sort countries by highest severity signal
	type countryData struct {
		code        string
		signals     []AIRedditSignal
		maxSeverity int
	}

	countries := make([]countryData, 0)
	for code, signals := range report.ReportData.ByCountry {
		maxSev := 0
		filtered := make([]AIRedditSignal, 0)
		for _, s := range signals {
			if s.Severity >= minSeverity {
				filtered = append(filtered, s)
				if s.Severity > maxSev {
					maxSev = s.Severity
				}
			}
		}
		if len(filtered) > 0 {
			countries = append(countries, countryData{
				code:        code,
				signals:     filtered,
				maxSeverity: maxSev,
			})
		}
	}

	// Sort by severity (highest first)
	sort.Slice(countries, func(i, j int) bool {
		return countries[i].maxSeverity > countries[j].maxSeverity
	})

	// Format each country's signals
	for _, c := range countries {
		sb.WriteString(fmt.Sprintf("## %s (max severity: %d/10)\n", c.code, c.maxSeverity))
		for _, signal := range c.signals {
			sb.WriteString(fmt.Sprintf("- [Severity %d] %s: %s\n",
				signal.Severity, signal.Title, signal.Description))
		}
		sb.WriteString("\n")
	}

	if len(countries) == 0 {
		sb.WriteString("No significant signals above the severity threshold.\n")
	}

	return sb.String()
}

// FormatAIRedditReportNarrative returns the pre-formatted narrative from the report
func FormatAIRedditReportNarrative(report *AIRedditReport) string {
	return fmt.Sprintf("AI signals briefing as of %s:\n\n%s",
		report.ReportDate, report.ReportContent)
}
