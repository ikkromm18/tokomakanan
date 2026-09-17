package service

import (
	"context"
	"fmt"

	"github.com/ikkromm18/tokomakanan/internal/dto"
	"github.com/ikkromm18/tokomakanan/internal/model"
	"github.com/ikkromm18/tokomakanan/internal/pkg/pagination"
	"github.com/ikkromm18/tokomakanan/internal/pkg/response"
	"github.com/ikkromm18/tokomakanan/internal/repository"
	"github.com/rs/zerolog/log"
)

type CustomerService interface {
	List(ctx context.Context, page, limit int, search string) ([]dto.CustomerResponse, dto.PaginationMeta, error)
	Create(ctx context.Context, req dto.CreateCustomerRequest, actorID uint64, ipAddress string) (*dto.CustomerResponse, error)
	GetByID(ctx context.Context, id uint64) (*dto.CustomerResponse, error)
	Update(ctx context.Context, id uint64, req dto.UpdateCustomerRequest, actorID uint64, ipAddress string) (*dto.CustomerResponse, error)
}

type customerService struct {
	customerRepo repository.CustomerRepository
	auditService AuditService
}

func NewCustomerService(customerRepo repository.CustomerRepository, auditService AuditService) CustomerService {
	return &customerService{
		customerRepo: customerRepo,
		auditService: auditService,
	}
}

func toCustomerResponse(c *model.Customer) *dto.CustomerResponse {
	return &dto.CustomerResponse{
		ID:        c.ID,
		Name:      c.Name,
		Phone:     c.Phone,
		Address:   c.Address,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}
}

func (s *customerService) List(ctx context.Context, page, limit int, search string) ([]dto.CustomerResponse, dto.PaginationMeta, error) {
	customers, totalRows, err := s.customerRepo.FindAll(ctx, page, limit, search)
	if err != nil {
		log.Error().Err(err).Str("component", "customerService.List").Msg("failed to list customers")
		return nil, dto.PaginationMeta{}, fmt.Errorf("customerService.List: %w", err)
	}

	responses := make([]dto.CustomerResponse, len(customers))
	for i := range customers {
		responses[i] = *toCustomerResponse(&customers[i])
	}

	meta := pagination.FormatPagination(page, limit, totalRows)
	return responses, meta, nil
}

func (s *customerService) Create(ctx context.Context, req dto.CreateCustomerRequest, actorID uint64, ipAddress string) (*dto.CustomerResponse, error) {
	existing, err := s.customerRepo.FindByPhone(ctx, req.Phone)
	if err != nil {
		log.Error().Err(err).Str("component", "customerService.Create").Msg("failed to check customer phone")
		return nil, fmt.Errorf("customerService.Create check phone: %w", err)
	}
	if existing != nil {
		return nil, response.NewServiceError(response.ErrDuplicate, "phone number already registered")
	}

	customer := &model.Customer{
		Name:    req.Name,
		Phone:   req.Phone,
		Address: req.Address,
	}

	if err := s.customerRepo.Create(ctx, customer); err != nil {
		log.Error().Err(err).Str("component", "customerService.Create").Msg("failed to create customer")
		return nil, fmt.Errorf("customerService.Create: %w", err)
	}

	res := toCustomerResponse(customer)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "CREATE",
			EntityType: "customer",
			EntityID:   &customer.ID,
			NewValue:   customer,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "customerService.Create").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *customerService) GetByID(ctx context.Context, id uint64) (*dto.CustomerResponse, error) {
	customer, err := s.customerRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "customerService.GetByID").Msg("failed to find customer")
		return nil, fmt.Errorf("customerService.GetByID: %w", err)
	}
	if customer == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Customer not found")
	}

	return toCustomerResponse(customer), nil
}

func (s *customerService) Update(ctx context.Context, id uint64, req dto.UpdateCustomerRequest, actorID uint64, ipAddress string) (*dto.CustomerResponse, error) {
	customer, err := s.customerRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "customerService.Update").Msg("failed to find customer")
		return nil, fmt.Errorf("customerService.Update: %w", err)
	}
	if customer == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Customer not found")
	}

	if customer.Phone != req.Phone {
		existing, err := s.customerRepo.FindByPhone(ctx, req.Phone)
		if err != nil {
			log.Error().Err(err).Str("component", "customerService.Update").Msg("failed to check customer phone")
			return nil, fmt.Errorf("customerService.Update check phone: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, response.NewServiceError(response.ErrDuplicate, "phone number already registered")
		}
	}

	oldVal := *customer

	customer.Name = req.Name
	customer.Phone = req.Phone
	customer.Address = req.Address

	if err := s.customerRepo.Update(ctx, customer); err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "customerService.Update").Msg("failed to update customer")
		return nil, fmt.Errorf("customerService.Update: %w", err)
	}

	res := toCustomerResponse(customer)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "UPDATE",
			EntityType: "customer",
			EntityID:   &customer.ID,
			OldValue:   oldVal,
			NewValue:   customer,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "customerService.Update").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}
