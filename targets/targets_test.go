package targets_test

import (
	"encoding/json"
	"os"
	"slices"
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/edgetx/cloudbuild/targets"
	"github.com/stretchr/testify/assert"
)

var targetsJSON = `{
  "releases": {
    "nightly": {
      "sha": "000",
      "semantic_version": "v2.0.0"
    },
    "v2.0.0-RC1": { "sha": "678" },
    "v1.3.0": { "sha": "567" },
    "v1.3.0-RC1": { "sha": "456" },
    "v1.2.3": { "sha": "345" },
    "v1.1.2": { "sha": "123" }
  },
  "targets": {
    "t1": {
      "description": "Always been there"
    },
    "t123": {
      "description": "Acme Dream Radio",
      "version_supported": ">= 1.2.3"
    },
    "x123": {
      "description": "XYZ Radio",
      "version_supported": ">= 1.2, < 2"
    }
  }
}`

func TestLoadTargetsJSON(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON), "")
	assert.Nil(t, err)
	assert.NotNil(t, defs)
}

func TestNightly(t *testing.T) {
	nightly := *targets.NightlyVersion
	assert.True(t, nightly.GreaterThan(semver.New(1, 0, 0, "", "")))
	assert.True(t, nightly.GreaterThan(semver.New(3, 0, 0, "foo", "bar")))
}

func TestVersions(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON), "")
	assert.Nil(t, err)

	//
	// verify map lookup
	//
	assert.True(t, defs.IsRefSupported("v1.1.2"))
	assert.True(t, defs.IsRefSupported("v1.2.3"))
	assert.True(t, defs.IsRefSupported("v1.3.0-RC1"))
	assert.True(t, defs.IsRefSupported("v1.3.0"))
	assert.True(t, defs.IsRefSupported("v2.0.0-RC1"))
	assert.True(t, defs.IsRefSupported("nightly"))

	//
	// Beware: doesn't work without 'v' (because of Version.original)
	//
	v1, err := semver.NewVersion("1.3.0")
	assert.Nil(t, err)
	v2, err := semver.NewVersion("v1.3.0")
	assert.Nil(t, err)

	// isn't equal according to Go's equality
	assert.NotEqual(t, v1, v2)
	assert.False(t, defs.IsRefSupported("1.3.0"))

	// ... but they compare the same with semver functions
	assert.True(t, v1.Equal(v2))

	//
	// release candidates
	//

	// v1.3.0 is greater than v1.3.0-RC1
	v3, err := semver.NewVersion("v1.3.0-RC1")
	assert.Nil(t, err)
	assert.True(t, v2.GreaterThan(v3))

	// v1.3.0-RC2 is greater than v1.3.0-RC1
	v4, err := semver.NewVersion("v1.3.0-RC2")
	assert.Nil(t, err)
	assert.True(t, v4.GreaterThan(v3))
}

func TestConstraints(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON), "")
	assert.Nil(t, err)

	assert.True(t, defs.IsTargetSupported("t1", "nightly"))
	assert.True(t, defs.IsTargetSupported("t1", "v1.1.2"))
	assert.True(t, defs.IsTargetSupported("t1", "v1.2.3"))
	assert.True(t, defs.IsTargetSupported("t1", "v1.3.0"))
	assert.True(t, defs.IsTargetSupported("t1", "v2.0.0-RC1"))

	assert.False(t, defs.IsTargetSupported("t123", "v1.1.2"))
	assert.True(t, defs.IsTargetSupported("t123", "v1.2.3"))
	assert.True(t, defs.IsTargetSupported("t123", "v1.3.0"))

	assert.False(t, defs.IsTargetSupported("x123", "nightly"))
	assert.False(t, defs.IsTargetSupported("x123", "v2.0.0-RC1"))
	assert.False(t, defs.IsTargetSupported("x123", "v1.1.2"))
	assert.True(t, defs.IsTargetSupported("x123", "v1.2.3"))
	assert.True(t, defs.IsTargetSupported("x123", "v1.3.0"))
}

func TestExcludeTargets(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON), "")
	assert.Nil(t, err)

	excl, err := defs.ExcludeTargetsFromRef("nightly")
	assert.Nil(t, err)
	assert.Equal(t, []string{"x123"}, excl)

	excl, err = defs.ExcludeTargetsFromRef("v1.1.2")
	assert.Nil(t, err)
	assert.Equal(t, []string{"t123", "x123"}, excl)
}

