package query

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

type PageRequest struct {
	Number int
	Size   int
}

func ParsePageFromGin(c *gin.Context) Page {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("limit", "10")

	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)

	return NewPage(page, size)
}

func Execute[DB any, Domain any](
	q *Query[DB],
	page Page,
	converter func(*DB) (*Domain, error),
) (*Result[*Domain], error) {
	return Paginate(q, page, converter)
}

func ExecuteAll[DB any, Domain any](
	q *Query[DB],
	converter func(*DB) (*Domain, error),
) ([]*Domain, error) {
	rows, err := q.Find()
	if err != nil {
		return nil, err
	}

	items := make([]*Domain, 0, len(rows))
	for i := range rows {
		item, err := converter(&rows[i])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}

func ExecuteWithLimit[DB any, Domain any](
	q *Query[DB],
	limit int,
	converter func(*DB) (*Domain, error),
) ([]*Domain, error) {
	rows, err := q.FindWithLimit(limit)
	if err != nil {
		return nil, err
	}

	items := make([]*Domain, 0, len(rows))
	for i := range rows {
		item, err := converter(&rows[i])
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
