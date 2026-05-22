package dto

// ─── Tag ──────────────────────────────────────────────────

type CreateTagRequest struct {
	Name  string  `json:"name"`
	Color string  `json:"color"`
	ParentID *string `json:"parentId"`
}

type UpdateTagRequest struct {
	Name     *string `json:"name"`
	Color    *string `json:"color"`
	ParentID *string `json:"parentId"`
}

type TagResponse struct {
	ID          string `json:"id"`
	SpaceID     string `json:"spaceId"`
	Name        string `json:"name"`
	Color       string `json:"color"`
	ParentID    *string `json:"parentId"`
	ObjectCount int    `json:"objectCount"`
	CreatedAt   int64  `json:"createdAt"`
	UpdatedAt   int64  `json:"updatedAt"`
}

type AssignTagsRequest struct {
	TagIDs []string `json:"tagIds"`
}
