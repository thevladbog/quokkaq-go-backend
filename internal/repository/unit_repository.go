package repository

import (
	"quokkaq-go-backend/internal/models"
	"quokkaq-go-backend/pkg/database"

	"gorm.io/gorm"
)

type UnitRepository interface {
	Create(unit *models.Unit) error
	FindAll() ([]models.Unit, error)
	FindByID(id string) (*models.Unit, error)
	Update(unit *models.Unit) error
	Delete(id string) error
	AddMaterial(material *models.UnitMaterial) error
	GetMaterials(unitID string) ([]models.UnitMaterial, error)
	DeleteMaterial(id string) error
}

type unitRepository struct {
	db *gorm.DB
}

func NewUnitRepository() UnitRepository {
	return &unitRepository{db: database.DB}
}

func (r *unitRepository) Create(unit *models.Unit) error {
	return r.db.Create(unit).Error
}

func (r *unitRepository) FindAll() ([]models.Unit, error) {
	var units []models.Unit
	err := r.db.Find(&units).Error
	return units, err
}

func (r *unitRepository) FindByID(id string) (*models.Unit, error) {
	var unit models.Unit
	err := r.db.Preload("Services").Preload("Counters").Preload("Tickets").First(&unit, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &unit, nil
}

func (r *unitRepository) Update(unit *models.Unit) error {
	return r.db.Save(unit).Error
}

func (r *unitRepository) Delete(id string) error {
	return r.db.Delete(&models.Unit{}, "id = ?", id).Error
}

func (r *unitRepository) AddMaterial(material *models.UnitMaterial) error {
	return r.db.Create(material).Error
}

func (r *unitRepository) GetMaterials(unitID string) ([]models.UnitMaterial, error) {
	var materials []models.UnitMaterial
	err := r.db.Where("unit_id = ?", unitID).Find(&materials).Error
	return materials, err
}

func (r *unitRepository) DeleteMaterial(id string) error {
	return r.db.Delete(&models.UnitMaterial{}, "id = ?", id).Error
}
