package firmware

import (
	"fmt"
	"regexp"
	"strings"
)

type BuildFlag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func (flag *BuildFlag) Format() string {
	if len(flag.Value) > 0 {
		return fmt.Sprintf("-D%s=%s", flag.Key, flag.Value)
	}
	return fmt.Sprintf("-D%s", flag.Key)
}

func EscapeFlagParam(input string) string {
	reg := regexp.MustCompile("[^A-z0-9_]+")
	return reg.ReplaceAllString(input, "")
}

func NewFlag(key string, value string) BuildFlag {
	return BuildFlag{
		Key:   EscapeFlagParam(key),
		Value: EscapeFlagParam(value),
	}
}

func CmakeFlags(flags []BuildFlag) string {
	data := make([]string, 0)
	for _, flag := range flags {
		data = append(data, flag.Format())
	}
	return strings.Join(data, " ")
}
