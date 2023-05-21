package database

import (
	"errors"

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

type Paginable interface {
	SetTotalRows(int64)
	GetOffset() int
	GetLimit() int
	GetSort() interface{}
}

type Pagination struct {
	Limit     int         `json:"limit,omitempty" form:"limit"`
	Offset    int         `json:"offset,omitempty" form:"offset"`
	Sort      string      `json:"sort,omitempty" form:"sort"`
	SortDesc  bool        `json:"sort_desc" form:"sort_desc"`
	TotalRows int64       `json:"total_rows"`
	Rows      interface{} `json:"rows"`
}

func (p *Pagination) SetTotalRows(rows int64) {
	p.TotalRows = rows
}

func (p *Pagination) GetOffset() int {
	return p.Offset
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

func Paginate(value interface{}, p Paginable, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var rows int64 = 0
	db.Model(value).Count(&rows)
	p.SetTotalRows(rows)
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(p.GetOffset()).Limit(p.GetLimit()).Order(p.GetSort())
	}
}
