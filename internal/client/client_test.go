package client

import "testing"

func TestParseProfileListOutputTableFormat(t *testing.T) {
	profiles := parseProfileListOutput("NAME     ACTIVE\ndefault  ✓\nother\n")

	if len(profiles) != 2 {
		t.Fatalf("len = %d, want 2", len(profiles))
	}
	if profiles[0].Name != "default" || !profiles[0].IsActive {
		t.Fatalf("first profile = %#v, want active default", profiles[0])
	}
	if profiles[1].Name != "other" || profiles[1].IsActive {
		t.Fatalf("second profile = %#v, want inactive other", profiles[1])
	}
}

func TestParseProfileListOutputLegacyFormat(t *testing.T) {
	profiles := parseProfileListOutput("Found 2 profiles:\n✕ default\n✓ shaban\n")

	if len(profiles) != 2 {
		t.Fatalf("len = %d, want 2", len(profiles))
	}
	if profiles[0].Name != "default" || profiles[0].IsActive {
		t.Fatalf("first profile = %#v, want inactive default", profiles[0])
	}
	if profiles[1].Name != "shaban" || !profiles[1].IsActive {
		t.Fatalf("second profile = %#v, want active shaban", profiles[1])
	}
}

func TestCompareVersions(t *testing.T) {
	for _, tc := range []struct {
		name string
		a    string
		b    string
		want int
	}{
		{name: "same", a: "0.73.2", b: "v0.73.2", want: 0},
		{name: "newer latest", a: "0.72.3", b: "0.73.2", want: -1},
		{name: "current newer", a: "0.74.0", b: "0.73.2", want: 1},
		{name: "numeric not lexical", a: "0.9.0", b: "0.10.0", want: -1},
		{name: "suffix", a: "0.73.2-dev", b: "0.73.2", want: 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := compareVersions(tc.a, tc.b)
			if got != tc.want {
				t.Fatalf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
