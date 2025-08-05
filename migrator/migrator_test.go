package migrator_test

import (
	"testing"

	"gihtub.com/ras0q/goalie/migrator"
	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

func TestCheck(t *testing.T) {
	analysistest.Run(t, testdata, migrator.GoalieAnalyzer, "a")
}

func TestFix(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, testdata, migrator.GoalieAnalyzer, "a")
}
