package service

import (
	"context"
	"database/sql"
	"errors"

	"nvr_core/apiserver/dto"
	// "nvr_core/db/models"
	"nvr_core/db/repository"
)

const SystemAdminUser = "admin"

type SystemService interface {
	GetDebugData(ctx context.Context) (dto.SystemDebugInfo, error)
	GetHealthData(ctx context.Context) (dto.SystemHealthInfo, error)
}


func NewSystemService(db *sql.DB, repo repository.UserRepository) SystemService {
	return &systemServiceBase{db: db, repo: repo}
}

func (s *systemServiceBase) GetHealthData(ctx context.Context) (dto.SystemHealthInfo, error) {

	health := dto.SystemHealthInfo{
		Health: "Ok",
		Configured: true,
	}

	user, err := s.repo.GetAdmin(ctx)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			health.Configured = false
			return health, nil
		}
	}

	if user.Password == "" {
		health.Configured = false
	}

	LOG.Info("[GetHealthData] ", "configured", health.Configured)

	return health, nil

}

func (s *systemServiceBase) GetDebugData(ctx context.Context) (dto.SystemDebugInfo, error) {

	// Query SQLite internal stats
	var totalSegments int
	var dbSizeBytes int64

	// Count total rows
	s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM segments").Scan(&totalSegments)

	// Check the physical file size of the database itself using SQLite pragmas
	s.db.QueryRowContext(ctx, "SELECT page_count * page_size FROM pragma_page_count(), pragma_page_size()").Scan(&dbSizeBytes)

	// segment, err := s.repo.GetByID(ctx)
	// if err != nil {
	// 	return dto.SystemDebugInfo{}, err
	// }

	// if segment == nil {
	// 	segment = &models.Segment{}
	// }

	// Start the first block
	data := dto.SystemDebugInfo{
		Status: "online",
		TotalSegments: totalSegments,
		DbSize: float64(dbSizeBytes) / 1024 / 1024,
		// LastSegment: dto.NewSegmentItemFrom(segment),
	}

	return data, nil
}