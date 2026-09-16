package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type AuditEntry struct {
	UserID     *uint64
	Action     string
	EntityType string
	EntityID   *uint64
	OldValue   any
	NewValue   any
	IPAddress  string
}

type AuditService interface {
	Log(ctx context.Context, entry AuditEntry) error
	List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error)
}

type auditService struct {
	repo repository.AuditRepository
}

func NewAuditService(repo repository.AuditRepository) AuditService {
	return &auditService{repo: repo}
}

func (s *auditService) Log(ctx context.Context, entry AuditEntry) error {
	var oldValStr *string
	if entry.OldValue != nil {
		str, err := toJSONString(entry.OldValue)
		if err != nil {
			log.Error().Err(err).Interface("entry", entry).Msg("audit log: failed to marshal old_value")
			return nil
		}
		oldValStr = str
	}

	var newValStr *string
	if entry.NewValue != nil {
		str, err := toJSONString(entry.NewValue)
		if err != nil {
			log.Error().Err(err).Interface("entry", entry).Msg("audit log: failed to marshal new_value")
			return nil
		}
		newValStr = str
	}

	var ipAddr *string
	if entry.IPAddress != "" {
		ipAddr = &entry.IPAddress
	}

	auditLog := &model.AuditLog{
		UserID:     entry.UserID,
		Action:     entry.Action,
		EntityType: entry.EntityType,
		EntityID:   entry.EntityID,
		OldValue:   oldValStr,
		NewValue:   newValStr,
		IPAddress:  ipAddr,
		CreatedAt:  time.Now(),
	}

	if err := s.repo.Create(ctx, auditLog); err != nil {
		log.Error().Err(err).Interface("entry", entry).Msg("audit log: failed to create audit log in repository")
		return nil
	}

	return nil
}

func (s *auditService) List(ctx context.Context, page, limit int, entityType, action string) ([]dto.AuditLogResponse, dto.PaginationMeta, error) {
	logs, totalRows, err := s.repo.FindAll(ctx, page, limit, entityType, action)
	if err != nil {
		return nil, dto.PaginationMeta{}, err
	}

	responses := make([]dto.AuditLogResponse, len(logs))
	for i, l := range logs {
		var oldVal any
		if l.OldValue != nil {
			oldVal = json.RawMessage(*l.OldValue)
		}
		var newVal any
		if l.NewValue != nil {
			newVal = json.RawMessage(*l.NewValue)
		}

		responses[i] = dto.AuditLogResponse{
			ID:         l.ID,
			UserID:     l.UserID,
			Action:     l.Action,
			EntityType: l.EntityType,
			EntityID:   l.EntityID,
			OldValue:   oldVal,
			NewValue:   newVal,
			IPAddress:  l.IPAddress,
			CreatedAt:  l.CreatedAt,
		}
	}

	meta := pagination.FormatPagination(page, limit, totalRows)
	return responses, meta, nil
}

func toJSONString(v any) (*string, error) {
	if v == nil {
		return nil, nil
	}
	switch val := v.(type) {
	case string:
		if json.Valid([]byte(val)) {
			return &val, nil
		}
	case []byte:
		if json.Valid(val) {
			s := string(val)
			return &s, nil
		}
	}
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	s := string(b)
	return &s, nil
}
