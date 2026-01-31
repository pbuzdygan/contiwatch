package server

import "testing"

func TestCompareVersionsSemver(t *testing.T) {
	if compareVersions("1.2.3", "1.3.0") >= 0 {
		t.Fatalf("expected 1.2.3 < 1.3.0")
	}
	if compareVersions("2.0.0", "1.9.9") <= 0 {
		t.Fatalf("expected 2.0.0 > 1.9.9")
	}
}

func TestCompareVersionsDevSemver(t *testing.T) {
	if compareVersions("dev0.5.13", "dev0.6.0") >= 0 {
		t.Fatalf("expected dev0.5.13 < dev0.6.0")
	}
	if compareVersions("dev1.2.0", "dev1.2.0") != 0 {
		t.Fatalf("expected dev1.2.0 == dev1.2.0")
	}
}

func TestCompareVersionsDevNumeric(t *testing.T) {
	if compareVersions("dev12", "dev2") <= 0 {
		t.Fatalf("expected dev12 > dev2")
	}
}

func TestCompareVersionsDevVsMain(t *testing.T) {
	if compareVersions("dev12", "1.0.0") >= 0 {
		t.Fatalf("expected dev12 < 1.0.0")
	}
}

func TestSelectReleaseForChannel(t *testing.T) {
	releases := []githubRelease{
		{TagName: "v1.1.0", TargetCommitish: "main"},
		{TagName: "dev12", TargetCommitish: "dev"},
		{TagName: "v1.2.0", TargetCommitish: "main"},
	}
	latestMain := selectReleaseForChannel(releases, "main")
	if latestMain == nil || latestMain.TagName != "v1.2.0" {
		t.Fatalf("expected main latest v1.2.0, got %+v", latestMain)
	}
	latestDev := selectReleaseForChannel(releases, "dev")
	if latestDev == nil || latestDev.TagName != "dev12" {
		t.Fatalf("expected dev latest dev12, got %+v", latestDev)
	}
}
