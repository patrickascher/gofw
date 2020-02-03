package grid

import (
	"github.com/patrickascher/gofw/sqlquery"
	"math"
	"strconv"
)

// paginationOffset has all pagination info.
type paginationOffset struct {
	Limit int

	Prev        int
	Next        int
	CurrentPage int
	Total       int
	TotalPages  int
}

// next checks if there is a next page.
// If its already the last page, 0 will return.
func (p *paginationOffset) next() int {
	if p.CurrentPage < p.TotalPages {
		return p.CurrentPage + 1
	}
	return 0
}

// offset returns the current offset.
func (p *paginationOffset) offset() int {
	if p.CurrentPage == 1 {
		return 0
	}

	return (p.CurrentPage - 1) * p.Limit
}

// prev checks if there is a previous page.
// If its the first one, 0 will return.
func (p *paginationOffset) prev() int {

	if p.CurrentPage > p.TotalPages {
		return p.TotalPages
	}

	if p.CurrentPage != 1 {
		return p.CurrentPage - 1
	}

	return 0
}

// totalPages returns the total number of pages.
func (p *paginationOffset) totalPages() int {

	if p.Total == 0 || p.Limit == -1 {
		return 1
	}

	return int(math.Ceil(float64(p.Total) / float64(p.Limit)))
}

// generate is setting the pagination to the controller data.
func (p *paginationOffset) generate(g *Grid, c *sqlquery.Condition) error {

	if c == nil {
		c = &sqlquery.Condition{}
	}

	count, err := g.src.Count(c)
	if err != nil {
		return err
	}
	p.Limit = p.paginationParam(g, "limit")
	p.Total = count
	p.TotalPages = p.totalPages()
	p.CurrentPage = p.paginationParam(g, "page")
	p.Next = p.next()
	p.Prev = p.prev()

	g.controller.Set("pagination", p)

	if p.Limit != -1 {
		c.Limit(p.Limit).Offset(p.offset())
	}

	return nil
}

// paginationParam is checking the request param limit and page.
func (p *paginationOffset) paginationParam(g *Grid, q string) int {

	var param []string
	var err error
	var rv int

	switch q {
	case "limit":
		param, err = g.controller.Context().Request.Param("limit")
		rv = PaginationDefaultLimit
	case "page":
		param, err = g.controller.Context().Request.Param("page")
		rv = 1
	}

	if err == nil {
		s, err := strconv.Atoi(param[0])
		if err == nil {
			rv = s
		}
	}
	return rv
}
