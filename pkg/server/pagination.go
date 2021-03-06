package server

import (
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	DefaultQueryPage = 1
	DefaultQuerySize = 10
	MaxQuerySize     = 1000000000
)

type Pagination struct {
	Page uint
	Size uint
}

func GetPagination(c echo.Context) Pagination {
	var p Pagination
	p.Page = uint(atoi(c.QueryParam("page"), DefaultQueryPage))
	p.Size = uint(atoi(c.QueryParam("size"), DefaultQuerySize))
	if p.Size > MaxQuerySize {
		p.Size = MaxQuerySize
	}
	return p
}

func (p *Pagination) Offset() uint {
	return (p.Page - 1) * p.Size
}

func atoi(s string, v int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return v
	}
	return i
}
