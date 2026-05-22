package model

type KnowledgeObject struct {
	ID          string  `json:"id"`
	SpaceID     string  `json:"spaceId"`
	TypeID      string  `json:"typeId"`
	Title       string  `json:"title"`
	Properties  string  `json:"properties"`
	Tags        string  `json:"tags"`
	CoverImage  *string `json:"coverImage"`
	Source      string  `json:"source"`
	SourceMeta  string  `json:"sourceMeta"`
	EmbeddingID *string `json:"embeddingId"`
	WordCount   int     `json:"wordCount"`
	Version     int     `json:"version"`
	IsArchived  bool    `json:"isArchived"`
	IsDeleted   bool    `json:"isDeleted"`
	LastReadAt  *int64  `json:"lastReadAt"`
	SyncStatus  string  `json:"syncStatus"`
	CreatedAt   int64   `json:"createdAt"`
	UpdatedAt   int64   `json:"updatedAt"`
}

type Block struct {
	ID          string `json:"id"`
	ObjectID    string `json:"objectId"`
	ParentID    *string `json:"parentId"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	Properties  string `json:"properties"`
	Position    float64 `json:"position"`
	Depth       int    `json:"depth"`
	Collapsed   bool   `json:"collapsed"`
	Color       string `json:"color"`
	Version     int    `json:"version"`
	SyncStatus  string `json:"syncStatus"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type ObjectType struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Color       string `json:"color"`
	Description string `json:"description"`
	IsBuiltin   bool   `json:"isBuiltin"`
	IsArchived  bool   `json:"isArchived"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type Tag struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	ParentID    *string `json:"parentId"`
	AIGenerated bool   `json:"aiGenerated"`
	ObjectCount int    `json:"objectCount"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type Relation struct {
	ID                string  `json:"id"`
	SourceID          string  `json:"sourceId"`
	TargetID          string  `json:"targetId"`
	Type              string  `json:"type"`
	CustomTypeID      *string `json:"customTypeId"`
	Weight            float64 `json:"weight"`
	Metadata          string  `json:"metadata"`
	AIGenerated       bool    `json:"aiGenerated"`
	SourceObjectType  string  `json:"sourceObjectType"`
	TargetObjectType  string  `json:"targetObjectType"`
	CreatedAt         int64   `json:"createdAt"`
}

// ObjectFilter 对象列表筛选参数
type ObjectFilter struct {
	SpaceID  string
	TypeID   string
	Query    string
	Limit    int
	Offset   int
	Archived *bool
}

// ─── 集合 ────────────────────────────────────────────────

type Collection struct {
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

type CollectionView struct {
	ID            string `json:"id"`
	CollectionID  string `json:"collectionId"`
	Name          string `json:"name"`
	ViewType      string `json:"viewType"`
	Filters       string `json:"filters"`
	Sorts         string `json:"sorts"`
	VisibleFields string `json:"visibleFields"`
	GroupBy       string `json:"groupBy"`
	CalendarField string `json:"calendarField"`
	KanbanField   string `json:"kanbanField"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
}

type CollectionItem struct {
	ID           string  `json:"id"`
	CollectionID string  `json:"collectionId"`
	ObjectID     string  `json:"objectId"`
	Position    float64 `json:"position"`
	Note        string  `json:"note"`
}
