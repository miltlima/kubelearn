package utils

import (
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
)

type Result struct {
	TestName   string
	Passed     bool
	Difficulty string
	// Feedback explains why the question is not solved yet. Empty when passed.
	Feedback string `json:"Feedback,omitempty"`
}

func RenderResultsTable(results []Result) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"KubeLearn - Test your knowledge of Kubernetes v0.2.1", "Result", "Difficulty"})
	table.SetAutoWrapText(false)

	for _, result := range results {
		passedStr := color.GreenString("✅ Pass")
		if !result.Passed {
			passedStr = color.RedString("🆘 Fail")
		}
		row := []string{result.TestName, passedStr, result.Difficulty}
		table.Append(row)
	}

	table.Render()
}
