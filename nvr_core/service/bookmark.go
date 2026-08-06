package service

import (
	"context"

	"nvr_core/db/models"
	"nvr_core/db/repository"
)


type BookmarkService interface {
	GetBookmarksBetween(ctx context.Context, start, end int64) ([]*models.Bookmark, error)
	GetCameraBookmarksBetween(ctx context.Context, camID, start, end int64) ([]*models.Bookmark, error)
	AddBookmark(ctx context.Context, bookmark *models.Bookmark) (*models.Bookmark, error)
	UpdateBookmark(ctx context.Context, bookmark *models.Bookmark) (error)
	Delete(ctx context.Context, id int64) (error)
}

func NewBookmarkService(repo repository.BookmarkRepository) BookmarkService {
	return &bookmarkServiceBase{repo: repo}
}

func (bm *bookmarkServiceBase) AddBookmark(ctx context.Context, bookmark *models.Bookmark) (*models.Bookmark, error) {
	id, err := bm.repo.Create(ctx, bookmark)
	if err != nil {
		return nil, err
	}
	bookmark.ID = id
	return bookmark, nil
}

// Update bookmark, only allows changes for duration and message
func (e *bookmarkServiceBase) UpdateBookmark(ctx context.Context, bookmark *models.Bookmark) (error) {
	return e.repo.Update(ctx, bookmark)
}

func (e *bookmarkServiceBase) Delete(ctx context.Context, id int64) (error) {
	return e.repo.Delete(ctx, id)
}

func (e *bookmarkServiceBase) GetBookmarksBetween(ctx context.Context, start, end int64) ([]*models.Bookmark, error) {
	return e.repo.GetByTimeRange(ctx, start, end)
}

func (e *bookmarkServiceBase) GetCameraBookmarksBetween(ctx context.Context, camID, start, end int64) ([]*models.Bookmark, error) {
	return e.repo.GetByCamera(ctx, camID, start, end)
}
