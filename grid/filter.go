package grid

import (
	"database/sql"
	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/skeleton-application/backend/model/auth"
)

type UserGrid struct {
	orm.Model
	ID     int
	GridID string
	UserID int

	Name    string
	GroupBy orm.NullString

	Filters []UserGridFilter
	Sorting []UserGridSort
	Fields  []UserGridField

	Default     bool
	RowsPerPage orm.NullInt
}

type UserGridFilter struct {
	orm.Model
	ID         int
	UserGridID int

	Key   string
	Op    string
	Value string
}

type UserGridSort struct {
	orm.Model
	ID         int
	UserGridID int

	Key  string
	Pos  orm.NullInt // because 0 should be allowed as well. TODO figure out a better solution
	Desc bool
}

type UserGridField struct {
	orm.Model
	ID         int
	UserGridID int

	Key  string
	Pos  orm.NullInt
	Show bool
}

type FeGridFilter struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type FeGridActive struct {
	ID          int      `json:"id,omitempty"`
	RowsPerPage int      `json:"rowsPerPage,omitempty"`
	Sort        []string `json:"sort,omitempty"`
	Group       string   `json:"group,omitempty"`
}

func filterBase(g *Grid) (*UserGrid, int, error) {
	g.controller.Context().Request.Token()
	claim := g.Controller().Context().Request.Raw().Context().Value(jwt.CLAIM).(*auth.Claim)

	userGrid := &UserGrid{}
	err := userGrid.Init(userGrid)
	if err != nil {
		return nil, 0, err
	}

	return userGrid, claim.UserID, nil
}

func getFilterByID(id int, g *Grid) (*UserGrid, error) {
	userGrid, userID, err := filterBase(g)
	if err != nil {
		return nil, err
	}
	userGrid.SetRelationCondition("Sorting", *sqlquery.NewCondition().Where("user_grid_id = ?", id).Order("pos ASC"))
	err = userGrid.First(sqlquery.NewCondition().Where("id = ? AND user_id = ? AND grid_id = ?", id, userID, g.gridID()))
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	return userGrid, nil
}

func getFilterList(g *Grid) ([]FeGridFilter, error) {
	userGrid, userID, err := filterBase(g)
	if err != nil {
		return nil, err
	}

	var res []UserGrid
	userGrid.SetWBList(orm.WHITELIST, "ID", "Name")
	err = userGrid.All(&res, sqlquery.NewCondition().Where("user_id = ? AND grid_id = ?", userID, g.gridID()))
	if err != nil {
		return nil, err
	}

	var rv []FeGridFilter
	for _, row := range res {
		rv = append(rv, FeGridFilter{ID: row.ID, Name: row.Name})
	}

	return rv, nil
}
