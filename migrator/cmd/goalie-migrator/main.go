package main

import (
	"gihtub.com/ras0q/goalie/migrator"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(migrator.GoalieAnalyzer)
}
