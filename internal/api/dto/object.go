package dto

type CreateObjectRequest struct {
	TypeID    string `json:"typeId" validate:"required"`
	Title     string `json:"title" validate:"required"`
	Tags      string `json:"tags,omitempty"`
	Source    string `json:"source,omitempty"`
	CoverImage string `json:"coverImage,omitempty"`
	Properties string `json:"properties,omitempty"`
}

type UpdateObjectRequest struct {
	Title      *string `json:"title,omitempty"`
	Tags       *string `json:"tags,omitempty"`
	CoverImage *string `json:"coverImage,omitempty"`
	Properties *string `json:"properties,omitempty"`
}

type ObjectResponse struct {
	ID         string `json:"id"`
	SpaceID    string `json:"spaceId"`
	TypeID     string `json:"typeId"`
	Title      string `json:"title"`
	Tags       string `json:"tags"`
	Source     string `json:"source"`
	WordCount  int    `json:"wordCount"`
	Version    int    `json:"version"`
	IsArchived bool   `json:"isArchived"`
	CreatedAt  int64  `json:"createdAt"`
	UpdatedAt  int64  `json:"updatedAt"`
}

type ListObjectsQuery struct {
	TypeID string `json:"typeId,omitempty"`
	Query  string `json:"query,omitempty"`
	Limit  int    `json:"limit,omitempty"`
	Offset int    `json:"offset,omitempty"`
}

type CreateBlockRequest struct {
	ParentID   *string `json:"parentId,omitempty"`
	Type       string  `json:"type" validate:"required"`
	Content    string  `json:"content,omitempty"`
	Properties string  `json:"properties,omitempty"`
	Position   float64 `json:"position"`
	Depth      int     `json:"depth,omitempty"`
	Color      string  `json:"color,omitempty"`
}

type UpdateBlockRequest struct {
	Content    *string  `json:"content,omitempty"`
	Type       *string  `json:"type,omitempty"`
	Position   *float64 `json:"position,omitempty"`
	Depth      *int     `json:"depth,omitempty"`
	Color      *string  `json:"color,omitempty"`
}
