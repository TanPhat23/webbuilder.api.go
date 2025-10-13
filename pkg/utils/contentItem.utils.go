package utils

import "my-go-app/internal/models"

// FlattenContentItem converts a ContentItem to a flattened map for API response
func FlattenContentItem(ci *models.ContentItem) map[string]any {
	result := map[string]any{
		"id":            ci.Id,
		"title":         ci.Title,
		"slug":          ci.Slug,
		"published":     ci.Published,
		"createdAt":     ci.CreatedAt,
		"updatedAt":     ci.UpdatedAt,
		"contentTypeId": ci.ContentTypeId,
		"contentType":   ci.ContentType,
	}
	for _, fv := range ci.FieldValues {
		if fv.Field.Name != "" {
			result[fv.Field.Name] = fv.Value
		}
	}
	return result
}

// FlattenContentItems converts a slice of ContentItem to flattened maps
func FlattenContentItems(cis []models.ContentItem) []map[string]any {
	result := make([]map[string]any, len(cis))
	for i, ci := range cis {
		result[i] = FlattenContentItem(&ci)
	}
	return result
}