var flagConstraintsJSON = `{
  "releases": {
    "nightly": { "sha": "000", "sem_ver": "v3.0.0" },
    "v3.0.0": { "sha": "300" },
    "v3.0.0-rc1": { "sha": "301" },
    "v2.12.6-rc1": { "sha": "212" },
    "v2.12.5": { "sha": "211" }
  },
  "flags": {
    "language": {
      "build_flag": "TRANSLATIONS",
      "values": ["EN", "FR"],
      "values_version_supported": { "FR": "< 2.12.5" }
    }
  },
  "tags": {
    "colorlcd": {
      "flags": {
        "language": {
          "values": ["AR", "CN"],
          "values_version_supported": { "AR": ">= 3.0.0-0" }
        }
      }
    }
  },
  "targets": {
    "color": { "description": "Color radio", "tags": ["colorlcd"] },
    "bw": { "description": "B&W radio" }
  }
}`

func TestOptionFlagVersionSupported(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(flagConstraintsJSON), "")
	assert.Nil(t, err)

	tests := []struct {
		target, value, ref string
		supported          bool
		constraint         string
	}{
		// unconstrained values work everywhere
		{"color", "CN", "v2.12.5", true, ""},
		{"color", "EN", "v2.12.5", true, ""},
		// AR only from 3.0.0, including its pre-releases and nightly
		{"color", "AR", "v2.12.5", false, ">=3.0.0-0"},
		{"color", "AR", "v2.12.6-rc1", false, ">=3.0.0-0"},
		{"color", "AR", "v3.0.0-rc1", true, ""},
		{"color", "AR", "v3.0.0", true, ""},
		{"color", "AR", "nightly", true, ""},
		// global version_supported constraints apply to every target
		{"bw", "FR", "v2.12.5", false, "<2.12.5"},
		{"color", "FR", "nightly", false, "<2.12.5"},
		// unknown release
		{"color", "CN", "v9.9.9", false, ""},
	}
	for _, tt := range tests {
		ok, c := defs.IsOptionFlagSupportedByRelease(tt.target, "language", tt.value, tt.ref)
		assert.Equal(t, tt.supported, ok, "%s %s@%s", tt.target, tt.value, tt.ref)
		assert.Equal(t, tt.constraint, c, "%s %s@%s", tt.target, tt.value, tt.ref)
	}

	// release-independent check is unchanged
	assert.True(t, defs.IsOptionFlagSupported("color", "language", "AR"))
	assert.False(t, defs.IsOptionFlagSupported("bw", "language", "AR"))
}

func TestExcludeFlags(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(flagConstraintsJSON), "")
	assert.Nil(t, err)

	assert.Equal(t, map[string][]string{"language": {"AR", "FR"}}, defs.Releases[ref(t, "v2.12.5")].ExcludeFlags)
	assert.Equal(t, map[string][]string{"language": {"AR", "FR"}}, defs.Releases[ref(t, "v2.12.6-rc1")].ExcludeFlags)
	assert.Equal(t, map[string][]string{"language": {"FR"}}, defs.Releases[ref(t, "v3.0.0-rc1")].ExcludeFlags)
	assert.Equal(t, map[string][]string{"language": {"FR"}}, defs.Releases[ref(t, "nightly")].ExcludeFlags)

	// no constraints: no exclude_flags at all
	defs, err = targets.ReadTargetsDefFromBytes([]byte(targetsJSON), "")
	assert.Nil(t, err)
	for _, r := range defs.Releases {
		assert.Nil(t, r.ExcludeFlags)
	}
}

func TestFlagVersionSupportedJSON(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(flagConstraintsJSON), "")
	assert.Nil(t, err)

	data, err := json.Marshal(defs)
	assert.Nil(t, err)

	var out struct {
		Releases map[string]struct {
			ExcludeFlags map[string][]string `json:"exclude_flags"`
		} `json:"releases"`
		Tags map[string]struct {
			Flags map[string]map[string]any `json:"flags"`
		} `json:"tags"`
	}
	assert.Nil(t, json.Unmarshal(data, &out))
	assert.Equal(t, []string{"AR", "FR"}, out.Releases["v2.12.5"].ExcludeFlags["language"])
	assert.Equal(t,
		map[string]any{"AR": ">=3.0.0-0"},
		out.Tags["colorlcd"].Flags["language"]["values_version_supported"])

	// flags without constraints don't grow a new field, and releases don't
	// get exclude_flags (targets have their own, unrelated version_supported)
	defs, err = targets.ReadTargetsDefFromBytes([]byte(`{
	  "releases": { "v1.0.0": { "sha": "1" } },
	  "flags": { "language": { "build_flag": "TRANSLATIONS", "values": ["EN"] } },
	  "tags": { "colorlcd": { "flags": { "language": { "values": ["CN"] } } } },
	  "targets": { "t1": { "description": "T1", "tags": ["colorlcd"] } }
	}`), "")
	assert.Nil(t, err)
	data, err = json.Marshal(defs)
	assert.Nil(t, err)
	var plain struct {
		Flags map[string]map[string]any `json:"flags"`
		Tags  map[string]struct {
			Flags map[string]map[string]any `json:"flags"`
		} `json:"tags"`
	}
	assert.Nil(t, json.Unmarshal(data, &plain))
	for _, key := range []string{"version_supported", "values_version_supported"} {
		assert.NotContains(t, plain.Flags["language"], key)
		assert.NotContains(t, plain.Tags["colorlcd"].Flags["language"], key)
	}
	assert.NotContains(t, string(data), "exclude_flags")
}

