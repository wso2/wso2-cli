package updatechk

import "testing"

// Version comparison drives the "Update available" banner, so a wrong answer
// either nags every user forever or never tells anyone about a release.
//
// The pre-v1.0.0 tag scheme concatenated a version and a timestamp with no
// separator ("v1.2.21" + "2606121500" -> "v1.2.212606121500"), which parses as
// patch=212606121500 and therefore outranks any sane v1.x. The cases below pin
// the behaviour that matters for the v1.0.0 reset.
func TestIsCurrentVersionOlder(t *testing.T) {
	cases := []struct {
		name        string
		current     string
		latest      string
		wantIsOlder bool
	}{
		{"same version is not older", "v1.0.0", "v1.0.0", false},
		{"patch behind", "v1.0.0", "v1.0.1", true},
		{"minor behind", "v1.0.0", "v1.1.0", true},
		{"major behind", "v1.0.0", "v2.0.0", true},
		{"ahead of latest", "v1.0.1", "v1.0.0", false},
		{"major ahead", "v2.0.0", "v1.9.9", false},
		{"v prefix optional on both sides", "1.0.0", "1.0.0", false},

		// Dev builds carry a git-describe + date suffix; the suffix must not
		// make an up-to-date build look stale.
		{"dev suffix on current is ignored", "v1.0.0-3-gabc1234-20260729", "v1.0.0", false},
		{"dev suffix on current, real update", "v1.0.0-3-gabc1234-20260729", "v1.0.2", true},

		// The legacy concatenated-timestamp tags outrank v1.x numerically.
		// This is why the version reset had to come with fresh history rather
		// than a v1.0.0 tag alongside the old ones.
		{"legacy concatenated tag outranks v1.0.0", "v1.0.0", "v1.2.212606121500", true},
		{"v2 beats legacy concatenated tag", "v2.0.0", "v1.2.212606121500", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := isCurrentVersionOlder(tc.current, tc.latest); got != tc.wantIsOlder {
				t.Errorf("isCurrentVersionOlder(%q, %q) = %v, want %v",
					tc.current, tc.latest, got, tc.wantIsOlder)
			}
		})
	}
}
