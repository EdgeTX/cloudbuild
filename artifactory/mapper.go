package artifactory

import (
	"encoding/json"

	"github.com/pkg/errors"

	"github.com/edgetx/cloudbuild/firmware"
)

func BuildJobDtoFromModel(model *BuildJobModel) (*BuildJobDto, error) {
	if model.BuildFlags == nil {
		return nil, errors.New("build flags are empty")
	}
	var buildFlags []firmware.BuildFlag
	err := json.Unmarshal([]byte(model.BuildFlags.String()), &buildFlags)
	if err != nil {
		return nil, err
	}
	artifacts := make([]ArtifactDto, 0)
	for i := range model.Artifacts {
		artifacts = append(artifacts, ArtifactDtoFromModel(&model.Artifacts[i]))
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
		ContainerImage: model.ContainerImage,
		BuildFlags:     buildFlags,
		BuildFlagsHash: model.BuildFlagsHash,
		Artifacts:      artifacts,
		AuditLogs:      auditLogs,
		BuildStartedAt: model.BuildStartedAt,
		BuildEndedAt:   model.BuildEndedAt,
		CreatedAt:      model.CreatedAt,
		UpdatedAt:      model.UpdatedAt,
	}, nil
}

func ArtifactDtoFromModel(model *ArtifactModel) ArtifactDto {
	return ArtifactDto{
		ID:          model.ID.String(),
		Slug:        model.Slug,
		DownloadURL: model.DownloadURL,
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
