package usecase

import (
	"database/sql"

	"github.com/gin-gonic/gin"

	"github.com/atam/atamlink/internal/mod_business/dto"
	"github.com/atam/atamlink/internal/mod_business/repository"
	masterRepo "github.com/atam/atamlink/internal/mod_master/repository"
	userRepo "github.com/atam/atamlink/internal/mod_user/repository"
	"github.com/atam/atamlink/internal/service"
)

// BusinessUseCase interface untuk business use case
type BusinessUseCase interface {
	// Core business operations
	Create(ctx *gin.Context, profileID int64, req *dto.CreateBusinessRequest) (*dto.BusinessResponse, error)
	GetByID(id int64, profileID int64) (*dto.BusinessResponse, error)
	GetBySlug(slug string) (*dto.BusinessResponse, error)
	List(profileID int64, filter *dto.BusinessFilter, page, perPage int, orderBy string) ([]*dto.BusinessListResponse, int64, error)
	Update(ctx *gin.Context, id int64, profileID int64, req *dto.UpdateBusinessRequest) (*dto.BusinessResponse, error)
	Delete(ctx *gin.Context, id int64, profileID int64) error

	// User management
	AddUser(ctx *gin.Context, businessID int64, profileID int64, req *dto.AddUserRequest) error
	UpdateUserRole(ctx *gin.Context, businessID int64, profileID int64, targetProfileID int64, role string) error
	RemoveUser(ctx *gin.Context, businessID int64, profileID int64, targetProfileID int64) error

	// Invite management
	CreateInvite(ctx *gin.Context, businessID int64, profileID int64, req *dto.CreateInviteRequest) (*dto.InviteResponse, error)
	AcceptInvite(ctx *gin.Context, req *dto.AcceptInviteRequest) error

	// Subscription management
	ActivateSubscription(ctx *gin.Context, profileID int64, req *dto.ActivateSubscriptionRequest) (*dto.SubscriptionResponse, error)
}

type businessUseCase struct {
	db            *sql.DB
	businessRepo  repository.BusinessRepository
	userRepo      userRepo.UserRepository
	masterRepo    masterRepo.MasterRepository
	slugService   service.SlugService
	uploadService service.UploadService
}

// NewBusinessUseCase membuat instance business use case baru
func NewBusinessUseCase(
	db *sql.DB,
	businessRepo repository.BusinessRepository,
	userRepo userRepo.UserRepository,
	masterRepo masterRepo.MasterRepository,
	slugService service.SlugService,
	uploadService service.UploadService,
) BusinessUseCase {
	return &businessUseCase{
		db:            db,
		businessRepo:  businessRepo,
		userRepo:      userRepo,
		masterRepo:    masterRepo,
		slugService:   slugService,
		uploadService: uploadService,
	}
}