package service

import (
	"context"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
)

type UserService interface {
	List(ctx context.Context, page, limit int, role, search string) ([]dto.UserResponse, dto.PaginationMeta, error)
	Create(ctx context.Context, req dto.CreateUserRequest, actorID uint64, ipAddress string) (*dto.UserResponse, error)
	GetByID(ctx context.Context, id uint64) (*dto.UserResponse, error)
	Update(ctx context.Context, id uint64, req dto.UpdateUserRequest, actorID uint64, ipAddress string) (*dto.UserResponse, error)
	Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error
}

type userService struct {
	userRepo     repository.UserRepository
	auditService AuditService
	cfg          *config.Config
}

func NewUserService(userRepo repository.UserRepository, auditService AuditService, cfg *config.Config) UserService {
	return &userService{
		userRepo:     userRepo,
		auditService: auditService,
		cfg:          cfg,
	}
}

func toUserResponse(u *model.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func (s *userService) List(ctx context.Context, page, limit int, role, search string) ([]dto.UserResponse, dto.PaginationMeta, error) {
	users, totalRows, err := s.userRepo.FindAll(ctx, page, limit, role, search)
	if err != nil {
		return nil, dto.PaginationMeta{}, err
	}

	responses := make([]dto.UserResponse, len(users))
	for i := range users {
		responses[i] = *toUserResponse(&users[i])
	}

	meta := pagination.FormatPagination(page, limit, totalRows)
	return responses, meta, nil
}

func (s *userService) Create(ctx context.Context, req dto.CreateUserRequest, actorID uint64, ipAddress string) (*dto.UserResponse, error) {
	existing, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, response.NewServiceError(response.ErrDuplicate, "Email already registered")
	}

	cost := 12
	if s.cfg != nil && s.cfg.BcryptCost > 0 {
		cost = s.cfg.BcryptCost
	}

	hashedPassword, err := jwtpkg.HashPassword(req.Password, cost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	res := toUserResponse(user)

	if s.auditService != nil {
		var actorIDPtr *uint64
		if actorID > 0 {
			actorIDPtr = &actorID
		}
		_ = s.auditService.Log(ctx, AuditEntry{
			UserID:     actorIDPtr,
			Action:     "CREATE",
			EntityType: "user",
			EntityID:   &user.ID,
			NewValue:   res,
			IPAddress:  ipAddress,
		})
	}

	return res, nil
}

func (s *userService) GetByID(ctx context.Context, id uint64) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "User not found")
	}

	return toUserResponse(user), nil
}

func (s *userService) Update(ctx context.Context, id uint64, req dto.UpdateUserRequest, actorID uint64, ipAddress string) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "User not found")
	}

	oldValue := toUserResponse(user)

	if req.Email != nil && *req.Email != user.Email {
		existing, err := s.userRepo.FindByEmail(ctx, *req.Email)
		if err != nil {
			return nil, err
		}
		if existing != nil && existing.ID != user.ID {
			return nil, response.NewServiceError(response.ErrDuplicate, "Email already registered")
		}
		user.Email = *req.Email
	}

	if req.Name != nil {
		user.Name = *req.Name
	}

	if req.Role != nil {
		user.Role = *req.Role
	}

	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if req.Password != nil && *req.Password != "" {
		cost := 12
		if s.cfg != nil && s.cfg.BcryptCost > 0 {
			cost = s.cfg.BcryptCost
		}

		hashedPassword, err := jwtpkg.HashPassword(*req.Password, cost)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hashedPassword
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	res := toUserResponse(user)

	if s.auditService != nil {
		var actorIDPtr *uint64
		if actorID > 0 {
			actorIDPtr = &actorID
		}
		_ = s.auditService.Log(ctx, AuditEntry{
			UserID:     actorIDPtr,
			Action:     "UPDATE",
			EntityType: "user",
			EntityID:   &user.ID,
			OldValue:   oldValue,
			NewValue:   res,
			IPAddress:  ipAddress,
		})
	}

	return res, nil
}

func (s *userService) Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error {
	user, err := s.userRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if user == nil {
		return response.NewServiceError(response.ErrNotFound, "User not found")
	}

	oldValue := toUserResponse(user)

	if err := s.userRepo.Delete(ctx, id); err != nil {
		return err
	}

	if s.auditService != nil {
		var actorIDPtr *uint64
		if actorID > 0 {
			actorIDPtr = &actorID
		}
		_ = s.auditService.Log(ctx, AuditEntry{
			UserID:     actorIDPtr,
			Action:     "DELETE",
			EntityType: "user",
			EntityID:   &id,
			OldValue:   oldValue,
			IPAddress:  ipAddress,
		})
	}

	return nil
}
