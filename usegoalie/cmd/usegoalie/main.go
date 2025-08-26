package main

import (
	"github.com/ras0q/goalie/usegoalie"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(usegoalie.Analyzer)
}