var wholeFlagJSON = `{
  "releases": {
    "nightly": { "sha": "0", "sem_ver": "v3.0.0" },
    "v3.0.0-rc1": { "sha": "1" },
    "v2.12.6-rc1": { "sha": "2" },
    "v2.12.5": { "sha": "3" },
    "v2.10.6": { "sha": "4" }
  },
  "flags": {
    "power_led_blue": {
      "build_flag": "POWER_LED_BLUE",
      "values": ["YES", "NO"],
      "version_supported": "< 3.0.0-0"
    },
    "new_opt": {
      "build_flag": "NEW_OPT",
      "values": ["YES", "NO", "X"],
      "version_supported": ">= 2.12.0-0",
      "values_version_supported": { "X": ">= 3.0.0-0" }
    }
  },
  "targets": { "t": { "description": "T" } }
}`

func TestWholeFlagVersionSupported(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(wholeFlagJSON), "")
	assert.Nil(t, err)

	tests := []struct {
		flag, value, ref string
		supported        bool
		constraint       string
	}{
		// whole flag removed from 3.0, pre-releases included
		{"power_led_blue", "YES", "v2.12.5", true, ""},
		{"power_led_blue", "NO", "v2.12.6-rc1", true, ""},
		{"power_led_blue", "YES", "v3.0.0-rc1", false, "<3.0.0-0"},
		{"power_led_blue", "NO", "nightly", false, "<3.0.0-0"},
		// flag from 2.12 plus one value from 3.0: the flag limit is checked first
		{"new_opt", "YES", "v2.10.6", false, ">=2.12.0-0"},
		{"new_opt", "X", "v2.10.6", false, ">=2.12.0-0"},
		{"new_opt", "YES", "v2.12.5", true, ""},
		{"new_opt", "X", "v2.12.5", false, ">=3.0.0-0"},
		{"new_opt", "X", "nightly", true, ""},
	}
	for _, tt := range tests {
		ok, c := defs.IsOptionFlagSupportedByRelease("t", tt.flag, tt.value, tt.ref)
		assert.Equal(t, tt.supported, ok, "%s=%s @ %s", tt.flag, tt.value, tt.ref)
		assert.Equal(t, tt.constraint, c, "%s=%s @ %s", tt.flag, tt.value, tt.ref)
	}

	// clients drop a flag whose values are all excluded
	excl := func(release string) map[string][]string {
		return defs.Releases[ref(t, release)].ExcludeFlags
	}
	assert.Equal(t, map[string][]string{"new_opt": {"NO", "X", "YES"}}, excl("v2.10.6"))
	assert.Equal(t, map[string][]string{"new_opt": {"X"}}, excl("v2.12.5"))
	assert.Equal(t, map[string][]string{"power_led_blue": {"NO", "YES"}}, excl("nightly"))

	// served as a plain string, like a target's version_supported
	data, err := json.Marshal(defs)
	assert.Nil(t, err)
	var out struct {
		Flags map[string]map[string]any `json:"flags"`
	}
	assert.Nil(t, json.Unmarshal(data, &out))
	assert.Equal(t, "<3.0.0-0", out.Flags["power_led_blue"]["version_supported"])
	assert.NotContains(t, out.Flags["power_led_blue"], "values_version_supported")

	// invalid whole-flag constraints fail to load
	_, err = targets.ReadTargetsDefFromBytes([]byte(`{
	  "releases": { "v1.0.0": { "sha": "1" } },
	  "flags": { "f": { "values": ["YES"], "version_supported": "not a constraint" } }
	}`), "")
	assert.NotNil(t, err)
}

