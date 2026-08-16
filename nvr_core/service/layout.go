package service

import (
	"context"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)


type LayoutService interface {
	GetUserLayouts(ctx context.Context, userID int64) ([]*models.Layout, error)
	GetLayout(ctx context.Context, userID int64, layout int64) (*models.Layout, error)

	AddLayout(ctx context.Context, userID int64, layout *models.Layout) (*models.Layout, error)
	UpdateLayout(ctx context.Context, userID int64, layout *models.Layout) (error)
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


func (bm *layoutServiceBase) AddLayout(ctx context.Context, userID int64, layout *models.Layout) (*models.Layout, error) {
	id, err := bm.repo.Create(ctx, userID, layout)
	if err != nil {
		return nil, err
	}
	layout.ID = id
	return layout, nil
}

// Update bookmark, only allows changes for duration and message
func (e *layoutServiceBase) UpdateLayout(ctx context.Context, userID int64, layout *models.Layout) (error) {
	return e.repo.Update(ctx, userID, layout)
}

func (e *layoutServiceBase) Delete(ctx context.Context, userID, id int64) (error) {
	return e.repo.Delete(ctx, userID, id)
}
