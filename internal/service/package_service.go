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

type PackageService interface {
	List(ctx context.Context, page, limit int, isActive *bool, search string, userRole string) ([]dto.PackageResponse, dto.PaginationMeta, error)
	Create(ctx context.Context, req dto.CreatePackageRequest, actorID uint64, userRole string, ipAddress string) (*dto.PackageResponse, error)
	GetByID(ctx context.Context, id uint64, userRole string) (*dto.PackageResponse, error)
	Update(ctx context.Context, id uint64, req dto.UpdatePackageRequest, actorID uint64, userRole string, ipAddress string) (*dto.PackageResponse, error)
	Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error
}

type packageService struct {
	packageRepo  repository.PackageRepository
	productRepo  repository.ProductRepository
	auditService AuditService
}

func NewPackageService(packageRepo repository.PackageRepository, productRepo repository.ProductRepository, auditService AuditService) PackageService {
	return &packageService{
		packageRepo:  packageRepo,
		productRepo:  productRepo,
		auditService: auditService,
	}
}

func toPackageResponse(pkg *model.ProductPackage, userRole string) *dto.PackageResponse {
	res := &dto.PackageResponse{
		ID:        pkg.ID,
		Name:      pkg.Name,
		SellPrice: pkg.SellPrice,
		IsActive:  pkg.IsActive,
		CreatedAt: pkg.CreatedAt,
		UpdatedAt: pkg.UpdatedAt,
	}

	if userRole != model.RoleAdmin {
		totalHPPVal := pkg.TotalHPP
		res.TotalHPP = &totalHPPVal
	}

	if len(pkg.Items) > 0 {
		res.Items = make([]dto.PackageItemResponse, len(pkg.Items))
		for i, it := range pkg.Items {
			itemResp := dto.PackageItemResponse{
				ID:          it.ID,
				ProductID:   it.ProductID,
				ProductName: it.Product.Name,
				Quantity:    it.Quantity,
				UnitSell:    it.Product.SellPrice,
			}
			if userRole != model.RoleAdmin {
				hppVal := it.Product.HPP
				itemResp.UnitHPP = &hppVal
			}
			res.Items[i] = itemResp
		}
	}

	return res
}

func (s *packageService) calculatePackageItems(ctx context.Context, itemReqs []dto.PackageItemRequest) ([]model.PackageItem, float64, error) {
	var totalHPP float64
	items := make([]model.PackageItem, len(itemReqs))
	seen := make(map[uint64]bool)

	for i, it := range itemReqs {
		if seen[it.ProductID] {
			return nil, 0, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("duplicate product ID %d in package items", it.ProductID))
		}
		seen[it.ProductID] = true

		prod, err := s.productRepo.FindByID(ctx, it.ProductID)
		if err != nil {
			return nil, 0, fmt.Errorf("calculatePackageItems find product %d: %w", it.ProductID, err)
		}
		if prod == nil {
			return nil, 0, response.NewServiceError(response.ErrNotFound, fmt.Sprintf("product with ID %d not found", it.ProductID))
		}
		if !prod.IsActive {
			return nil, 0, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("product %s is not active", prod.Name))
		}

		totalHPP += prod.HPP * float64(it.Quantity)
		items[i] = model.PackageItem{
			ProductID: it.ProductID,
			Quantity:  it.Quantity,
			Product:   *prod,
		}
	}

	return items, totalHPP, nil
}

func (s *packageService) List(ctx context.Context, page, limit int, isActive *bool, search string, userRole string) ([]dto.PackageResponse, dto.PaginationMeta, error) {
	packages, totalRows, err := s.packageRepo.FindAll(ctx, page, limit, isActive, search)
	if err != nil {
		log.Error().Err(err).Str("component", "packageService.List").Msg("failed to list packages")
		return nil, dto.PaginationMeta{}, fmt.Errorf("packageService.List: %w", err)
	}

	responses := make([]dto.PackageResponse, len(packages))
	for i := range packages {
		responses[i] = *toPackageResponse(&packages[i], userRole)
	}

	meta := pagination.FormatPagination(page, limit, totalRows)
	return responses, meta, nil
}

