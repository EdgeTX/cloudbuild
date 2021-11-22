package cli_test

import (
	"testing"

	"github.com/edgetx/cloudbuild/cli"
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/stretchr/testify/assert"
)

func TestParseCmake1(t *testing.T) {
	data, err := cli.ParseCmakeString("-DARG1 -DARG2=VALUE -DARG3=\"VALUE\"")
	assert.Nil(t, err)
	expected := []firmware.BuildFlag{
		{Key: "ARG1", Value: ""},
		{Key: "ARG2", Value: "VALUE"},
		{Key: "ARG3", Value: "VALUE"},
	}
	assert.Equal(t, data, expected)
}

func TestParseCmake2(t *testing.T) {
	data, err := cli.ParseCmakeString("-DARG1 -DARG2=VALUE -DARG3=\"VALUE\" -DARG4")
	assert.Nil(t, err)
	expected := []firmware.BuildFlag{
		{Key: "ARG1", Value: ""},
		{Key: "ARG2", Value: "VALUE"},
		{Key: "ARG3", Value: "VALUE"},
		{Key: "ARG4", Value: ""},
	}
	assert.Equal(t, data, expected)
}
