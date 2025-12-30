package contracts

// CategoryGroupResponse resposta de grupo de categoria
type CategoryGroupResponse struct {
	Id         string                  `json:"id"`
	Name       string                  `json:"name"`
	Type       string                  `json:"type"`
	Icon       string                  `json:"icon"`
	Color      string                  `json:"color"`
	SortOrder  int                     `json:"sortOrder"`
	Categories []*CategoryTreeResponse `json:"categories,omitempty"`
}

// CategoryTreeResponse resposta de categoria em árvore
type CategoryTreeResponse struct {
	Id            string                  `json:"id"`
	Name          string                  `json:"name"`
	Icon          string                  `json:"icon"`
	Color         string                  `json:"color"`
	GroupId       *string                 `json:"groupId,omitempty"`
	GroupName     string                  `json:"groupName,omitempty"`
	ParentId      *string                 `json:"parentId,omitempty"`
	SortOrder     int                     `json:"sortOrder"`
	Subcategories []*CategoryTreeResponse `json:"subcategories,omitempty"`
}

// CategoryTreeFullResponse árvore completa de categorias
type CategoryTreeFullResponse struct {
	ExpenseGroups []*CategoryGroupResponse `json:"expenseGroups"`
	ReceiptGroups []*CategoryGroupResponse `json:"receiptGroups"`
}
