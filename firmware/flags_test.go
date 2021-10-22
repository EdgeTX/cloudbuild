package firmware_test

import (
	"testing"

	"github.com/edgetx/cloudbuild/firmware"
	"github.com/stretchr/testify/assert"
)

func TestNewFlagSanitize(t *testing.T) {
	f := firmware.NewFlag("TE\n\t&&ST-1234A___.%\n$#&_", "12345")
	assert.Equal(t, "TEST1234A____", f.Key)
}

func TestCmakeFlags(t *testing.T) {
	flags := []firmware.BuildFlag{
		firmware.NewFlag("PCB", "X10"),
		firmware.NewFlag("PCBREV", "TX16S"),
		firmware.NewFlag("DEFAULT_MODE", "2"),
		firmware.NewFlag("LUA", "YES"),
		firmware.NewFlag("INTERNAL_GPS", "YES"),
		firmware.NewFlag("CMAKE_BUILD_TYPE", "Release"),
	}
	expected := "-DPCB=X10 -DPCBREV=TX16S -DDEFAULT_MODE=2 -DLUA=YES -DINTERNAL_GPS=YES -DCMAKE_BUILD_TYPE=Release"
	assert.Equal(t, expected, firmware.CmakeFlags(flags))
}
