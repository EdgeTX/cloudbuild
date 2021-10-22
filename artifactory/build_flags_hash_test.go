package artifactory_test

import (
	"testing"

	"github.com/edgetx/cloudbuild/artifactory"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/stretchr/testify/assert"
)

func TestBuildHashIsTheSameForDiffOrderFlagArrays(t *testing.T) {
	flags1 := []firmware.BuildFlag{
		firmware.NewFlag("A", "first"),
		firmware.NewFlag("B", "second"),
	}
	flags2 := []firmware.BuildFlag{
		firmware.NewFlag("B", "second"),
		firmware.NewFlag("A", "first"),
	}
	assert.Equal(t, artifactory.HashBuildFlags(flags1), artifactory.HashBuildFlags(flags2))
}
