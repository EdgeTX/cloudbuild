package artifactory

import (
	"time"

	"github.com/edgetx/cloudbuild/firmware"
)

type BuildJobDto struct {
	ID             string               `json:"id"`
	Status         BuildStatus          `json:"status"`
	BuildAttempts  int64                `json:"build_attempts"`
	CommitHash     string               `json:"commit_hash"`
	CommitRef      string               `json:"release"`
	Target         string               `json:"target"`
	Flags          []OptionFlag         `json:"flags"`
	BuildFlags     []firmware.BuildFlag `json:"build_flags"`
	Artifacts      []ArtifactDto        `json:"artifacts,omitempty"`
	AuditLogs      []AuditLogDto        `json:"build_logs,omitempty"`
	ContainerImage string               `json:"container_image"`
	BuildFlagsHash string               `json:"build_flags_hash"`
	BuildStartedAt time.Time            `json:"build_started_at"`
	BuildEndedAt   time.Time            `json:"build_ended_at"`
	CreatedAt      time.Time            `json:"created_at"`
	UpdatedAt      time.Time            `json:"updated_at"`
}

type ArtifactDto struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	DownloadURL string    `json:"download_url"`
	Size        int64     `json:"size"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AuditLogDto struct {
	ID        string      `json:"id"`
	From      BuildStatus `json:"from"`
	To        BuildStatus `json:"to"`
	StdOut    string      `json:"std_out"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
