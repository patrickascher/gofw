package grid

import (
	"math"
	"strconv"

	"github.com/patrickascher/gofw/sqlquery"
)

const defaultLimit = 15

// pagination holds information about the source rows.
type pagination struct {
	Limit       int // -1 is infinity
	Prev        int
	Next        int
	CurrentPage int
	Total       int
	TotalPages  int
}

// newPagination creates a new pagination struct and requests the data of the given source.
func (g *Grid) newPagination(c *sqlquery.Condition) (*pagination, error) {

	p := &pagination{}

	if c == nil {
		c = &sqlquery.Condition{}
	}

	count, err := g.src.Count(c)
	if err != nil {
		return nil, err
	}
	p.Limit = p.paginationParam(g, "limit")
	p.Total = count
	p.TotalPages = p.totalPages()
	p.CurrentPage = p.paginationParam(g, "page")
	p.Next = p.next()
	p.Prev = p.prev()

	if p.Limit != -1 {
		c.Limit(p.Limit).Offset(p.offset())
	}

	return p, nil
}

// next checks if there is a next page.
// If its already the last page, 0 will return.
func (p *pagination) next() int {
	if p.CurrentPage < p.TotalPages {
		return p.CurrentPage + 1
	}
	return 0
}

// prev checks if there is a previous page.
// If its the first one, 0 will return.
func (p *pagination) prev() int {

	if p.CurrentPage > p.TotalPages {
		return p.TotalPages
	}

	if p.CurrentPage > 1 {
		return p.CurrentPage - 1
	}

	return 0
}

// offset returns the current offset.
func (p *pagination) offset() int {
	if p.CurrentPage <= 1 {
		return 0
	}

	return (p.CurrentPage - 1) * p.Limit
}

// totalPages returns the total number of pages.
// if there were no rows found or the limit is infinity, 1 will return.
func (p *pagination) totalPages() int {

	if p.Total == 0 || p.Limit == -1 {
		return 1
	}

	return int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

// paginationParam is checking the request param limit and page.
// if no limit per page was set, a default limit will be set.
func (p *pagination) paginationParam(g *Grid, q string) int {

	var param []string
	var err error
	var rv int

	switch q {
	case "limit":
		param, err = g.controller.Context().Request.Param("limit")
		rv = defaultLimit
	case "page":
		param, err = g.controller.Context().Request.Param("page")
		rv = 1
	}

	if err == nil && len(param) > 0 {
		s, err := strconv.Atoi(param[0])
		if err == nil {
			rv = s
		}
	}
	return rv
}
