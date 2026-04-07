package services

import (
	"encoding/json"
	"errors"
	"quokkaq-go-backend/internal/models"
	"quokkaq-go-backend/internal/repository"

	"github.com/google/uuid"
)

type UnitService interface {
	CreateUnit(unit *models.Unit) error
	GetAllUnits() ([]models.Unit, error)
	GetUnitByID(id string) (*models.Unit, error)
	UpdateUnit(unit *models.Unit) error
	DeleteUnit(id string) error
	AddMaterial(material *models.UnitMaterial) error
	GetMaterials(unitID string) ([]models.UnitMaterial, error)
	DeleteMaterial(id string) error
	UpdateAdSettings(unitID string, settings map[string]interface{}) error
}

type unitService struct {
	repo repository.UnitRepository
}

func NewUnitService(repo repository.UnitRepository) UnitService {
	return &unitService{repo: repo}
}

func (s *unitService) CreateUnit(unit *models.Unit) error {
	if unit.CompanyID == "" {
		// Check if this is the first unit
		count, err := s.repo.Count()
		if err != nil {
			return err
		}

		if count == 0 {
			// Auto-create company
			company := &models.Company{
				ID:   uuid.New().String(),
				Name: "Default Company",
			}
			if err := s.repo.CreateCompany(company); err != nil {
				return err
			}
			unit.CompanyID = company.ID
		} else {
			return errors.New("companyId is required")
		}
	}
	return s.repo.Create(unit)
}

func (s *unitService) GetAllUnits() ([]models.Unit, error) {
	return s.repo.FindAll()
}

func (s *unitService) GetUnitByID(id string) (*models.Unit, error) {
	return s.repo.FindByID(id)
}

func (s *unitService) UpdateUnit(unit *models.Unit) error {
	return s.repo.Update(unit)
}

func (s *unitService) DeleteUnit(id string) error {
	return s.repo.Delete(id)
}

func (s *unitService) AddMaterial(material *models.UnitMaterial) error {
	return s.repo.AddMaterial(material)
}

func (s *unitService) GetMaterials(unitID string) ([]models.UnitMaterial, error) {
	return s.repo.GetMaterials(unitID)
}

func (s *unitService) DeleteMaterial(id string) error {
	return s.repo.DeleteMaterial(id)
}

func (s *unitService) UpdateAdSettings(unitID string, settings map[string]interface{}) error {
	unit, err := s.repo.FindByIDLight(unitID)
	if err != nil {
		return err
	}

	// Update config with ad settings
	// Assuming settings is the full ad settings object
	// We might need to merge it if Config contains other things.
	// For now, let's assume Config IS the ad settings or contains it.
	// If Config is jsonb, we can marshal settings to bytes.

	// TODO: Better handling of Config structure.
	// For now, just overwrite Config with new settings.

	bytes, err := json.Marshal(settings)
	if err != nil {
		return err
	}
	unit.Config = json.RawMessage(bytes)

	return s.repo.Update(unit)
}
