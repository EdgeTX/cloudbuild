package targets

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"sync/atomic"

	semver "github.com/Masterminds/semver/v3"
	"golang.org/x/exp/slices"
)

var (
	// global targets pointer.
	targetsDef = atomic.Pointer[TargetsDef]{}

	// absurdly high version.
	NightlyVersion = semver.New(math.MaxUint64, 0, 0, "", "")

	ErrMissingSHA = errors.New("missing SHA")
	ErrMissingRef = errors.New("missing ref")
)

type RemoteAPI struct {
	URL  string `json:"url"`
	Path string `json:"json_path"`
}

type Release struct {
	SHA            string          `json:"sha"`
	Remote         *RemoteSHA      `json:"remote,omitempty"`
	ExcludeTargets []string        `json:"exclude_targets,omitempty"`
	BuildContainer string          `json:"build_container,omitempty"`
	SemVer         string          `json:"sem_ver,omitempty"`
	Version        *semver.Version `json:"-"`
}

type OptionFlag struct {
	BuildFlag string   `json:"build_flag,omitempty"`
	Values    []string `json:"values"`
}

type BuildFlags map[string]string

type Target struct {
	Description      string             `json:"description"`
	Tags             []string           `json:"tags,omitempty"`
	BuildFlags       BuildFlags         `json:"build_flags,omitempty"`
	VersionSupported semver.Constraints `json:"version_supported,omitempty"`
}

type OptionFlags map[string]OptionFlag

type TagDef struct {
	Flags OptionFlags `json:"flags"`
}

type VersionRef struct {
	v semver.Version
}

type TargetsDef struct {
	Releases    map[VersionRef]*Release `json:"releases"`
	OptionFlags OptionFlags             `json:"flags"`
	Tags        map[string]TagDef       `json:"tags"`
	Targets     map[string]*Target      `json:"targets"`
}

func ReadTargetsDefFromBytes(data []byte) (*TargetsDef, error) {
	defs := TargetsDef{}
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, err
	}
	if err := defs.validateSHA(); err != nil {
		return nil, err
	}
	if err := defs.fillSemVer(); err != nil {
		return nil, err
	}
	defs.fillExcludeTargets()
	return &defs, nil
}

func ReadTargetsDef(path string) (*TargetsDef, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadTargetsDefFromBytes(bytes)
}

func NewVersionRef(v string) (*VersionRef, error) {
	if v == "nightly" {
		return &VersionRef{*NightlyVersion}, nil
	}
	if version, err := semver.NewVersion(v); err != nil {
		return nil, err
	} else {
		return &VersionRef{*version}, nil
	}
}

func (r *VersionRef) String() string {
	return r.v.String()
}

func (r *VersionRef) UnmarshalText(text []byte) error {
	v, err := NewVersionRef(string(text))
	if err != nil {
		return err
	}
	*r = *v
	return nil
}

func (r VersionRef) MarshalText() ([]byte, error) {
	v := r.v
	if v == *NightlyVersion {
		return []byte("nightly"), nil
	}
	return []byte(v.Original()), nil
}

func (opts OptionFlags) HasOptionValue(name, value string) bool {
	if opt, ok := opts[name]; ok {
		return slices.Contains(opt.Values, value)
	}
	return false
}

func (target *Target) SupportsRelease(r *Release) bool {
	if len(target.VersionSupported.String()) > 0 {
		return target.VersionSupported.Check(r.Version)
	}
	return true
}

func (def *TargetsDef) validateSHA() error {
	for k := range def.Releases {
		v := def.Releases[k]
		if v.SHA == "" {
			if v.Remote == nil || v.Remote.URL == "" {
				return fmt.Errorf("%v: %w", k, ErrMissingSHA)
			}
			sha, err := v.Remote.Fetch()
			if err != nil {
				return fmt.Errorf("%v: %w", k, err)
			}
			v.SHA = sha
		}
	}
	return nil
}

func (def *TargetsDef) fillExcludeTargets() {
	for k := range def.Releases {
		r := def.Releases[k]
		r.ExcludeTargets = make([]string, 0)
		for t := range def.Targets {
			if !def.Targets[t].SupportsRelease(r) {
				r.ExcludeTargets = append(r.ExcludeTargets, t)
			}
		}
		slices.Sort(r.ExcludeTargets)
	}
}

func (def *TargetsDef) fillSemVer() error {
	for k := range def.Releases {
		r := def.Releases[k]
		if r.SemVer != "" {
			if v, err := semver.NewVersion(r.SemVer); err != nil {
				return err
			} else {
				r.Version = v
			}
		} else {
			r.Version = &k.v
		}
	}
	return nil
}

func (def *TargetsDef) IsRefSupported(ref string) bool {
	v, err := NewVersionRef(ref)
	if err != nil {
		return false
	}
	_, ok := def.Releases[*v]
	return ok
}

func (def *TargetsDef) IsTargetSupported(name, ref string) bool {
	v, err := NewVersionRef(ref)
	if err != nil {
		return false
	}
	r, ok := def.Releases[*v]
	if !ok {
		return false
	}
	target, ok := def.Targets[name]
	if !ok {
		return false
	}
	return target.SupportsRelease(r)
}

func (def *TargetsDef) IsOptionFlagSupported(target, name, value string) bool {
	if def.OptionFlags.HasOptionValue(name, value) {
		return true
	}

	if t, ok := def.Targets[target]; ok {
		for _, tag := range t.Tags {
			tagDef, ok := def.Tags[tag]
			if ok && tagDef.Flags.HasOptionValue(name, value) {
				return true
			}
		}
	}

	return false
}

func (def *TargetsDef) GetCommitHashByRef(ref string) string {
	v, err := NewVersionRef(ref)
	if err != nil {
		return ""
	}
	release, ok := def.Releases[*v]
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

func (def *TargetsDef) GetOptionBuildFlag(target, name string) string {
	if opt, ok := def.OptionFlags[name]; ok {
		return opt.BuildFlag
	}

	if t, ok := def.Targets[target]; ok {
		for _, tag := range t.Tags {
			tagDef, ok := def.Tags[tag]
			if ok {
				if opt, ok := tagDef.Flags[name]; ok {
					return opt.BuildFlag
				}
			}
		}
	}

	return ""
}

func (def *TargetsDef) GetBuildContainer(ref string) string {
	v, err := NewVersionRef(ref)
	if err != nil {
		return ""
	}
	release, ok := def.Releases[*v]
	if !ok {
		return ""
	}
	return release.BuildContainer
}

func (def *TargetsDef) ExcludeTargetsFromRef(ref string) ([]string, error) {
	v, err := NewVersionRef(ref)
	if err != nil {
		return nil, err
	}
	r, ok := def.Releases[*v]
	if !ok {
		return nil, ErrMissingRef
	}
	return r.ExcludeTargets, nil
}

func SetTargets(defs *TargetsDef) {
	targetsDef.Store(defs)
}

func GetTargets() *TargetsDef {
	return targetsDef.Load()
}
