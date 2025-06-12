package tables

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uuid.UUID `gorm:"type:char(36);primaryKey"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

// GORM hook to create a UUID it it is nil
func (b *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return
}

type Status int

const (
	StatusConcept      Status = iota // Project is at the idea stage, being explored.
	StatusPlanning                   // Project approved, defining scope, timeline, etc.
	StatusReadyToStart               // Planning complete, awaiting formal initiation.

	StatusInProgress // Project is actively running.
	StatusOnHold     // Project is temporarily paused.
	StatusDelayed    // Project is behind schedule.
	StatusIssues     // Serious problems hindering progress.

	StatusCompleted // Project has been finalized.
	StatusCancelled // Project officially stopped before completion.
)

type Users struct {
	BaseModel
	FirstName string `gorm:"size:100"`
	LastName  string `gorm:"size:100"`
	Email     string `gorm:"uniqueIndex"`
	Password  string `gorm:"type:text"`
}

type Companies struct {
	BaseModel
	Name         string    `gorm:"size:100"`
	CreatedByID  uuid.UUID `gorm:"type:char(36);not null"`
	CreatedBy    Users     `gorm:"foreignKey:CreatedByID;references:ID"`
	ModifiedAt   time.Time `gorm:"autoUpdateTime"`
	ModifiedByID uuid.UUID `gorm:"type:char(36);not null"`
	ModifiedBy   Users     `gorm:"foreignKey:ModifiedByID;references:ID"`
}

type Projects struct {
	BaseModel
	Name          string    `gorm:"size:100"`
	Description   string    `gorm:"type:text"`
	Status        Status    `gorm:"type:integer"`
	EstimatedCost uint      `gorm:"type:integer"`
	CreatedByID   uuid.UUID `gorm:"type:char(36);not null"`
	CreatedBy     Users     `gorm:"foreignKey:CreatedByID;references:ID"`
	ModifiedAt    time.Time `gorm:"autoUpdateTime"`
	ModifiedByID  uuid.UUID `gorm:"type:char(36);not null"`
	ModifiedBy    Users     `gorm:"foreignKey:ModifiedByID;references:ID"`
}

type Budgets struct {
	BaseModel
	ProjectID    uuid.UUID `gorm:"type:char(36);not null;index"`
	Project      Projects  `gorm:"foreignKey:ProjectID;references:ID"`
	BudgetUsed   uint      `gorm:"type:integer"`
	CreatedByID  uuid.UUID `gorm:"type:char(36);not null"`
	CreatedBy    Users     `gorm:"foreignKey:CreatedByID;references:ID"`
	ModifiedAt   time.Time `gorm:"autoUpdateTime"`
	ModifiedByID uuid.UUID `gorm:"type:char(36);not null"`
	ModifiedBy   Users     `gorm:"foreignKey:ModifiedByID;references:ID"`
}

type ProjectUsers struct {
	BaseModel
	ProjectID uuid.UUID `gorm:"type:char(36);not null;index"`
	Project   Projects  `gorm:"foreignKey:ProjectID;references:ID"`
	UserID    uuid.UUID `gorm:"type:char(36);not null;index"`
	User      Users     `gorm:"foreignKey:UserID;references:ID"`
}

type CompanyProjects struct {
	BaseModel
	CompanyID uuid.UUID `gorm:"type:char(36);not null;index"`
	Company   Companies `gorm:"foreignKey:CompanyID;references:ID"`
	ProjectID uuid.UUID `gorm:"type:char(36);not null;index"`
	Project   Projects  `gorm:"foreignKey:ProjectID;references:ID"`
}

func RunMigrations(db *gorm.DB) {
	db.AutoMigrate(
		&Users{},
		&Companies{},
		&Projects{},
		&Budgets{},
		&ProjectUsers{},
		&CompanyProjects{},
	)
}
