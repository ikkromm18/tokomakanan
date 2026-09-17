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

type ProductService interface {
	List(ctx context.Context, page, limit int, category string, isActive *bool, search string, userRole string) ([]dto.ProductResponse, dto.PaginationMeta, error)
	Create(ctx context.Context, req dto.CreateProductRequest, actorID uint64, userRole string, ipAddress string) (*dto.ProductResponse, error)
	GetByID(ctx context.Context, id uint64, userRole string) (*dto.ProductResponse, error)
	Update(ctx context.Context, id uint64, req dto.UpdateProductRequest, actorID uint64, userRole string, ipAddress string) (*dto.ProductResponse, error)
	Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error
}

type productService struct {
	productRepo  repository.ProductRepository
	auditService AuditService
}

func NewProductService(productRepo repository.ProductRepository, auditService AuditService) ProductService {
	return &productService{
		productRepo:  productRepo,
		auditService: auditService,
	}
}

func toProductResponse(p *model.Product, userRole string) *dto.ProductResponse {
	res := &dto.ProductResponse{
		ID:        p.ID,
		Name:      p.Name,
		Category:  p.Category,
		SellPrice: p.SellPrice,
		IsActive:  p.IsActive,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	// Admins cannot see HPP
	if userRole != model.RoleAdmin {
		hppVal := p.HPP
		res.HPP = &hppVal
	}

	return res
}

func (s *productService) List(ctx context.Context, page, limit int, category string, isActive *bool, search string, userRole string) ([]dto.ProductResponse, dto.PaginationMeta, error) {
	products, totalRows, err := s.productRepo.FindAll(ctx, page, limit, category, isActive, search)
	if err != nil {
		log.Error().Err(err).Str("component", "productService.List").Msg("failed to list products")
		return nil, dto.PaginationMeta{}, fmt.Errorf("productService.List: %w", err)
	}

	responses := make([]dto.ProductResponse, len(products))
	for i := range products {
		responses[i] = *toProductResponse(&products[i], userRole)
	}

	meta := pagination.FormatPagination(page, limit, totalRows)
	return responses, meta, nil
}

func (s *productService) Create(ctx context.Context, req dto.CreateProductRequest, actorID uint64, userRole string, ipAddress string) (*dto.ProductResponse, error) {
	// Business Rule: sell_price >= hpp
	if req.SellPrice < req.HPP {
		return nil, response.NewServiceError(response.ErrBusinessRule, "sell price cannot be less than HPP")
	}

	// Unique name within category
	existing, err := s.productRepo.FindByNameAndCategory(ctx, req.Name, req.Category)
	if err != nil {
		log.Error().Err(err).Str("component", "productService.Create").Msg("failed to check duplicate product")
		return nil, fmt.Errorf("productService.Create check duplicate: %w", err)
	}
	if existing != nil {
		return nil, response.NewServiceError(response.ErrDuplicate, "product with this name and category already exists")
	}

	product := &model.Product{
		Name:      req.Name,
		Category:  req.Category,
		HPP:       req.HPP,
		SellPrice: req.SellPrice,
		IsActive:  true,
	}

	if err := s.productRepo.Create(ctx, product); err != nil {
		log.Error().Err(err).Str("component", "productService.Create").Msg("failed to create product")
		return nil, fmt.Errorf("productService.Create: %w", err)
	}

	res := toProductResponse(product, userRole)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "CREATE",
			EntityType: "product",
			EntityID:   &product.ID,
			NewValue:   product,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "productService.Create").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *productService) GetByID(ctx context.Context, id uint64, userRole string) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "productService.GetByID").Msg("failed to find product")
		return nil, fmt.Errorf("productService.GetByID: %w", err)
	}
	if product == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Product not found")
	}

	return toProductResponse(product, userRole), nil
}

func (s *productService) Update(ctx context.Context, id uint64, req dto.UpdateProductRequest, actorID uint64, userRole string, ipAddress string) (*dto.ProductResponse, error) {
	// Business Rule: sell_price >= hpp
	if req.SellPrice < req.HPP {
		return nil, response.NewServiceError(response.ErrBusinessRule, "sell price cannot be less than HPP")
	}

	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "productService.Update").Msg("failed to find product")
		return nil, fmt.Errorf("productService.Update: %w", err)
	}
	if product == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Product not found")
	}

	// Check name uniqueness if name or category changed
	if product.Name != req.Name || product.Category != req.Category {
		existing, err := s.productRepo.FindByNameAndCategory(ctx, req.Name, req.Category)
		if err != nil {
			log.Error().Err(err).Str("component", "productService.Update").Msg("failed to check duplicate product")
			return nil, fmt.Errorf("productService.Update check duplicate: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, response.NewServiceError(response.ErrDuplicate, "product with this name and category already exists")
		}
	}

	oldVal := *product

	product.Name = req.Name
	product.Category = req.Category
	product.HPP = req.HPP
	product.SellPrice = req.SellPrice
	if req.IsActive != nil {
		product.IsActive = *req.IsActive
	}

	if err := s.productRepo.Update(ctx, product); err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "productService.Update").Msg("failed to update product")
		return nil, fmt.Errorf("productService.Update: %w", err)
	}

	res := toProductResponse(product, userRole)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "UPDATE",
			EntityType: "product",
			EntityID:   &product.ID,
			OldValue:   oldVal,
			NewValue:   product,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "productService.Update").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *productService) Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error {
	product, err := s.productRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "productService.Delete").Msg("failed to find product")
		return fmt.Errorf("productService.Delete: %w", err)
	}
	if product == nil {
		return response.NewServiceError(response.ErrNotFound, "Product not found")
	}

	// Constraint: cannot delete if part of active package
	inPackages, err := s.productRepo.ExistsInActivePackages(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "productService.Delete").Msg("failed to check package constraints")
		return fmt.Errorf("productService.Delete check packages: %w", err)
	}
	if inPackages {
		return response.NewServiceError(response.ErrBusinessRule, "cannot delete product: product is used in active package")
	}

	oldVal := *product

	if err := s.productRepo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "productService.Delete").Msg("failed to delete product")
		return fmt.Errorf("productService.Delete: %w", err)
	}

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "DELETE",
			EntityType: "product",
			EntityID:   &id,
			OldValue:   oldVal,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "productService.Delete").Msg("audit log: failed to record entry")
		}
	}

	return nil
}
