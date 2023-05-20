package artifactory

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"sort"

	"github.com/edgetx/cloudbuild/firmware"
	"github.com/edgetx/cloudbuild/targets"
)

var (
	ErrReleaseNotSupported    = errors.New("release not supported")
	ErrTargetNotSupported     = errors.New("target not supported")
	ErrOptionFlagNotSupported = errors.New("option flag not supported")
)

type OptionFlag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BuildRequest struct {
	Release string       `json:"release"`
	Target  string       `json:"target"`
	Flags   []OptionFlag `json:"flags"`
}

type BuildRequestError struct {
	What string
	Err  error
}

func (opt *OptionFlag) String() string {
	return fmt.Sprintf("%s=%s", opt.Name, opt.Value)
}

func (e *BuildRequestError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err.Error(), e.What)
}

func (req *BuildRequest) Validate() error {
	if !targets.IsRefSupported(req.Release) {
		return &BuildRequestError{
			Err:  ErrReleaseNotSupported,
			What: req.Release,
		}
	}
	if !targets.IsTargetSupported(req.Target, req.Release) {
		return &BuildRequestError{
			Err:  ErrTargetNotSupported,
			What: req.Target,
		}
	}
	for _, flag := range req.Flags {
		if !targets.IsOptionFlagSupported(req.Target, flag.Name, flag.Value) {
			return &BuildRequestError{
				Err:  ErrOptionFlagNotSupported,
				What: flag.String(),
			}
		}
	}
	return nil
}

func (req *BuildRequest) HashTargetAndFlags() string {
	// sort that array
	sort.Slice(req.Flags, func(i, j int) bool {
		lhs := &req.Flags[i]
		rhs := &req.Flags[j]
		if lhs.Name != rhs.Name {
			return lhs.Name < rhs.Name
		} else {
			return lhs.Value < rhs.Value
		}
	})

	// hash target + flags
	var hashData bytes.Buffer
	hashData.WriteString(req.Target)
	for i := range req.Flags {
		hashData.WriteString(req.Flags[i].String())
	}
	hash := sha256.New()
	hash.Write(hashData.Bytes())
	md := hash.Sum(nil)
	return hex.EncodeToString(md)
}

func (req *BuildRequest) GetBuildFlags() (*[]firmware.BuildFlag, error) {
	buildFlags := make([]firmware.BuildFlag, 0)
	// fetch target flags first
	targetFlags := targets.GetTargetBuildFlags(req.Target)
	if targetFlags == nil {
		return nil, ErrTargetNotSupported
	}
	for k, v := range *targetFlags {
		buildFlags = append(buildFlags, firmware.BuildFlag{
			Key:   k,
			Value: v,
		})
	}
	// then the option flags
	for _, optFlag := range req.Flags {
		buildFlag := targets.GetOptionBuildFlag(optFlag.Name)
		buildFlags = append(buildFlags, firmware.BuildFlag{
			Key:   buildFlag,
			Value: optFlag.Value,
		})
	}
	return &buildFlags, nil
}
