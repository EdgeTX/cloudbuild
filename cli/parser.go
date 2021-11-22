package cli

import (
	"regexp"

	"github.com/edgetx/cloudbuild/firmware"
	"github.com/pkg/errors"
)

var CmakeFlagsRegexp = regexp.MustCompile("#*?-D([A-z0-9-_]*?)(=((\"(.*?)\")|([A-z0-9-_]+))|\\s|$)")

func ParseCmakeString(data string) ([]firmware.BuildFlag, error) {
	matches := CmakeFlagsRegexp.FindAllStringSubmatch(data, -1)
	if len(matches) == 0 {
		return nil, errors.New("failed to extract build flags")
	}

	results := make([]firmware.BuildFlag, 0)
	for _, match := range matches {
		var key string
		if len(match) >= 2 {
			key = match[1]
		}
		var value string
		if len(match) >= 4 {
			value = match[3]
		}
		results = append(results, firmware.NewFlag(key, value))
	}

	return results, nil
}
