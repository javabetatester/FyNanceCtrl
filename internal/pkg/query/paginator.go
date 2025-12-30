package query

type Page struct {
	Number int
	Size   int
}

func (p *Page) offset() int {
	if p.Number < 1 {
		p.Number = 1
	}
	return (p.Number - 1) * p.Size
}

func (p *Page) normalize() {
	if p.Number < 1 {
		p.Number = 1
	}
	if p.Size < 1 {
		p.Size = 10
	}
	if p.Size > 100 {
		p.Size = 100
	}
}

type Result[T any] struct {
	Data       []T   `json:"data"`
	Page       int   `json:"page"`
	Size       int   `json:"size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func (r *Result[T]) HasNext() bool {
	return r.Page < r.TotalPages
}

func (r *Result[T]) HasPrev() bool {
	return r.Page > 1
}

func Paginate[DBModel any, Domain any](
	q *Query[DBModel],
	page Page,
	converter func(*DBModel) (*Domain, error),
) (*Result[*Domain], error) {
	page.normalize()

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	db := q.DB()
	if q.OrderBy() != "" {
		db = db.Order(q.OrderBy())
	}

	var rows []DBModel
	err = db.Offset(page.offset()).Limit(page.Size).Find(&rows).Error
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

	totalPages := int(total) / page.Size
	if int(total)%page.Size > 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return &Result[*Domain]{
		Data:       items,
		Page:       page.Number,
		Size:       page.Size,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func PaginateRaw[T any](q *Query[T], page Page) (*Result[T], error) {
	page.normalize()

	total, err := q.Count()
	if err != nil {
		return nil, err
	}

	db := q.DB()
	if q.OrderBy() != "" {
		db = db.Order(q.OrderBy())
	}

	var rows []T
	err = db.Offset(page.offset()).Limit(page.Size).Find(&rows).Error
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / page.Size
	if int(total)%page.Size > 0 {
		totalPages++
	}
	if totalPages == 0 {
		totalPages = 1
	}

	return &Result[T]{
		Data:       rows,
		Page:       page.Number,
		Size:       page.Size,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

func DefaultPage() Page {
	return Page{Number: 1, Size: 10}
}

func NewPage(number, size int) Page {
	p := Page{Number: number, Size: size}
	p.normalize()
	return p
}
