package updatecheck

import "testing"

func TestIsNewer(t *testing.T) {
	cases := []struct {
		latest  string
		current string
		want    bool
	}{
		{"v0.1.4", "v0.1.3", true},
		{"0.2.0", "v0.1.9", true},
		{"v0.1.3", "v0.1.3", false},
		{"v0.1.2", "v0.1.3", false},
		{"dev", "v0.1.3", false},
	}
	for _, tc := range cases {
		if got := IsNewer(tc.latest, tc.current); got != tc.want {
			t.Fatalf("IsNewer(%q, %q) = %v, want %v", tc.latest, tc.current, got, tc.want)
		}
	}
}

func TestNoticeFromReleaseSkipsConfiguredVersion(t *testing.T) {
	got := noticeFromRelease("v0.1.4", "https://example.com", "v0.1.3", []string{"0.1.4"})
	if got.Available {
		t.Fatalf("notice should not be available for skipped version: %+v", got)
	}
}

func TestNoticeFromReleaseReportsNewerVersion(t *testing.T) {
	got := noticeFromRelease("v0.1.4", "https://example.com", "v0.1.3", nil)
	if !got.Available || got.LatestVersion != "v0.1.4" || got.URL == "" {
		t.Fatalf("notice = %+v", got)
	}
}
