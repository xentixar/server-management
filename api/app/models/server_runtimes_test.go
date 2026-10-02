package models

import (
	"testing"

	"api/app/constants"
)

func TestSeedInstallingAndUpsert(t *testing.T) {
	r := SeedInstalling([]string{"8.2", "8.3"}, []string{"20", "22"}, []string{"redis"})
	if r.PHP.Default != "8.3" {
		t.Fatalf("php default = %q", r.PHP.Default)
	}
	if r.Node.Default != "22" {
		t.Fatalf("node default = %q", r.Node.Default)
	}
	if len(r.Addons) != 1 || r.Addons[0] != "redis" {
		t.Fatalf("addons = %v", r.Addons)
	}
	r.PHP.Upsert("8.3", constants.RuntimeStatusInstalled)
	v := r.PHP.Find("8.3")
	if v == nil || v.Status != constants.RuntimeStatusInstalled {
		t.Fatal("expected installed 8.3")
	}
}

func TestNormalizeVersionList(t *testing.T) {
	got, errMsg := NormalizeVersionList([]string{"8.3", "bad", "8.3"}, constants.IsAllowedPHPVersion)
	if errMsg != "" {
		t.Fatal(errMsg)
	}
	if len(got) != 1 || got[0] != "8.3" {
		t.Fatalf("got %v", got)
	}
	_, errMsg = NormalizeVersionList([]string{"bad"}, constants.IsAllowedPHPVersion)
	if errMsg == "" {
		t.Fatal("expected error")
	}
}

func TestAddonList(t *testing.T) {
	got := AddonList(map[string]bool{
		"redis":    true,
		"postgres": true,
		"mysql":    false,
	})
	if len(got) != 2 || got[0] != "redis" || got[1] != "postgres" {
		t.Fatalf("got %v", got)
	}
}
