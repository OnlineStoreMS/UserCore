package service

import (
	"errors"
	"strings"

	"usercore/internal/dto"
	"usercore/internal/model"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

type CompanyService struct {
	repos *repo.Repos
}

func NewCompanyService(repos *repo.Repos) *CompanyService {
	return &CompanyService{repos: repos}
}

func (s *CompanyService) List(q dto.PageQuery) ([]dto.CompanyDTO, int64, error) {
	list, total, err := s.repos.Company.List(strings.TrimSpace(q.Keyword), q.Page, q.PageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.CompanyDTO, 0, len(list))
	for i := range list {
		out = append(out, s.toDTO(&list[i]))
	}
	return out, total, nil
}

func (s *CompanyService) Get(id uint64) (*dto.CompanyDTO, error) {
	c, err := s.repos.Company.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d := s.toDTO(c)
	return &d, nil
}

func (s *CompanyService) Create(req dto.CreateCompanyRequest) (*dto.CompanyDTO, error) {
	c := &model.Company{
		Name:   strings.TrimSpace(req.Name),
		Code:   strings.TrimSpace(req.Code),
		Status: 1,
		Remark: strings.TrimSpace(req.Remark),
	}
	if err := s.repos.Company.Create(c); err != nil {
		return nil, err
	}
	d := s.toDTO(c)
	return &d, nil
}

func (s *CompanyService) Update(id uint64, req dto.UpdateCompanyRequest) (*dto.CompanyDTO, error) {
	c, err := s.repos.Company.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if req.Name != "" {
		c.Name = strings.TrimSpace(req.Name)
	}
	if req.Code != "" {
		c.Code = strings.TrimSpace(req.Code)
	}
	if req.Status != nil {
		c.Status = *req.Status
	}
	if req.Remark != "" {
		c.Remark = strings.TrimSpace(req.Remark)
	}
	if err := s.repos.Company.Save(c); err != nil {
		return nil, err
	}
	d := s.toDTO(c)
	return &d, nil
}

func (s *CompanyService) toDTO(c *model.Company) dto.CompanyDTO {
	count, _ := s.repos.Company.CountTenants(c.ID)
	return dto.CompanyDTO{
		ID: c.ID, Name: c.Name, Code: c.Code, Status: c.Status, Remark: c.Remark,
		TenantCount: count,
	}
}
