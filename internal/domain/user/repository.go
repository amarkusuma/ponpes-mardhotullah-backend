package user

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/simpeg/backend/internal/model"
	"github.com/simpeg/backend/internal/model/akademik"
	"github.com/simpeg/backend/internal/model/keuangan"
	"github.com/simpeg/backend/pkg/pagination"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// ── User CRUD ────────────────────────────────────────────────────────────────

func (r *Repository) List(params pagination.Params, filter ListUsersFilter) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	q := r.db.Model(&model.User{}).Preload("Roles")

	if filter.Search != "" {
		q = q.Where("name ILIKE ? OR email ILIKE ?",
			"%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.RoleName != "" {
		q = q.Joins("JOIN user_roles ur ON ur.user_id = users.id").
			Joins("JOIN roles r ON r.id = ur.role_id AND r.name = ?", filter.RoleName)
	}
	if filter.IsActive != nil {
		q = q.Where("is_active = ?", *filter.IsActive)
	}

	q.Count(&total)
	err := q.Order(params.OrderClause()).
		Limit(params.Limit).
		Offset(params.Offset()).
		Find(&users).Error

	return users, total, err
}

func (r *Repository) FindByID(id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.Preload("Roles.Permissions").First(&u, "id = ?", id).Error
	return &u, err
}

func (r *Repository) FindByEmail(email string) (*model.User, error) {
	var u model.User
	err := r.db.Where("email = ?", email).First(&u).Error
	return &u, err
}

func (r *Repository) Create(u *model.User) error {
	return r.db.Create(u).Error
}

func (r *Repository) Update(u *model.User) error {
	return r.db.Save(u).Error
}

func (r *Repository) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.User{}, "id = ?", id).Error
}

func (r *Repository) AssignRoles(userID uuid.UUID, roleIDs []uuid.UUID) error {
	var roles []model.Role
	if err := r.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
		return err
	}
	return r.db.Model(&model.User{Base: model.Base{ID: userID}}).
		Association("Roles").Replace(roles)
}

// ── Roles ────────────────────────────────────────────────────────────────────

func (r *Repository) ListRoles() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Preload("Permissions").Find(&roles).Error
	return roles, err
}

// ── Dashboard ────────────────────────────────────────────────────────────────

func (r *Repository) CountSantri() (int64, error) {
	var count int64
	err := r.db.Model(&akademik.Santri{}).Count(&count).Error
	return count, err
}

func (r *Repository) CountGuru() (int64, error) {
	var count int64
	err := r.db.Model(&akademik.Guru{}).Count(&count).Error
	return count, err
}

func (r *Repository) CountWali() (int64, error) {
	var count int64
	err := r.db.Model(&akademik.Wali{}).Count(&count).Error
	return count, err
}

func (r *Repository) TagihanStatsBulanIni() (total float64, lunas, belum int64, err error) {
	now := time.Now()
	bulan, tahun := int(now.Month()), now.Year()

	r.db.Model(&keuangan.Tagihan{}).
		Where("bulan = ? AND tahun = ?", bulan, tahun).
		Select("COALESCE(SUM(nominal), 0)").
		Scan(&total)

	r.db.Model(&keuangan.Tagihan{}).
		Where("bulan = ? AND tahun = ? AND status = ?", bulan, tahun, "lunas").
		Count(&lunas)

	r.db.Model(&keuangan.Tagihan{}).
		Where("bulan = ? AND tahun = ? AND status = ?", bulan, tahun, "belum_bayar").
		Count(&belum)

	return
}
