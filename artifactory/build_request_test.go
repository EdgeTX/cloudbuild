package artifactory_test

import (
	"testing"

	"github.com/edgetx/cloudbuild/artifactory"
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

	req1 := artifactory.BuildRequest{
		Target: "abcd",
		Flags: []artifactory.OptionFlag{
			flag1,
			flag2,
		},
	}
	req2 := artifactory.BuildRequest{
		Target: "abcd",
		Flags: []artifactory.OptionFlag{
			flag2,
			flag1,
		},
	}

	assert.Equal(t, req1.HashTargetAndFlags(), req2.HashTargetAndFlags())
}
