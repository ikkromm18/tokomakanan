package service

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/config"
	"github.com/ikkromm18/tokomakanan/internal/dto"
	jwtpkg "github.com/ikkromm18/tokomakanan/internal/pkg/jwt"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
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
		return nil, fmt.Errorf("authService.Login: %w", err)
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
		return nil, fmt.Errorf("authService.Login token: %w", err)
	}

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     &user.ID,
			Action:     "LOGIN",
			EntityType: "user",
			EntityID:   &user.ID,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Msg("audit log: failed to record entry")
		}
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
		return nil, fmt.Errorf("authService.GetMe: %w", err)
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
		return fmt.Errorf("authService.ChangePassword: %w", err)
	}
	if user == nil {
		return response.NewServiceError(response.ErrNotFound, "User not found")
	}

	if err := jwtpkg.ComparePassword(user.PasswordHash, req.OldPassword); err != nil {
		return response.NewServiceError(response.ErrUnauthorized, "Invalid old password")
	}

	cost := getBcryptCost(s.cfg)

	newHash, err := jwtpkg.HashPassword(req.NewPassword, cost)
	if err != nil {
		return fmt.Errorf("authService.ChangePassword hash: %w", err)
	}

	user.PasswordHash = newHash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return fmt.Errorf("authService.ChangePassword update: %w", err)
	}

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     &user.ID,
			Action:     "CHANGE_PASSWORD",
			EntityType: "user",
			EntityID:   &user.ID,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Msg("audit log: failed to record entry")
		}
	}

	return nil
}
