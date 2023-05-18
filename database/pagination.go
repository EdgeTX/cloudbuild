package database

import (
	"errors"

	log "github.com/sirupsen/logrus"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	MinPageSize = 10
	MaxPageSize = 50
)

var (
	ErrBadSortAttribute = errors.New("bad sorting attribute")
)

type Pagination struct {
	Limit     int         `json:"limit,omitempty" form:"limit"`
	Offset    int         `json:"offset,omitempty" form:"offset"`
	Sort      string      `json:"sort,omitempty" form:"sort"`
	SortDesc  bool        `json:"sort_desc" form:"sort_desc"`
	TotalRows int64       `json:"total_rows"`
	Rows      interface{} `json:"rows"`
}

func (p *Pagination) GetLimit() int {
	if p.Limit <= 0 {
		p.Limit = MinPageSize
	} else if p.Limit > MaxPageSize {
		p.Limit = MaxPageSize
	}
	return p.Limit
}

func (p *Pagination) GetSort() interface{} {
	if p.Sort == "" {
		return nil
	}
	return clause.OrderByColumn{
		Column: clause.Column{Name: p.Sort},
		Desc:   p.SortDesc,
	}
}

func Paginate(value interface{}, p *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	db.Model(value).Count(&p.TotalRows)
	log.Debugln("Sort:", p.Sort, "SoftDesc:", p.SortDesc)
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.Offset).Limit(p.GetLimit()).Order(p.GetSort())
	}
}