func TestInvalidFlagVersionSupported(t *testing.T) {
	undefinedValue := `{
	  "releases": { "v1.0.0": { "sha": "1" } },
	  "tags": { "colorlcd": { "flags": { "language": {
	    "values": ["CN"], "values_version_supported": { "AR": ">= 3.0.0-0" }
	  }}}}
	}`
	_, err := targets.ReadTargetsDefFromBytes([]byte(undefinedValue), "")
	assert.ErrorIs(t, err, targets.ErrInvalidValuesVersionSupported)

	badConstraint := `{
	  "releases": { "v1.0.0": { "sha": "1" } },
	  "flags": { "language": {
	    "values": ["AR"], "values_version_supported": { "AR": "not a constraint" }
	  }}
	}`
	_, err = targets.ReadTargetsDefFromBytes([]byte(badConstraint), "")
	assert.NotNil(t, err)
}

// loadShippedTargets loads the shipped targets.json. json.Unmarshal rather
// than ReadTargetsDefFromBytes: releases have no SHA in the file, they are
// resolved from the git repo at runtime.
func loadShippedTargets(t *testing.T) *targets.TargetsDef {
	t.Helper()
	data, err := os.ReadFile("../targets.json")
	assert.Nil(t, err)
	var defs targets.TargetsDef
	assert.Nil(t, json.Unmarshal(data, &defs))

	// older releases are still offered, so the limits tested are exercised
	oldest, err := targets.NewVersionRef("v2.8.5")
	assert.Nil(t, err)
	assert.Contains(t, defs.Releases, *oldest)
	return &defs
}

func colorTarget(t *testing.T, defs *targets.TargetsDef) string {
	t.Helper()
	for name, target := range defs.Targets {
		if slices.Contains(target.Tags, "colorlcd") {
			return name
		}
	}
	t.Fatal("no colorlcd target")
	return ""
}

// isAtLeast reports whether release ref is >= minVersion (nightly always is).
func isAtLeast(ref string, minVersion *semver.Version) bool {
	v, err := semver.NewVersion(ref)
	return ref == "nightly" || (err == nil && !v.LessThan(minVersion))
}

// Languages must only be offered from the release where they appear in
// EdgeTX's RADIO_LANGUAGES (radio/src/CMakeLists.txt).
func TestShippedTargetsLanguageReleases(t *testing.T) {
	defs := loadShippedTargets(t)
	color := colorTarget(t, defs)

	// language -> first EdgeTX release that has it
	firstRelease := map[string]string{
		"HE": "2.9.0",
		"RU": "2.10.0",
		"UA": "2.10.0",
		"KO": "2.11.0",
		"AR": "3.0.0",
	}
	for lang, first := range firstRelease {
		minVersion := semver.MustParse(first)
		for k, r := range defs.Releases {
			want := isAtLeast(k.String(), minVersion)
			ok, _ := defs.IsOptionFlagSupportedByRelease(color, "language", lang, k.String())
			assert.Equal(t, want, ok, "%s @ %s", lang, k.String())
			assert.Equal(t, want, !slices.Contains(r.ExcludeFlags["language"], lang),
				"exclude_flags %s @ %s", lang, k.String())
		}
	}
}

// Languages without B&W (stdlcd) fonts must only be offered to colour
// targets, e.g. KO only has LVGL fonts (EdgeTX/edgetx#6118).
func TestShippedTargetsColorOnlyLanguages(t *testing.T) {
	defs := loadShippedTargets(t)
	colorOnly := []string{"AR", "CN", "HE", "JP", "KO", "SK", "TW"}
	for name, target := range defs.Targets {
		isColor := slices.Contains(target.Tags, "colorlcd")
		for _, lang := range colorOnly {
			assert.Equal(t, isColor, defs.IsOptionFlagSupported(name, "language", lang),
				"%s on %s", lang, name)
		}
	}
}

// POWER_LED_BLUE was removed in EdgeTX 3.0.
func TestShippedTargetsPowerLedBlue(t *testing.T) {
	defs := loadShippedTargets(t)
	v3 := semver.MustParse("3.0.0-0")
	for k, r := range defs.Releases {
		want := !isAtLeast(k.String(), v3)
		for _, value := range []string{"YES", "NO"} {
			ok, _ := defs.IsOptionFlagSupportedByRelease("x9d", "power_led_blue", value, k.String())
			assert.Equal(t, want, ok, "power_led_blue=%s @ %s", value, k.String())
		}
		assert.Equal(t, want, len(r.ExcludeFlags["power_led_blue"]) == 0,
			"exclude_flags power_led_blue @ %s", k.String())
	}
}
func ref(t *testing.T, s string) targets.VersionRef {
	t.Helper()
	r, err := targets.NewVersionRef(s)
	assert.Nil(t, err)
	return *r
}
