package service

import (
	"context"
	"errors"

	"nvr_core/apiserver/dto"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)

var (
	ErrLayoutNotFound = errors.New("Layout does not exist")
)

type LayoutService interface {
	GetUserLayouts(ctx context.Context, userID int64) ([]*models.Layout, error)
	GetLayout(ctx context.Context, userID int64, layout int64) (*models.Layout, error)

	AddLayout(ctx context.Context, userID int64, layout *models.Layout) (*models.Layout, error)
	UpdateLayout(ctx context.Context, userID int64, layout *models.Layout) (error)
	UpdatePartial(ctx context.Context, userID int64, layoutID int64, req *dto.LayoutPartialUpdateRequest) error
	Delete(ctx context.Context, userID, id int64) (error)
}

func NewLayoutService(repo repository.LayoutRepository) LayoutService {
	return &layoutServiceBase{repo: repo}
}

func (lo *layoutServiceBase) GetUserLayouts(ctx context.Context, userID int64) ([]*models.Layout, error) {
	return lo.repo.GetByUser(ctx, userID)
}

func (lo *layoutServiceBase) GetLayout(ctx context.Context, userID int64, layoutID int64) (*models.Layout, error) {
	return lo.repo.GetLayout(ctx, userID, layoutID)
}


func (lo *layoutServiceBase) AddLayout(ctx context.Context, userID int64, layout *models.Layout) (*models.Layout, error) {
	id, err := lo.repo.Create(ctx, userID, layout)
	if err != nil {
		return nil, err
	}
	layout.ID = id
	return layout, nil
}

// Update layout, only allows changes for duration and message
func (e *layoutServiceBase) UpdateLayout(ctx context.Context, userID int64, layout *models.Layout) (error) {
	return e.repo.Update(ctx, userID, layout)
}

func (lo *layoutServiceBase) UpdatePartial(ctx context.Context, userID int64, layoutID int64, req *dto.LayoutPartialUpdateRequest) error {

	theLayout, err := lo.GetLayout(ctx, userID, layoutID)
	if err != nil {
		return ErrLayoutNotFound
	}

	updateItems := false

	LOG.Info("[UpdatePartial] ", "req", req)

	if req.Name != nil {
		theLayout.Name = *req.Name
	}
	if req.Mode != nil {
		theLayout.Mode = *req.Mode
	}
	if req.Payload != nil {
		theLayout.Payload = req.Payload
	}

	if req.Items != nil {
		updateItems = true

		// Map the DTO items to Database Models
		theLayout.Items = make([]models.LayoutItem, len(*req.Items))
		for i, itemReq := range *req.Items {
			theLayout.Items[i] = models.LayoutItem{
				Type:    itemReq.Type,
				Payload: itemReq.Payload,
			}
		}
	}

	return lo.repo.UpdatePartial(ctx, userID, theLayout, updateItems)

}

func (lo *layoutServiceBase) Delete(ctx context.Context, userID, id int64) (error) {
	return lo.repo.Delete(ctx, userID, id)
}
