package targets

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
	"sync/atomic"

	semver "github.com/Masterminds/semver/v3"
	"github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/config"
	"github.com/go-git/go-git/v6/storage/memory"
	log "github.com/sirupsen/logrus"
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
	SHA            string   `json:"sha"`
	ExcludeTargets []string `json:"exclude_targets,omitempty"`
	BuildContainer string   `json:"build_container,omitempty"`
	SemVer         string   `json:"sem_ver,omitempty"`
	update         bool
	version        *semver.Version
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

func ReadTargetsDefFromBytes(data []byte, repoURL string) (*TargetsDef, error) {
	defs := TargetsDef{}
	if err := json.Unmarshal(data, &defs); err != nil {
		return nil, err
	}
	if err := defs.validateSHA(repoURL); err != nil {
		return nil, err
	}
	return &defs, nil
}

func ReadTargetsDef(path, repoURL string) (*TargetsDef, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ReadTargetsDefFromBytes(bytes, repoURL)
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
	v := r.v
	if v == *NightlyVersion {
		return "nightly"
	}
	return v.Original()
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
	return []byte(r.String()), nil
}

func (opts OptionFlags) HasOptionValue(name, value string) bool {
	if opt, ok := opts[name]; ok {
		return slices.Contains(opt.Values, value)
	}
	return false
}

func (target *Target) SupportsRelease(r *Release) bool {
	if len(target.VersionSupported.String()) > 0 {
		return target.VersionSupported.Check(r.version)
	}
	return true
}

func (def *TargetsDef) UnmarshalJSON(text []byte) error {
	type targetsDef TargetsDef
	var alias targetsDef

	if err := json.Unmarshal(text, &alias); err != nil {
		return err
	}

	tmp := TargetsDef(alias)
	if err := tmp.fillSemVer(); err != nil {
		return err
	}
	tmp.fillExcludeTargets()

	*def = tmp
	return nil
}

func (def *TargetsDef) validateSHA(repoURL string) error {
	var (
		tags map[string]string
		err  error
	)

	log.Debugf("Repository URL: %s", repoURL)
	if repoURL != "" {
		log.Debugf("Listing tags...")
		tags, err = ListTags(repoURL)
		if err != nil {
			return fmt.Errorf("Could not list tags from %s: %w", repoURL, err)
		}
	} else {
		tags = make(map[string]string)
	}

	for k := range def.Releases {
		v := def.Releases[k]
		if v.SHA == "" {
			tag := k.String()
			sha, ok := tags[tag]
			if !ok || (sha == "") {
				return fmt.Errorf("%s: %w", tag, ErrMissingSHA)
			}
			v.SHA = sha
			v.update = true
			log.Debugf("%s -> %s", tag, v.SHA)
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
				r.version = v
			}
		} else {
			r.version = &k.v
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

func ListTags(repoURL string) (tags map[string]string, err error) {
	remote := git.NewRemote(memory.NewStorage(), &config.RemoteConfig{
		Name: "origin",
		URLs: []string{repoURL},
	})

	refs, err := remote.List(&git.ListOptions{
		PeelingOption: git.AppendPeeled,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list remote references: %w", err)
	}

	tags = make(map[string]string)
	for _, ref := range refs {
		if ref.Name().IsTag() {
			shortRef, isPeeled := strings.CutSuffix(ref.Name().Short(), "^{}")
			_, ok := tags[shortRef]
			if !ok || isPeeled {
				tags[shortRef] = ref.Hash().String()
			}
		}
	}

	return tags, nil
}
