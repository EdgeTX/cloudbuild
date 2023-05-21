package targets

import (
	"encoding/json"
	"os"

	"golang.org/x/exp/slices"
)

var (
	targetsDef = &TargetsDef{}
)

type Release struct {
	SHA            string   `json:"sha"`
	ExcludeTargets []string `json:"exclude_targets"`
}

type OptionFlag struct {
	BuildFlag string   `json:"build_flag"`
	Values    []string `json:"values"`
}

type BuildFlags map[string]string

type Target struct {
	Description string     `json:"desription"`
	Tags        []string   `json:"tags"`
	BuildFlags  BuildFlags `json:"build_flags"`
}

type OptionFlags map[string]OptionFlag

type TargetsDef struct {
	Releases    map[string]Release     `json:"releases"`
	OptionFlags OptionFlags            `json:"flags"`
	Tags        map[string]OptionFlags `json:"tags"`
	Targets     map[string]Target      `json:"targets"`
}

func ReadTargetsDefFromBytes(data []byte) error {
	if err := json.Unmarshal(data, targetsDef); err != nil {
		return err
	}
	return nil
}

func ReadTargetsDef(path string) error {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return ReadTargetsDefFromBytes(bytes)
}

func (opts OptionFlags) HasOptionValue(name, value string) bool {
	if opt, ok := opts[name]; ok {
		return slices.Contains(opt.Values, value)
	}
	return false
}

func (def *TargetsDef) IsRefSupported(ref string) bool {
	_, ok := def.Releases[ref]
	return ok
}

func (def *TargetsDef) IsTargetSupported(name, ref string) bool {
	release, ok := def.Releases[ref]
	if !ok {
		return false
	}

	if slices.Contains(release.ExcludeTargets, name) {
		return false
	}

	_, ok = def.Targets[name]
	return ok
}

func (def *TargetsDef) IsOptionFlagSupported(target, name, value string) bool {
	if def.OptionFlags.HasOptionValue(name, value) {
		return true
	}

	if t, ok := def.Targets[target]; ok {
		for _, tag := range t.Tags {
			flags, ok := def.Tags[tag]
			if ok {
				return flags.HasOptionValue(name, value)
			}
		}
	}

	return false
}

func (def *TargetsDef) GetCommitHashByRef(ref string) string {
	release, ok := def.Releases[ref]
	if !ok {
		return ""
	}
	return release.SHA
}

func (def *TargetsDef) GetTargetBuildFlags(target string) *BuildFlags {
	if t, ok := def.Targets[target]; ok {
		return &t.BuildFlags
	}
	return nil
}

func (def *TargetsDef) GetOptionBuildFlag(name string) string {
	if opt, ok := def.OptionFlags[name]; ok {
		return opt.BuildFlag
	}
	return ""
}

func IsRefSupported(ref string) bool {
	return targetsDef.IsRefSupported(ref)
}

func IsTargetSupported(name, ref string) bool {
	return targetsDef.IsTargetSupported(name, ref)
}

func IsOptionFlagSupported(target, name, value string) bool {
	return targetsDef.IsOptionFlagSupported(target, name, value)
}

func GetCommitHashByRef(ref string) string {
	return targetsDef.GetCommitHashByRef(ref)
}

func GetTargetBuildFlags(target string) *BuildFlags {
	return targetsDef.GetTargetBuildFlags(target)
}

func GetOptionBuildFlag(name string) string {
	return targetsDef.GetOptionBuildFlag(name)
}
