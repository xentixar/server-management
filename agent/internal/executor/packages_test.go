package executor

import "testing"

func TestPHPPackagesAllowlist(t *testing.T) {
	pkgs, err := phpPackages("8.3")
	if err != nil {
		t.Fatal(err)
	}
	if len(pkgs) == 0 {
		t.Fatal("expected php packages")
	}
	want := "php8.3-fpm"
	found := false
	for _, p := range pkgs {
		if p == want {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing %s in %v", want, pkgs)
	}

	if _, err := phpPackages("7.4"); err == nil {
		t.Fatal("expected unsupported php version error")
	}
}

func TestNodeVersionAllowlist(t *testing.T) {
	if err := validateNodeVersion("20"); err != nil {
		t.Fatal(err)
	}
	if err := validateNodeVersion("16"); err == nil {
		t.Fatal("expected unsupported node version error")
	}
}

func TestRandomPassword(t *testing.T) {
	a, err := randomPassword(24)
	if err != nil {
		t.Fatal(err)
	}
	b, err := randomPassword(24)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 24 || len(b) != 24 {
		t.Fatalf("length: %d %d", len(a), len(b))
	}
	if a == b {
		t.Fatal("expected distinct passwords")
	}
}

func TestAddonPackageMap(t *testing.T) {
	for _, name := range []string{"redis", "mysql", "mariadb", "postgres"} {
		pkgs := addonPackages[name]
		if len(pkgs) == 0 {
			t.Fatalf("addon %s has no packages", name)
		}
	}
	if len(basePackages) == 0 {
		t.Fatal("base packages empty")
	}
}

func TestPayloadHelpers(t *testing.T) {
	payload := map[string]any{
		"php_versions": []any{"8.2", "8.3"},
		"addons": map[string]any{
			"redis":   true,
			"mariadb": false,
		},
	}
	got := payloadStringSlice(payload, "php_versions")
	if len(got) != 2 || got[0] != "8.2" {
		t.Fatalf("unexpected php_versions %v", got)
	}
	flags := payloadBoolMap(payload, "addons")
	if !flags["redis"] || flags["mariadb"] {
		t.Fatalf("unexpected addons %v", flags)
	}
}
