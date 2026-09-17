package service

import (
	"context"
	"errors"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"gorm.io/gorm"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest, ipAddress string) (*dto.LoginResponse, error)
	GetMe(ctx context.Context, userID uint64) (*dto.UserInfo, error)
	ChangePassword(ctx context.Context, userID uint64, req dto.ChangePasswordRequest, ipAddress string) error
}

type authService struct {
	userRepo     repository.UserRepository
	auditService AuditService
	cfg          *config.Config
}

func NewAuthService(userRepo repository.UserRepository, auditService AuditService, cfg *config.Config) AuthService {
	return &authService{
		userRepo:     userRepo,
		auditService: auditService,
		cfg:          cfg,
	}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest, ipAddress string) (*dto.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewServiceError(response.ErrUnauthorized, "Invalid credentials")
		}
		return nil, err
	}
	if user == nil {
		return nil, response.NewServiceError(response.ErrUnauthorized, "Invalid credentials")
	}

	if !user.IsActive {
		return nil, response.NewServiceError(response.ErrUnauthorized, "Account is deactivated")
	}

	if err := jwtpkg.ComparePassword(user.PasswordHash, req.Password); err != nil {
		return nil, response.NewServiceError(response.ErrUnauthorized, "Invalid credentials")
	}

	expiry := 24
	secret := ""
	if s.cfg != nil {
		if s.cfg.JWTExpiryHours > 0 {
			expiry = s.cfg.JWTExpiryHours
		}
		secret = s.cfg.JWTSecret
	}

	token, err := jwtpkg.GenerateToken(user.ID, user.Role, secret, expiry)
	if err != nil {
		return nil, err
	}

	if s.auditService != nil {
		_ = s.auditService.Log(ctx, AuditEntry{
			UserID:     &user.ID,
			Action:     "LOGIN",
			EntityType: "user",
			EntityID:   &user.ID,
			IPAddress:  ipAddress,
		})
	}

	return &dto.LoginResponse{
		Token: token,
		User: dto.UserInfo{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Role:  user.Role,
		},
	}, nil
}

func (s *authService) GetMe(ctx context.Context, userID uint64) (*dto.UserInfo, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, response.NewServiceError(response.ErrNotFound, "User not found")
		}
		return nil, err
	}
	if user == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "User not found")
	}

	return &dto.UserInfo{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Role:  user.Role,
	}, nil
}

func (s *authService) ChangePassword(ctx context.Context, userID uint64, req dto.ChangePasswordRequest, ipAddress string) error {
	if req.NewPassword != req.ConfirmPassword {
		return response.NewServiceError(response.ErrBusinessRule, "New password and confirm password do not match")
	}

	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.NewServiceError(response.ErrNotFound, "User not found")
		}
		return err
	}
	if user == nil {
		return response.NewServiceError(response.ErrNotFound, "User not found")
	}

	if err := jwtpkg.ComparePassword(user.PasswordHash, req.OldPassword); err != nil {
		return response.NewServiceError(response.ErrUnauthorized, "Invalid old password")
	}

	cost := 12
	if s.cfg != nil && s.cfg.BcryptCost > 0 {
		cost = s.cfg.BcryptCost
	}

	newHash, err := jwtpkg.HashPassword(req.NewPassword, cost)
	if err != nil {
		return err
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	if s.auditService != nil {
		_ = s.auditService.Log(ctx, AuditEntry{
			UserID:     &user.ID,
			Action:     "UPDATE",
			EntityType: "user",
			EntityID:   &user.ID,
			IPAddress:  ipAddress,
		})
	}

	return nil
}
