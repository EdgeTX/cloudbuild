package targets

import (
	"encoding/json"
	"os"
	"sync/atomic"

	"golang.org/x/exp/slices"
)

var (
	targetsDef = atomic.Pointer[TargetsDef]{}
)

type Release struct {
	SHA            string   `json:"sha"`
	ExcludeTargets []string `json:"exclude_targets,omitempty"`
}

type OptionFlag struct {
	BuildFlag string   `json:"build_flag,omitempty"`
	Values    []string `json:"values"`
}

type BuildFlags map[string]string

type Target struct {
	Description string     `json:"description"`
	Tags        []string   `json:"tags,omitempty"`
	BuildFlags  BuildFlags `json:"build_flags"`
}

type OptionFlags map[string]OptionFlag

type TagDef struct {
	Flags OptionFlags `json:"flags"`
}

type TargetsDef struct {
	Releases    map[string]Release `json:"releases"`
	OptionFlags OptionFlags        `json:"flags"`
	Tags        map[string]TagDef  `json:"tags"`
	Targets     map[string]Target  `json:"targets"`
}

func ReadTargetsDefFromBytes(data []byte) error {
	defs := TargetsDef{}
	if err := json.Unmarshal(data, &defs); err != nil {
		return err
	}
	targetsDef.Store(&defs)
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
			tagDef, ok := def.Tags[tag]
			if ok {
				return tagDef.Flags.HasOptionValue(name, value)
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
	return targetsDef.Load().IsRefSupported(ref)
}

func IsTargetSupported(name, ref string) bool {
	return targetsDef.Load().IsTargetSupported(name, ref)
}

func IsOptionFlagSupported(target, name, value string) bool {
	return targetsDef.Load().IsOptionFlagSupported(target, name, value)
}

func GetCommitHashByRef(ref string) string {
	return targetsDef.Load().GetCommitHashByRef(ref)
}

func GetTargetBuildFlags(target string) *BuildFlags {
	return targetsDef.Load().GetTargetBuildFlags(target)
}

func GetOptionBuildFlag(name string) string {
	return targetsDef.Load().GetOptionBuildFlag(name)
}

func SetTargets(defs *TargetsDef) {
	targetsDef.Store(defs)
}

func GetTargets() *TargetsDef {
	return targetsDef.Load()
}
