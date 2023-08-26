package appstoreconnect

// AppInfo ...
type AppInfo struct {
	IssuerID string
	KeyID    string
	Token    string
	ExpireAt int64
}

// Paging ...
type Paging struct {
	Total int `json:"total,omitempty"`
	Limit int `json:"limit,omitempty"`
}

// PagingInformation ...
type PagingInformation struct {
	Paging Paging `json:"paging,omitempty"`
}

// DocumentLinks ...
type DocumentLinks struct {
	Self string `json:"self,omitempty"`
}

// PagedDocumentLinks ...
type PagedDocumentLinks struct {
	First string `json:"first,omitempty"`
	Next  string `json:"next,omitempty"`
	Self  string `json:"self,omitempty"`
}

// RelationshipLinks ...
type RelationshipLinks struct {
	Self    string `json:"self,omitempty"`
	Related string `json:"related,omitempty"`
}

// ResourceLinks ...
type ResourceLinks struct {
	Self string `json:"self,omitempty"`
}

// RelationshipData ...
type RelationshipData struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
}

// Relationship ...
type Relationship struct {
	Data  RelationshipData  `json:"data,omitempty"`
	Links RelationshipLinks `json:"links,omitempty"`
	Meta  PagingInformation `json:"meta,omitempty"`
}

// RelationshipArr ...
type RelationshipArr struct {
	Data  []RelationshipData `json:"data,omitempty"`
	Links RelationshipLinks  `json:"links,omitempty"`
	Meta  PagingInformation  `json:"meta,omitempty"`
}

// ImageAsset ...
type ImageAsset struct {
	TemplateURL string `json:"templateUrl,omitempty"`
	Height      int    `json:"height,omitempty"`
	Width       int    `json:"width,omitempty"`
}

// LinkagesRequest ...
type LinkagesRequest struct {
	Data []RelationshipData `json:"data,omitempty"`
}
