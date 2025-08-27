package targets_test

import (
	"testing"

	semver "github.com/Masterminds/semver/v3"
	"github.com/edgetx/cloudbuild/targets"
	"github.com/stretchr/testify/assert"
)

var targetsJSON = `{
	  "releases": {
	    "nightly": { "sha": "000" },
	  	"v1.3.0": { "sha": "123" },
	  	"v1.3.0-RC1": { "sha": "pre-123" },
	  	"v1.2.3": { "sha": "345" },
	  	"v1.1.2": { "sha": "456" }
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
	      "version_supported": "1.2 - 2.0"
	    }
	  }
	}`

func TestLoadTargetsJSON(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON))
	assert.Nil(t, err)
	assert.NotNil(t, defs)
}

func TestNightly(t *testing.T) {
	nightly := *targets.NightlyVersion
	assert.True(t, nightly.GreaterThan(semver.New(1, 0, 0, "", "")))
	assert.True(t, nightly.GreaterThan(semver.New(3, 0, 0, "foo", "bar")))
}

func TestVersions(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON))
	assert.Nil(t, err)

	//
	// verify map lookup
	//
	assert.True(t, defs.IsRefSupported("v1.1.2"))
	assert.True(t, defs.IsRefSupported("v1.2.3"))
	assert.True(t, defs.IsRefSupported("v1.3.0"))
	assert.True(t, defs.IsRefSupported("v1.3.0-RC1"))
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
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON))
	assert.Nil(t, err)

	assert.True(t, defs.IsTargetSupported("t1", "nightly"))
	assert.True(t, defs.IsTargetSupported("t1", "v1.1.2"))
	assert.True(t, defs.IsTargetSupported("t1", "v1.2.3"))
	assert.True(t, defs.IsTargetSupported("t1", "v1.3.0"))

	assert.False(t, defs.IsTargetSupported("t123", "v1.1.2"))
	assert.True(t, defs.IsTargetSupported("t123", "v1.2.3"))
	assert.True(t, defs.IsTargetSupported("t123", "v1.3.0"))

	assert.False(t, defs.IsTargetSupported("x123", "nightly"))
	assert.False(t, defs.IsTargetSupported("x123", "v1.1.2"))
	assert.True(t, defs.IsTargetSupported("x123", "v1.2.3"))
	assert.True(t, defs.IsTargetSupported("x123", "v1.3.0"))
}

func TestExcludeTargets(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(targetsJSON))
	assert.Nil(t, err)

	excl, err := defs.ExcludeTargetsFromRef("nightly")
	assert.Nil(t, err)
	assert.Equal(t, []string{"x123"}, excl)

	excl, err = defs.ExcludeTargetsFromRef("v1.1.2")
	assert.Nil(t, err)
	assert.Equal(t, []string{"t123", "x123"}, excl)
}
