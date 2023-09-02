package artifactory

import (
	"encoding/json"
	"net/url"

	"github.com/pkg/errors"

	"github.com/edgetx/cloudbuild/firmware"
)

var (
	ErrTypeError = errors.New("wrong type returned")
)

func BuildJobDtoFromModel(model *BuildJobModel, prefixURL *url.URL) (*BuildJobDto, error) {
	var optFlags []OptionFlag
	if model.Flags != nil {
		if err := json.Unmarshal([]byte(model.Flags.String()), &optFlags); err != nil {
			return nil, err
		}
	}
	var buildFlags []firmware.BuildFlag
	if model.BuildFlags != nil {
		strFlags := model.BuildFlags.String()
		if err := json.Unmarshal([]byte(strFlags), &buildFlags); err != nil {
			return nil, err
		}
	}
	artifacts := make([]ArtifactDto, 0)
	for i := range model.Artifacts {
		art := &model.Artifacts[i]
		downloadURL := prefixURL.JoinPath(art.Filename).String()
		artifacts = append(artifacts, ArtifactDtoFromModel(art, downloadURL))
	}
	auditLogs := make([]AuditLogDto, 0)
	for i := range model.AuditLogs {
		auditLogs = append(auditLogs, AuditLogDtoFromModel(&model.AuditLogs[i]))
	}
	return &BuildJobDto{
		ID:             model.ID.String(),
		Status:         model.Status,
		BuildAttempts:  model.BuildAttempts,
		CommitHash:     model.CommitHash,
		CommitRef:      model.CommitRef,
		Target:         model.Target,
		Flags:          optFlags,
		BuildFlags:     buildFlags,
		Artifacts:      artifacts,
		AuditLogs:      auditLogs,
		ContainerImage: model.ContainerImage,
		BuildFlagsHash: model.BuildFlagsHash,
		BuildStartedAt: model.BuildStartedAt,
		BuildEndedAt:   model.BuildEndedAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

func BuildJobsDtoFromInterface(input interface{}, prefixURL *url.URL) (*[]BuildJobDto, error) {
	jobs, ok := input.(*[]BuildJobModel)
	if !ok {
		return nil, ErrTypeError
	}

	resJobs := make([]BuildJobDto, len(*jobs))
	for i := range *jobs {
		j, err := BuildJobDtoFromModel(&(*jobs)[i], prefixURL)
		if err != nil {
			return nil, err
		}
		resJobs[i] = *j
	}
	return &resJobs, nil
}

func ArtifactDtoFromModel(model *ArtifactModel, downloadURL string) ArtifactDto {
	return ArtifactDto{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		DownloadURL: downloadURL,
		Size:        model.Size,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func AuditLogDtoFromModel(model *AuditLogModel) AuditLogDto {
	return AuditLogDto{
		ID:        model.ID.String(),
		From:      model.From,
		To:        model.To,
		StdOut:    model.StdOut,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}
