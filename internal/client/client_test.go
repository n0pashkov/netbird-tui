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
