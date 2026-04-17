package model

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	AffiliateStatusPending  = "pending"
	AffiliateStatusApproved = "approved"
	AffiliateStatusRejected = "rejected"
)

type AffiliateApplication struct {
	Id             int    `json:"id"`
	Name           string `json:"name" gorm:"type:varchar(64);not null"`
	Email          string `json:"email" gorm:"type:varchar(128);not null;uniqueIndex"`
	Country        string `json:"country" gorm:"type:varchar(64)"`
	Phone          string `json:"phone" gorm:"type:varchar(32)"`
	Instagram      string `json:"instagram" gorm:"type:varchar(128)"`
	Tiktok         string `json:"tiktok" gorm:"type:varchar(128)"`
	Youtube        string `json:"youtube" gorm:"type:varchar(256)"`
	OtherSocial    string `json:"other_social" gorm:"type:varchar(256)"`
	Status         string `json:"status" gorm:"type:varchar(16);default:'pending';index"`
	KolInviteToken *string `json:"kol_invite_token,omitempty" gorm:"type:varchar(64);uniqueIndex"`
	TokenUsed      bool   `json:"token_used" gorm:"default:false"`
	CreatedAt      int64  `json:"created_at" gorm:"bigint"`
	UpdatedAt      int64  `json:"updated_at" gorm:"bigint"`
}

func (a *AffiliateApplication) BeforeCreate(tx *gorm.DB) error {
	now := common.GetTimestamp()
	a.CreatedAt = now
	a.UpdatedAt = now
	return nil
}

func (a *AffiliateApplication) BeforeUpdate(tx *gorm.DB) error {
	a.UpdatedAt = common.GetTimestamp()
	return nil
}

// GetApplicationByEmail returns the application for a given email, or nil if not found.
func GetApplicationByEmail(email string) (*AffiliateApplication, error) {
	var app AffiliateApplication
	err := DB.Where("email = ?", email).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &app, err
}

// CreateOrUpdateApplication creates a new application or updates an existing one.
// Returns ErrAffiliateAlreadyApproved if the existing record has been approved.
func CreateOrUpdateApplication(app *AffiliateApplication) error {
	existing, err := GetApplicationByEmail(app.Email)
	if err != nil {
		return err
	}
	if existing == nil {
		return DB.Create(app).Error
	}
	if existing.Status == AffiliateStatusApproved {
		return ErrAffiliateAlreadyApproved
	}
	// Update existing record
	app.Id = existing.Id
	app.Status = AffiliateStatusPending
	app.KolInviteToken = existing.KolInviteToken // preserve existing token (nil until approved)
	app.TokenUsed = existing.TokenUsed
	app.CreatedAt = existing.CreatedAt
	return DB.Save(app).Error
}

// GetAllApplications returns a paginated list of applications, optionally filtered by status.
func GetAllApplications(status string, page, pageSize int) ([]AffiliateApplication, int64, error) {
	var apps []AffiliateApplication
	var total int64

	tx := DB.Model(&AffiliateApplication{})
	if status != "" {
		tx = tx.Where("status = ?", status)
	}
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := tx.Order("created_at desc").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&apps).Error; err != nil {
		return nil, 0, err
	}
	return apps, total, nil
}

// ApproveApplication transitions a pending application to approved and generates a one-time KOL invite token.
func ApproveApplication(id int) (*AffiliateApplication, error) {
	var app AffiliateApplication
	if err := DB.First(&app, id).Error; err != nil {
		return nil, err
	}
	if app.Status != AffiliateStatusPending {
		return nil, ErrAffiliateNotPending
	}
	token := uuid.New().String()
	app.Status = AffiliateStatusApproved
	app.KolInviteToken = &token
	if err := DB.Save(&app).Error; err != nil {
		return nil, err
	}
	return &app, nil
}

// RejectApplication transitions a pending application to rejected.
func RejectApplication(id int) error {
	var app AffiliateApplication
	if err := DB.First(&app, id).Error; err != nil {
		return err
	}
	if app.Status != AffiliateStatusPending {
		return ErrAffiliateNotPending
	}
	app.Status = AffiliateStatusRejected
	return DB.Save(&app).Error
}

// GetApplicationByToken returns the application for a given KOL invite token, or nil if not found.
func GetApplicationByToken(token string) (*AffiliateApplication, error) {
	var app AffiliateApplication
	err := DB.Where("kol_invite_token = ?", token).First(&app).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &app, err
}

// MarkTokenUsed marks the invite token as used.
func MarkTokenUsed(token string) error {
	return DB.Model(&AffiliateApplication{}).
		Where("kol_invite_token = ?", token).
		Update("token_used", true).Error
}

var (
	ErrAffiliateAlreadyApproved error = gorm.ErrInvalidData
	ErrAffiliateNotPending      error = gorm.ErrInvalidData
)

func init() {
	// Use distinct sentinel errors
	ErrAffiliateAlreadyApproved = &affiliateError{"申请已通过审核，无法再次修改"}
	ErrAffiliateNotPending      = &affiliateError{"申请状态不是 pending，无法操作"}
}

type affiliateError struct {
	msg string
}

func (e *affiliateError) Error() string { return e.msg }
