package dto

import (
	"my-go-app/internal/models"
	"time"
)

type CreateWebhookRequest struct {
	Name          string   `json:"name"          validate:"required,min=1,max=255"`
	Url           string   `json:"url"           validate:"required,url,max=2048"`
	Events        []string `json:"events"        validate:"required,min=1"`
	ContentTypeId *string  `json:"contentTypeId" validate:"omitempty,min=1"`
	Enabled       *bool    `json:"enabled"`
}

type UpdateWebhookRequest struct {
	Name          *string   `json:"name"          validate:"omitempty,min=1,max=255"`
	Url           *string   `json:"url"           validate:"omitempty,url,max=2048"`
	Events        *[]string `json:"events"        validate:"omitempty,min=1"`
	ContentTypeId *string   `json:"contentTypeId"`
	Enabled       *bool     `json:"enabled"`
}

type WebhookResponse struct {
	Id            string    `json:"id"`
	ProjectId     string    `json:"projectId"`
	ContentTypeId *string   `json:"contentTypeId,omitempty"`
	Name          string    `json:"name"`
	Url           string    `json:"url"`
	Secret        string    `json:"secret,omitempty"`
	Events        []string  `json:"events"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

func ToWebhookResponse(w models.ProjectWebhook, includeSecret bool) WebhookResponse {
	secret := "whsec_****"
	if includeSecret {
		secret = w.Secret
	}
	events := []string(w.Events)
	if events == nil {
		events = []string{}
	}
	return WebhookResponse{
		Id:            w.Id,
		ProjectId:     w.ProjectId,
		ContentTypeId: w.ContentTypeId,
		Name:          w.Name,
		Url:           w.Url,
		Secret:        secret,
		Events:        events,
		Enabled:       w.Enabled,
		CreatedAt:     w.CreatedAt,
		UpdatedAt:     w.UpdatedAt,
	}
}
