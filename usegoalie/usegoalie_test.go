package usegoalie_test

import (
	"testing"

	"github.com/ras0q/goalie/usegoalie"
	"golang.org/x/tools/go/analysis/analysistest"
)

var testdata = analysistest.TestData()

func TestCheck(t *testing.T) {
	analysistest.Run(t, testdata, usegoalie.Analyzer, "a")
}

func TestFix(t *testing.T) {
	analysistest.RunWithSuggestedFixes(t, testdata, usegoalie.Analyzer, "a")
}
