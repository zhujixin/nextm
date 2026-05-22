package dto

type CreateRelationRequest struct {
	SourceID     string  `json:"sourceId"`
	TargetID     string  `json:"targetId"`
	Type         string  `json:"type"`
	CustomTypeID *string `json:"customTypeId"`
	Weight      float64 `json:"weight"`
}

type UpdateRelationRequest struct {
	Type         *string  `json:"type"`
	CustomTypeID *string  `json:"customTypeId"`
	Weight      *float64 `json:"weight"`
	Metadata    *string  `json:"metadata"`
}

type RelationResponse struct {
	ID                string  `json:"id"`
	SourceID          string  `json:"sourceId"`
	TargetID          string  `json:"targetId"`
	Type              string  `json:"type"`
	CustomTypeID      *string `json:"customTypeId"`
	Weight           float64 `json:"weight"`
	Metadata         string  `json:"metadata"`
	AIGenerated      bool    `json:"aiGenerated"`
	SourceObjectType string  `json:"sourceObjectType"`
	TargetObjectType string  `json:"targetObjectType"`
	CreatedAt        int64   `json:"createdAt"`
}
