package server

import "github.com/edgetx/cloudbuild/firmware"

type BuildRequest struct {
	CommitHash string               `json:"commit_hash"`
	Flags      []firmware.BuildFlag `json:"flags"`
}

func (req *BuildRequest) Validate() []error {
	result := make([]error, 0)
	result = append(result, validateCommitHash(req.CommitHash)...)
	result = append(result, validateFlags(req.Flags)...)

	return result
}
