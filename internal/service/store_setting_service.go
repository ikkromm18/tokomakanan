package service

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

const (
	defaultStoreName = "Toko Makanan"
)

type StoreSettingService interface {
	Get(ctx context.Context) (*dto.StoreSettingResponse, error)
	Update(ctx context.Context, actorID uint64, req dto.UpdateStoreSettingRequest, ipAddress string) (*dto.StoreSettingResponse, error)
}

type storeSettingService struct {
	repo         repository.StoreSettingRepository
	auditService AuditService
}

func NewStoreSettingService(repo repository.StoreSettingRepository, auditService AuditService) StoreSettingService {
	return &storeSettingService{
		repo:         repo,
		auditService: auditService,
	}
}

func toStoreSettingResponse(s *model.StoreSetting) *dto.StoreSettingResponse {
	return &dto.StoreSettingResponse{
		ID:            s.ID,
		Name:          s.Name,
		Address:       s.Address,
		Phone:         s.Phone,
		LogoURL:       s.LogoURL,
		ReceiptFooter: s.ReceiptFooter,
		UpdatedAt:     s.UpdatedAt,
	}
}

func (s *storeSettingService) Get(ctx context.Context) (*dto.StoreSettingResponse, error) {
	setting, err := s.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("storeSettingService.Get: %w", err)
	}
	if setting == nil {
		setting = &model.StoreSetting{
			Name: defaultStoreName,
		}
		if err := s.repo.Update(ctx, setting); err != nil {
			return nil, fmt.Errorf("storeSettingService.Get init: %w", err)
		}
	}
	return toStoreSettingResponse(setting), nil
}

func (s *storeSettingService) Update(ctx context.Context, actorID uint64, req dto.UpdateStoreSettingRequest, ipAddress string) (*dto.StoreSettingResponse, error) {
	setting, err := s.repo.Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("storeSettingService.Update get: %w", err)
	}

	var oldVal any
	if setting == nil {
		setting = &model.StoreSetting{
			Name: defaultStoreName,
		}
	} else {
		oldVal = toStoreSettingResponse(setting)
	}

	setting.Name = req.Name
	if req.Address != nil {
		setting.Address = req.Address
	}
	if req.Phone != nil {
		setting.Phone = req.Phone
	}
	if req.LogoURL != nil {
		setting.LogoURL = req.LogoURL
	}
	if req.ReceiptFooter != nil {
		setting.ReceiptFooter = req.ReceiptFooter
	}

	if err := s.repo.Update(ctx, setting); err != nil {
		return nil, fmt.Errorf("storeSettingService.Update save: %w", err)
	}

	res := toStoreSettingResponse(setting)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "UPDATE",
			EntityType: "store_setting",
			EntityID:   &setting.ID,
			OldValue:   oldVal,
			NewValue:   res,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}
