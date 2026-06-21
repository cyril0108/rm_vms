package service

import (
	"context"
	"errors"

	"nvr_core/apiserver/dto"
	"nvr_core/db/models"
	"nvr_core/db/repository"
	"nvr_core/hardware"
	"nvr_core/security"

)

var ErrLicenseMachineNotMatch = errors.New("Machine id does not match.")
var ErrLicenseExpired = errors.New("License is expired.")


type LicenseService interface {
	ProcessLicenses(ctx context.Context, licenses[]string) (dto.LicenseProcessResult, []*models.License)

	CreateLicense(ctx context.Context, lic *models.License) (error)
	GetAllLicenses(ctx context.Context) ([]*models.License, error)
	GetValidLicenses(ctx context.Context) ([]*models.License, error)
}

func NewLicenseService(repo repository.LicenseRepository) LicenseService {
	return &licenseServiceBase{repo: repo}
}

func (s *licenseServiceBase) CreateLicense(ctx context.Context, lic *models.License) (error) {
	// if s.IsValidLicense()
	return s.repo.Create(ctx, lic)
}

func (s *licenseServiceBase) GetAllLicenses(ctx context.Context) ([]*models.License, error) {
	return s.repo.GetAll(ctx)
}

func (s *licenseServiceBase) GetValidLicenses(ctx context.Context) ([]*models.License, error) {
	return s.repo.GetValidLicenses(ctx)
}

// Process the licenses and check their info. Accepted licenses will be stored in database.
func (s *licenseServiceBase) ProcessLicenses(ctx context.Context, licenses[]string) (dto.LicenseProcessResult, []*models.License) {

	ll := LOG.Prefix("[ProcessLicenses]")

	var acceptedLics []*models.License
	var result dto.LicenseProcessResult
	machine := hardware.GetPersistentMachineID()

	for _, lic := range licenses {

		claims, err := security.GetLicenseInfo(lic)
		if err != nil {

			result.Rejected = append(result.Rejected, &dto.LicenseResult{
				Token: lic,
				Error: err.Error(),
			})

		} else {

			lice := &models.License{
				RawToken: lic,
			}

			lice.LoadClaims(&claims)

			valid, err := IsValidLicense(machine, lice);
			if valid {

				r :=  &dto.LicenseResult{
					Token: lic,
					Claims: &claims,
				}

				if err != nil {
					r.Error = err.Error()
				}

				result.Accepted = append(result.Accepted, r)


				if err := s.CreateLicense(ctx, lice); err != nil {

					ll.Error("Error when add license to database", "error", err)

				} else {

					acceptedLics = append(acceptedLics, lice)

				}

			} else {

				result.Rejected = append(result.Rejected, &dto.LicenseResult{
					Token: lic,
					Error: err.Error(),
				})

			}

		}

	}

	return result, acceptedLics

}

func IsValidLicense(machine string, lic *models.License) (bool, error) {

	if lic.MachineID != machine {
		return false, ErrLicenseMachineNotMatch
	}

	return true, nil

}

