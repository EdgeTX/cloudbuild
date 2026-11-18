package artifactory_test

import (
	"testing"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/targets"
	"github.com/stretchr/testify/assert"
)

func TestBuildHashIsTheSameForDiffOrderFlagArrays(t *testing.T) {
	flag1 := artifactory.OptionFlag{
		Name:  "A",
		Value: "first",
	}
	flag2 := artifactory.OptionFlag{
		Name:  "B",
		Value: "second",
	}

	req1 := artifactory.NewBuildRequestWithParams(
		"v0.0.0",
		"abcd",
		[]artifactory.OptionFlag{
			flag1,
			flag2,
		},
	)
	req2 := artifactory.NewBuildRequestWithParams(
		"v0.0.0",
		"abcd",
		[]artifactory.OptionFlag{
			flag2,
			flag1,
		},
	)

	assert.Equal(t, req1.HashTargetAndFlags(), req2.HashTargetAndFlags())
}

func TestValidateOptionFlagRelease(t *testing.T) {
	defs, err := targets.ReadTargetsDefFromBytes([]byte(`{
	  "releases": {
	    "nightly": { "sha": "000", "sem_ver": "v3.0.0" },
	    "v2.12.5": { "sha": "211" }
	  },
	  "flags": {
	    "language": { "build_flag": "TRANSLATIONS", "values": ["EN"] },
	    "power_led_blue": {
	      "build_flag": "POWER_LED_BLUE", "values": ["YES", "NO"],
	      "version_supported": "< 3.0.0-0"
	    }
	  },
	  "tags": {
	    "colorlcd": { "flags": { "language": {
	      "values": ["AR"], "values_version_supported": { "AR": ">= 3.0.0-0" }
	    }}}
	  },
	  "targets": {
	    "tx16s": { "description": "TX16S", "tags": ["colorlcd"] },
	    "x7": { "description": "X7" }
	  }
	}`), "")
	assert.Nil(t, err)

	prev := targets.GetTargets()
	targets.SetTargets(defs)
	defer targets.SetTargets(prev)

	ar := []artifactory.OptionFlag{{Name: "language", Value: "AR"}}

	assert.Nil(t, artifactory.NewBuildRequestWithParams("nightly", "tx16s", ar).Validate())

	err = artifactory.NewBuildRequestWithParams("v2.12.5", "tx16s", ar).Validate()
	assert.ErrorIs(t, err, artifactory.ErrOptionFlagNotSupportedByRelease)
	assert.EqualError(t, err,
		"option flag not supported by release: language=AR requires release >=3.0.0-0, got v2.12.5")

	// value unknown for the target: unchanged error
	err = artifactory.NewBuildRequestWithParams("nightly", "x7", ar).Validate()
	assert.ErrorIs(t, err, artifactory.ErrOptionFlagNotSupported)

	// whole flag limited to releases before 3.0
	blue := []artifactory.OptionFlag{{Name: "power_led_blue", Value: "NO"}}
	assert.Nil(t, artifactory.NewBuildRequestWithParams("v2.12.5", "x7", blue).Validate())
	err = artifactory.NewBuildRequestWithParams("nightly", "x7", blue).Validate()
	assert.EqualError(t, err,
		"option flag not supported by release: power_led_blue=NO requires release <3.0.0-0, got nightly")
}