func (s *packageService) Create(ctx context.Context, req dto.CreatePackageRequest, actorID uint64, userRole string, ipAddress string) (*dto.PackageResponse, error) {
	if len(req.Items) == 0 {
		return nil, response.NewServiceError(response.ErrBusinessRule, "package must have at least one product item")
	}

	existing, err := s.packageRepo.FindByName(ctx, req.Name)
	if err != nil {
		log.Error().Err(err).Str("component", "packageService.Create").Msg("failed to check package name uniqueness")
		return nil, fmt.Errorf("packageService.Create check name: %w", err)
	}
	if existing != nil {
		return nil, response.NewServiceError(response.ErrDuplicate, "package with this name already exists")
	}

	items, totalHPP, err := s.calculatePackageItems(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	// Business rule: sell_price >= total_hpp
	if req.SellPrice < totalHPP {
		return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("sell price cannot be less than total HPP (calculated total HPP: %.2f)", totalHPP))
	}

	pkg := &model.ProductPackage{
		Name:      req.Name,
		TotalHPP:  totalHPP,
		SellPrice: req.SellPrice,
		IsActive:  true,
		Items:     items,
	}

	if err := s.packageRepo.Create(ctx, pkg); err != nil {
		log.Error().Err(err).Str("component", "packageService.Create").Msg("failed to create package")
		return nil, fmt.Errorf("packageService.Create: %w", err)
	}

	res := toPackageResponse(pkg, userRole)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "CREATE",
			EntityType: "package",
			EntityID:   &pkg.ID,
			NewValue:   pkg,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "packageService.Create").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *packageService) GetByID(ctx context.Context, id uint64, userRole string) (*dto.PackageResponse, error) {
	pkg, err := s.packageRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "packageService.GetByID").Msg("failed to get package")
		return nil, fmt.Errorf("packageService.GetByID: %w", err)
	}
	if pkg == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Package not found")
	}

	return toPackageResponse(pkg, userRole), nil
}

func (s *packageService) Update(ctx context.Context, id uint64, req dto.UpdatePackageRequest, actorID uint64, userRole string, ipAddress string) (*dto.PackageResponse, error) {
	if len(req.Items) == 0 {
		return nil, response.NewServiceError(response.ErrBusinessRule, "package must have at least one product item")
	}

	pkg, err := s.packageRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "packageService.Update").Msg("failed to find package")
		return nil, fmt.Errorf("packageService.Update: %w", err)
	}
	if pkg == nil {
		return nil, response.NewServiceError(response.ErrNotFound, "Package not found")
	}

	if pkg.Name != req.Name {
		existing, err := s.packageRepo.FindByName(ctx, req.Name)
		if err != nil {
			log.Error().Err(err).Str("component", "packageService.Update").Msg("failed to check package name")
			return nil, fmt.Errorf("packageService.Update check duplicate: %w", err)
		}
		if existing != nil && existing.ID != id {
			return nil, response.NewServiceError(response.ErrDuplicate, "package with this name already exists")
		}
	}

	items, totalHPP, err := s.calculatePackageItems(ctx, req.Items)
	if err != nil {
		return nil, err
	}

	// Business rule: sell_price >= total_hpp
	if req.SellPrice < totalHPP {
		return nil, response.NewServiceError(response.ErrBusinessRule, fmt.Sprintf("sell price cannot be less than total HPP (calculated total HPP: %.2f)", totalHPP))
	}

	oldVal := *pkg

	pkg.Name = req.Name
	pkg.TotalHPP = totalHPP
	pkg.SellPrice = req.SellPrice
	if req.IsActive != nil {
		pkg.IsActive = *req.IsActive
	}
	pkg.Items = items

	if err := s.packageRepo.Update(ctx, pkg, items); err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "packageService.Update").Msg("failed to update package")
		return nil, fmt.Errorf("packageService.Update: %w", err)
	}

	res := toPackageResponse(pkg, userRole)

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "UPDATE",
			EntityType: "package",
			EntityID:   &pkg.ID,
			OldValue:   oldVal,
			NewValue:   pkg,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "packageService.Update").Msg("audit log: failed to record entry")
		}
	}

	return res, nil
}

func (s *packageService) Delete(ctx context.Context, id uint64, actorID uint64, ipAddress string) error {
	pkg, err := s.packageRepo.FindByID(ctx, id)
	if err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "packageService.Delete").Msg("failed to find package")
		return fmt.Errorf("packageService.Delete: %w", err)
	}
	if pkg == nil {
		return response.NewServiceError(response.ErrNotFound, "Package not found")
	}

	oldVal := *pkg

	if err := s.packageRepo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Uint64("id", id).Str("component", "packageService.Delete").Msg("failed to delete package")
		return fmt.Errorf("packageService.Delete: %w", err)
	}

	if s.auditService != nil {
		if err := s.auditService.Log(ctx, AuditEntry{
			UserID:     toActorPtr(actorID),
			Action:     "DELETE",
			EntityType: "package",
			EntityID:   &id,
			OldValue:   oldVal,
			IPAddress:  ipAddress,
		}); err != nil {
			log.Warn().Err(err).Str("component", "packageService.Delete").Msg("audit log: failed to record entry")
		}
	}

	return nil
}
