package dto

type CreateCollectionRequest struct {
	Name        string `json:"name"`
	SourceType  string `json:"sourceType"`
	SourceConfig string `json:"sourceConfig"`
	Layout      string `json:"layout"`
}

type UpdateCollectionRequest struct {
	Name        *string `json:"name"`
	SourceType  *string `json:"sourceType"`
	SourceConfig *string `json:"sourceConfig"`
	Layout      *string `json:"layout"`
}

type CollectionResponse struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	SourceType  string `json:"sourceType"`
	SourceConfig string `json:"sourceConfig"`
	Layout      string `json:"layout"`
	ObjectCount int    `json:"objectCount"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

// CollectionView

type CreateCollectionViewRequest struct {
	Name          string `json:"name"`
	ViewType      string `json:"viewType"`
	Filters       string `json:"filters"`
	Sorts         string `json:"sorts"`
	VisibleFields string `json:"visibleFields"`
	GroupBy       string `json:"groupBy"`
	CalendarField string `json:"calendarField"`
	KanbanField   string `json:"kanbanField"`
}

type UpdateCollectionViewRequest struct {
	Name          *string `json:"name"`
	ViewType      *string `json:"viewType"`
	Filters       *string `json:"filters"`
	Sorts         *string `json:"sorts"`
	VisibleFields *string `json:"visibleFields"`
	GroupBy       *string `json:"groupBy"`
	CalendarField *string `json:"calendarField"`
	KanbanField   *string `json:"kanbanField"`
}

// CollectionItem

type AddCollectionItemRequest struct {
	ObjectID string  `json:"objectId"`
	Position float64 `json:"position"`
	Note     string  `json:"note"`
}
