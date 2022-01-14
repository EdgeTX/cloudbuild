package server

import (
	"github.com/edgetx/cloudbuild/firmware"
	"github.com/pkg/errors"
)

func validateCommitHash(commitHash string) []error {
	result := make([]error, 0)

	if len(commitHash) == 0 {
		result = append(result, errors.New("commit hash is empty"))
	}

	if !(len(commitHash) == 40) {
		result = append(result, errors.New("commit hash length is invalid"))
	}

	return result
}

func isFlagKeyValid(key string) bool {
	validKeys := []string{
		"PCB",
		"PCBREV",
		"AFHDS2",
		"AFHDS3",
		"AUTOUPDATE",
		"BLUETOOTH",
		"CROSSFIRE",
		"DEBUG",
		"DEFAULT_MODE",
		"DSM2",
		"FAI",
		"FLYSKY_HALL_STICKS",
		"FLYSKY_HALL_STICKS_REVERSE",
		"FRSKY_STICKS",
		"GHOST",
		"GVARS",
		"HARDWARE_EXTERNAL_ACCESS_MOD",
		"HARDWARE_TRAINER_MULTI",
		"HELI",
		"IMU_LSM6DS33",
		"INTERNAL_GPS",
		"INTERNAL_MODULE_MULTI",
		"INTERNAL_MODULE_CRSF",
		"LOG_BLUETOOTH",
		"LOG_TELEMETRY",
		"LUA",
		"MODULE_PROTOCOL_D8",
		"MODULE_PROTOCOL_FCC",
		"MODULE_PROTOCOL_FLEX",
		"MODULE_PROTOCOL_LBT",
		"MODULE_SIZE_STD",
		"MULTIMODULE",
		"OVERRIDE_CHANNEL_FUNCTION",
		"PPM",
		"PPM_UNIT",
		"PXX1",
		"RADIOMASTER_RTF_RELEASE",
		"SBUS",
		"TRANSLATIONS",
		"UNEXPECTED_SHUTDOWN",
		"USB_SERIAL",
		"XJT",
		"YAML_STORAGE",
		// cmake related
		"VERBOSE_CMAKELISTS",
		"DISABLE_COMPANION",
		"CMAKE_BUILD_TYPE",
		"TRACE_SIMPGMSPACE",
		"CMAKE_RULE_MESSAGES",
	}
	for i := range validKeys {
		if key == validKeys[i] {
			return true
		}
	}
	return false
}

func validateFlags(firmwareFlags []firmware.BuildFlag) []error {
	result := make([]error, 0)

	if len(firmwareFlags) == 0 {
		result = append(result, errors.New("flags are null"))
	}

	for _, f := range firmwareFlags {
		if len(firmware.EscapeFlagParam(f.Key)) != len(f.Key) {
			result = append(result, errors.Errorf("flag key: %s contains not allowed characters", f.Key))
		}
		if !isFlagKeyValid(f.Key) {
			result = append(result, errors.Errorf("flag key: %s is not among allowed build flags list", f.Key))
		}
		if len(firmware.EscapeFlagParam(f.Value)) != len(f.Value) {
			result = append(result, errors.Errorf("flag value: %s contains not allowed characters", f.Key))
		}
	}

	return result
}
