package service

import (
	"context"

	"nvr_core/db/repository"
)

const (
	SetKeyServerName = "server_name"
)

type SystemSettingService interface {
	CreateSetting(ctx context.Context, key string, value string) (error)
	GetSetting(ctx context.Context, key string) (string, error)
	SetSetting(ctx context.Context, key string, value string) error

	GetServerName(ctx context.Context) (string, error)
	SetServerName(ctx context.Context, value string) error
}

func NewSysSettingService(repo repository.SystemSettingsRepository) SystemSettingService {
	return &sysSettingServiceBase{repo: repo}
}

func (ss *sysSettingServiceBase) CreateSetting(ctx context.Context, key string, value string) (error) {
	return ss.repo.Create(ctx, key, value)
}

func (ss *sysSettingServiceBase) GetSetting(ctx context.Context, key string) (string, error) {
	return ss.repo.Get(ctx, key)
}

func (ss *sysSettingServiceBase) SetSetting(ctx context.Context, key string, value string) error {
	return ss.repo.Set(ctx, key, value)
}


func (ss *sysSettingServiceBase) GetServerName(ctx context.Context) (string, error) {
	return ss.repo.Get(ctx, SetKeyServerName)
}

func (ss *sysSettingServiceBase) SetServerName(ctx context.Context, value string) error {
	return ss.repo.Set(ctx, SetKeyServerName, value)
}

