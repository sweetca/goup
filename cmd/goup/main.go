// Copyright (c) 2020 Hervé Gouchet. All rights reserved.
// Use of this source code is governed by the MIT License
// that can be found in the LICENSE file.

package main

import (
	"context"
	"flag"
	"os"
	"time"

	"github.com/rvflash/goup/internal/app"
)

// Filled by the CI when building.
var buildVersion string

const timeout = 10 * time.Second

func main() {
	a := app.New(buildVersion)
	s := "exclude indirect modules"
	flag.BoolVar(&a.ExcludeIndirect, "i", false, s)
	s = "exit on first error occurred"
	flag.BoolVar(&a.Fast, "f", false, s)
	s = "ensure to have the latest major version"
	flag.BoolVar(&a.Major, "M", false, s)
	s = "ensure to have the latest couple major with minor version"
	flag.BoolVar(&a.MajorMinor, "m", false, s)
	s = "comma separated list of repositories (or part of), where forcing tag usage"
	flag.StringVar(&a.OnlyReleases, "r", "", s)
	s = "maximum time duration"
	flag.DurationVar(&a.Timeout, "t", timeout, s)
	s = "verbose output"
	flag.BoolVar(&a.Verbose, "v", false, s)
	s = "print version"
	flag.BoolVar(&a.Version, "V", false, s)
	flag.Parse()

	ctx := context.Background()
	if !a.Check(ctx, flag.Args()) {
		os.Exit(1)
	}
}
