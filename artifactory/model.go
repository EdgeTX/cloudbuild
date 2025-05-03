package artifactory

import (
	"time"

	"github.com/edgetx/cloudbuild/database"
	uuid "github.com/satori/go.uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

type BuildStatus string

const (
	VoidStatus      BuildStatus = "VOID"
	WaitingForBuild BuildStatus = "WAITING_FOR_BUILD"
	BuildInProgress BuildStatus = "BUILD_IN_PROGRESS"
	BuildSuccess    BuildStatus = "BUILD_SUCCESS"
	BuildError      BuildStatus = "BUILD_ERROR"
)

type BuildErrorType string

const (
	MaxBuildAttempts = 3
	MaxBuildDuration = time.Minute * 15
)

type BuildJobModel struct {
	ID             uuid.UUID   `gorm:"type:uuid;primary_key;"`
	Status         BuildStatus `gorm:"index:build_job_status_idx"`
	BuildAttempts  int64
	CommitHash     string `gorm:"index:commit_hash_idx"`
	CommitRef      string `gorm:"index:commit_ref_idx"`
	Target         string `gorm:"index:target_idx"`
	Flags          datatypes.JSON
	BuildFlags     datatypes.JSON
	ContainerImage string
	BuildFlagsHash string          `gorm:"index:build_flags_hash_idx"`
	Artifacts      []ArtifactModel `gorm:"foreignKey:BuildJobID;constraint:OnDelete:CASCADE"`
	AuditLogs      []AuditLogModel `gorm:"foreignKey:BuildJobID;constraint:OnDelete:CASCADE"`
	BuildStartedAt time.Time
	BuildEndedAt   time.Time
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

func (BuildJobModel) TableName() string {
	return "build_jobs"
}

func (base *BuildJobModel) BeforeCreate(db *gorm.DB) error {
	base.ID = uuid.NewV4()
	return nil
}

type ArtifactModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;"`
	Slug       string
	BuildJobID string
	BuildJob   BuildJobModel `gorm:"foreignKey:BuildJobID"`
	Filename   string
	Size       int64
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (ArtifactModel) TableName() string {
	return "artifacts"
}

func (base *ArtifactModel) BeforeCreate(db *gorm.DB) error {
	base.ID = uuid.NewV4()
	return nil
}

type AuditLogModel struct {
	ID         uuid.UUID `gorm:"type:uuid;primary_key;"`
	BuildJobID string
	BuildJob   BuildJobModel `gorm:"foreignKey:BuildJobID"`
	RequestIP  string
	From       BuildStatus
	To         BuildStatus
	StdOut     string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (AuditLogModel) TableName() string {
	return "audit_logs"
}

func (base *AuditLogModel) BeforeCreate(db *gorm.DB) error {
	base.ID = uuid.NewV4()
	return nil
}

func init() {
	database.RegisterModels(
		&BuildJobModel{},
		&ArtifactModel{},
		&AuditLogModel{},
	)
}
