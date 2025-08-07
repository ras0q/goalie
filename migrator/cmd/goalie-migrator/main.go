package main

import (
	"github.com/ras0q/goalie/migrator"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(migrator.GoalieAnalyzer)
}
