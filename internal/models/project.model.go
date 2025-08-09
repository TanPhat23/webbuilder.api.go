package models

type Project struct {
	ID          string                 `json:"id" bson:"_id"`
	Name        string                 `json:"name" bson:"name"`
	Description *string                 `json:"description" bson:"description"`
	Styles      map[string]any `json:"styles" bson:"styles"`
	Published   bool                   `json:"published" bson:"published" `
	Subdomain   *string                 `json:"subdomain" bson:"subdomain"`
	UserId      string                 `json:"userId" bson:"userId"`
}
