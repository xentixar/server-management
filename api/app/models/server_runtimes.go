package models

import (
	"sort"
	"strings"

	"api/app/constants"
)

type RuntimeVersion struct {
	Version string `json:"version"`
	Status  string `json:"status"`
}

type RuntimeFamily struct {
	Default  string           `json:"default"`
	Versions []RuntimeVersion `json:"versions"`
}

type ServerRuntimes struct {
	PHP    RuntimeFamily `json:"php"`
	Node   RuntimeFamily `json:"node"`
	Addons []string      `json:"addons"`
}

func EmptyServerRuntimes() ServerRuntimes {
	return ServerRuntimes{
		PHP:    RuntimeFamily{Versions: []RuntimeVersion{}},
		Node:   RuntimeFamily{Versions: []RuntimeVersion{}},
		Addons: []string{},
	}
}

func (r *ServerRuntimes) Normalize() {
	if r.PHP.Versions == nil {
		r.PHP.Versions = []RuntimeVersion{}
	}
	if r.Node.Versions == nil {
		r.Node.Versions = []RuntimeVersion{}
	}
	if r.Addons == nil {
		r.Addons = []string{}
	}
}

func (r *ServerRuntimes) Family(runtime string) *RuntimeFamily {
	switch runtime {
	case constants.RuntimePHP:
		return &r.PHP
	case constants.RuntimeNode:
		return &r.Node
	default:
		return nil
	}
}

func (f *RuntimeFamily) Upsert(version, status string) {
	for i := range f.Versions {
		if f.Versions[i].Version == version {
			f.Versions[i].Status = status
			return
		}
	}
	f.Versions = append(f.Versions, RuntimeVersion{Version: version, Status: status})
	sort.Slice(f.Versions, func(i, j int) bool {
		return f.Versions[i].Version < f.Versions[j].Version
	})
}

func (f *RuntimeFamily) SetStatus(version, status string) bool {
	for i := range f.Versions {
		if f.Versions[i].Version == version {
			f.Versions[i].Status = status
			return true
		}
	}
	return false
}

func (f *RuntimeFamily) Find(version string) *RuntimeVersion {
	for i := range f.Versions {
		if f.Versions[i].Version == version {
			return &f.Versions[i]
		}
	}
	return nil
}

func (f *RuntimeFamily) Remove(version string) {
	out := f.Versions[:0]
	for _, v := range f.Versions {
		if v.Version != version {
			out = append(out, v)
		}
	}
	f.Versions = out
	if f.Default == version {
		f.Default = ""
		for _, v := range f.Versions {
			if v.Status == constants.RuntimeStatusInstalled {
				f.Default = v.Version
			}
		}
	}
}

func (f *RuntimeFamily) HighestInstalledOrSelected(selected []string) string {
	best := ""
	for _, v := range f.Versions {
		if v.Status == constants.RuntimeStatusInstalled && v.Version > best {
			best = v.Version
		}
	}
	if best != "" {
		return best
	}
	for _, v := range selected {
		if v > best {
			best = v
		}
	}
	return best
}

func HighestVersion(versions []string) string {
	best := ""
	for _, v := range versions {
		if v > best {
			best = v
		}
	}
	return best
}

func SeedInstalling(phpVersions, nodeVersions []string, addons []string) ServerRuntimes {
	r := EmptyServerRuntimes()
	for _, v := range phpVersions {
		r.PHP.Upsert(v, constants.RuntimeStatusInstalling)
	}
	for _, v := range nodeVersions {
		r.Node.Upsert(v, constants.RuntimeStatusInstalling)
	}
	r.PHP.Default = HighestVersion(phpVersions)
	r.Node.Default = HighestVersion(nodeVersions)
	r.Addons = append([]string{}, addons...)
	sort.Strings(r.Addons)
	return r
}

func AddonList(flags map[string]bool) []string {
	order := []string{
		constants.AddonRedis,
		constants.AddonMySQL,
		constants.AddonMariaDB,
		constants.AddonPostgres,
	}
	var out []string
	for _, name := range order {
		if flags[name] {
			out = append(out, name)
		}
	}
	return out
}

func NormalizeVersionList(versions []string, allow func(string) bool) ([]string, string) {
	seen := map[string]struct{}{}
	var out []string
	for _, raw := range versions {
		v := strings.TrimSpace(raw)
		if v == "" || !allow(v) {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	sort.Strings(out)
	if len(out) == 0 {
		return nil, "at least one valid version is required"
	}
	return out, ""
}
