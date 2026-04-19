package service

import (
	"context"
	"fmt"

	"github.com/sergehall/lavoval/apps/api/internal/domain"
	"github.com/sergehall/lavoval/apps/api/internal/repository"
)

type CatalogService struct {
	catalog *repository.CatalogRepository
}

func NewCatalogService(catalog *repository.CatalogRepository) *CatalogService {
	return &CatalogService{catalog: catalog}
}

func (s *CatalogService) ListCategories(ctx context.Context) ([]domain.Category, error) {
	cats, err := s.catalog.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	return cats, nil
}

func (s *CatalogService) ListSubcategories(ctx context.Context, categoryID string) ([]domain.Subcategory, error) {
	subs, err := s.catalog.ListSubcategories(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("list subcategories: %w", err)
	}
	return subs, nil
}

func (s *CatalogService) ListAllSubcategories(ctx context.Context) ([]domain.Subcategory, error) {
	subs, err := s.catalog.ListAllSubcategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all subcategories: %w", err)
	}
	return subs, nil
}

func (s *CatalogService) ListTags(ctx context.Context) ([]domain.Tag, error) {
	tags, err := s.catalog.ListTags(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}
	return tags, nil
}
