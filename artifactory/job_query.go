package artifactory

import (
	"github.com/edgetx/cloudbuild/database"
)

type JobQuery struct {
	database.Pagination
	Status  string `form:"status"`
	Release string `form:"release"`
	Target  string `form:"target"`
	Sha     string `form:"sha"`
	NotSha  string `form:"not-sha"`
}

func (q *JobQuery) Validate() error {
	switch q.Sort {
	case "", "created_at", "updated_at":
		return nil
	case "build_started_at", "build_ended_at":
		return nil
	case "duration":
		return nil
	default:
		return database.ErrBadSortAttribute
	}
}

func (q *JobQuery) GetSort() interface{} {
	if q.Sort == "duration" {
		sort := "(build_ended_at - build_started_at)"
		if q.SortDesc {
			sort += " desc"
		}
		return sort
	}
	return q.Pagination.GetSort()
}
