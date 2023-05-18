package processor

import (
	"time"
)

type WorkerDto struct {
	ID        string    `json:"id"`
	Hostname  string    `json:"hostname"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func WorkerDtoFromModel(model *WorkerModel) WorkerDto {
	return WorkerDto{
		ID:        model.ID.String(),
		Hostname:  model.Hostname,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	}
}

func WorkersDtoFromModels(models *[]WorkerModel) *[]WorkerDto {
	dtos := make([]WorkerDto, len(*models))
	for i := range *models {
		dtos[i] = WorkerDtoFromModel(&(*models)[i])
	}
	return &dtos
}
