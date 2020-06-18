package grid

import (
	"github.com/patrickascher/gofw/middleware/jwt"
	"github.com/patrickascher/gofw/orm"
	"github.com/patrickascher/gofw/sqlquery"
	"github.com/patrickascher/skeleton-application/backend/model/auth"
)

type UserGrid struct {
	orm.Model
	ID     int
	GridID string
	UserID string

	Name    string
	GroupBy orm.NullString

	Filters   []UserGridFilter
	Sorting   []UserGridSort
	Positions []UserGridPosition
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
	Desc bool
}

type UserGridPosition struct {
	orm.Model
	ID         int
	UserGridID int

	Key string
	Pos int
}

func getFilterList(g *Grid) (*UserGrid, error) {

	g.controller.Context().Request.Token()
	claim := g.Controller().Context().Request.Raw().Context().Value(jwt.CLAIM).(*auth.Claim)

	userGrid := &UserGrid{}
	err := userGrid.Init(userGrid)
	if err != nil {
		return nil, err
	}

	err = userGrid.First(sqlquery.NewCondition().Where("user_id = ? AND grid_id = ?", claim.UserID, g.gridID()))
	if err != nil {
		return nil, err
	}

	return userGrid, err
}
