package generate

import (
	"testing"

	"github.com/alecthomas/assert"
)

func TestDetermineBumpStrategy(t *testing.T) {
	tests := map[string]struct {
		SourceBranch    string
		DestBranch      string
		Bump            string
		ExpectedMethod  string
		ExpectedVersion string
	}{
		"source branch bugfix, dest branch develop and auto bump": {
			SourceBranch:    "bugfix/some",
			DestBranch:      "develop",
			Bump:            "auto",
			ExpectedMethod:  "build",
			ExpectedVersion: "patch",
		},
		"source branch feature, dest branch develop and auto bump": {
			SourceBranch:    "feature/some",
			DestBranch:      "develop",
			Bump:            "auto",
			ExpectedMethod:  "build",
			ExpectedVersion: "minor",
		},
		"source branch major, dest branch develop and auto bump": {
			SourceBranch:    "major/some",
			DestBranch:      "develop",
			Bump:            "auto",
			ExpectedMethod:  "build",
			ExpectedVersion: "major",
		},
		"source branch hotfix, dest branch master and auto bump": {
			SourceBranch:   "hotfix/some",
			DestBranch:     "master",
			Bump:           "auto",
			ExpectedMethod: "hotfix",
		},
		"source branch develop, dest branch master and auto bump": {
			SourceBranch:   "develop",
			DestBranch:     "master",
			Bump:           "auto",
			ExpectedMethod: "final",
		},
		"not a valid source branch prefix and auto bump": {
			SourceBranch:   "some-branch",
			Bump:           "auto",
			ExpectedMethod: "build",
		},
		"patch bump": {
			Bump:           "patch",
			ExpectedMethod: "patch",
		},
		"minor bump": {
			Bump:           "minor",
			ExpectedMethod: "minor",
		},
		"major bump": {
			Bump:           "major",
			ExpectedMethod: "major",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			method, version := determineBumpStrategy(test.Bump, test.SourceBranch, test.DestBranch, "master", "develop")

			assert.Equal(t, test.ExpectedMethod, method)
			assert.Equal(t, test.ExpectedVersion, version)
		})
	}
}
