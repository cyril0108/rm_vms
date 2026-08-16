package dto

import (
	"encoding/json"
	"nvr_core/db/models"
	"time"
)

// ================================
// Requests (Incoming from Vue UI)
// ================================

type LayoutRequest struct {
	ID      int64               `json:"id,omitempty"` // Present during Update, omitted during Create
	Name    string              `json:"name"`
	Mode    string              `json:"mode"`
	Payload json.RawMessage     `json:"payload"`
	Items   []LayoutItemRequest `json:"items"`
}

type LayoutItemRequest struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type LayoutPartialUpdateRequest struct {
	Name    *string               `json:"name"`
	Mode    *string               `json:"mode"`
	Payload json.RawMessage       `json:"payload"` // nil if omitted, []byte if present
	Items   *[]LayoutItemRequest  `json:"items"`   // nil if omitted, pointer to array if present
}

// ================================
// Results (Outgoing to Vue UI)
// ================================

type LayoutResult struct {
	ID        int64              `json:"id"`
	UserID    int64              `json:"user_id"`
	Name      string             `json:"name"`
	Mode      string             `json:"mode"`
	Payload   json.RawMessage    `json:"payload"`
	Items     []LayoutItemResult `json:"items"`
	CreatedAt time.Time          `json:"created_at"`
}

type LayoutItemResult struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// ================================
// LayoutRequest Mapping
// ================================

// AsModel converts the incoming API request into the database model.
// Note: We explicitly require the userID here so that the handler injects it 
// securely from the JWT, rather than trusting the frontend request.
func (lr *LayoutRequest) AsModel(userID int64) *models.Layout {
	layout := &models.Layout{
		ID:      lr.ID, // Retain ID in case this is an Update request
		UserID:  userID,
		Name:    lr.Name,
		Mode:    lr.Mode,
		Payload: lr.Payload,
	}

	// Map the nested items
	if len(lr.Items) > 0 {
		layout.Items = make([]models.LayoutItem, len(lr.Items))
		for i, item := range lr.Items {
			layout.Items[i] = models.LayoutItem{
				Type:    item.Type,
				Payload: item.Payload,
			}
		}
	}

	return layout
}

// ================================
// LayoutResult Mapping
// ================================

// LoadFrom maps the raw database model into a clean API response struct.
func (lr *LayoutResult) LoadFrom(layout *models.Layout) {
	lr.ID = layout.ID
	lr.UserID = layout.UserID
	lr.Name = layout.Name
	lr.Mode = layout.Mode
	lr.Payload = layout.Payload
	lr.CreatedAt = layout.CreatedAt

	// Map the nested items safely
	if len(layout.Items) > 0 {
		lr.Items = make([]LayoutItemResult, len(layout.Items))
		for i, item := range layout.Items {
			lr.Items[i] = LayoutItemResult{
				Type:    item.Type,
				Payload: item.Payload,
			}
		}
	} else {
		// Prevent null slices in JSON by initializing an empty array.
		// Vue.js developers prefer `[]` over `null` to avoid "cannot read properties of undefined" errors.
		lr.Items = make([]LayoutItemResult, 0)
	}
}