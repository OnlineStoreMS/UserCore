package service

import (
	"errors"
	"strings"

	"usercore/internal/dto"
	"usercore/internal/model"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

type RoleService struct {
	repos *repo.Repos
}

func NewRoleService(repos *repo.Repos) *RoleService {
	return &RoleService{repos: repos}
}

func (s *RoleService) List(tenantID uint64) ([]dto.RoleDTO, error) {
	roles, err := s.repos.Role.ListByTenant(tenantID)
	if err != nil {
		return nil, err
	}
	out := make([]dto.RoleDTO, 0, len(roles))
	for i := range roles {
		d, err := s.toDTO(&roles[i])
		if err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, nil
}

func (s *RoleService) Create(tenantID uint64, req dto.CreateRoleRequest) (*dto.RoleDTO, error) {
	code := strings.TrimSpace(req.Code)
	role := &model.Role{
		TenantID:    tenantID,
		Code:        code,
		Name:        strings.TrimSpace(req.Name),
		Description: strings.TrimSpace(req.Description),
	}
	if err := s.repos.Role.Create(role); err != nil {
		return nil, err
	}
	if err := s.repos.Role.SetPermissions(role.ID, req.Permissions); err != nil {
		return nil, err
	}
	d, err := s.toDTO(role)
	return &d, err
}

func (s *RoleService) Update(tenantID, roleID uint64, req dto.UpdateRoleRequest) (*dto.RoleDTO, error) {
	role, err := s.repos.Role.GetByID(roleID)
	if err != nil {
		return nil, err
	}
	if role.TenantID != tenantID {
		return nil, ErrForbidden
	}
	if role.IsBuiltin == 1 {
		return nil, errors.New("内置角色不可修改")
	}
	if req.Name != "" {
		role.Name = strings.TrimSpace(req.Name)
	}
	if req.Description != "" {
		role.Description = strings.TrimSpace(req.Description)
	}
	if err := s.repos.Role.Update(role); err != nil {
		return nil, err
	}
	if req.Permissions != nil {
		if err := s.repos.Role.SetPermissions(role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}
	d, err := s.toDTO(role)
	return &d, err
}

func (s *RoleService) Delete(tenantID, roleID uint64) error {
	role, err := s.repos.Role.GetByID(roleID)
	if err != nil {
		return err
	}
	if role.TenantID != tenantID {
		return ErrForbidden
	}
	if role.IsBuiltin == 1 {
		return errors.New("内置角色不可删除")
	}
	return s.repos.Role.Delete(roleID)
}

func (s *RoleService) ListPermissions() ([]dto.PermissionDTO, error) {
	list, err := s.repos.Role.ListPermissions()
	if err != nil {
		return nil, err
	}
	out := make([]dto.PermissionDTO, 0, len(list))
	for _, p := range list {
		out = append(out, dto.PermissionDTO{Code: p.Code, Name: p.Name, AppCode: p.AppCode})
	}
	return out, nil
}

func (s *RoleService) toDTO(role *model.Role) (dto.RoleDTO, error) {
	perms, err := s.repos.Role.GetPermissionCodes(role.ID)
	if err != nil {
		return dto.RoleDTO{}, err
	}
	return dto.RoleDTO{
		ID: role.ID, Code: role.Code, Name: role.Name,
		Description: role.Description, IsBuiltin: role.IsBuiltin == 1,
		Permissions: perms,
	}, nil
}

type TenantService struct {
	repos *repo.Repos
}

func NewTenantService(repos *repo.Repos) *TenantService {
	return &TenantService{repos: repos}
}

func (s *TenantService) List(q dto.PageQuery) ([]dto.TenantDTO, int64, error) {
	list, total, err := s.repos.Tenant.ListAll(strings.TrimSpace(q.Keyword), q.Page, q.PageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.TenantDTO, 0, len(list))
	for i := range list {
		out = append(out, s.toDTO(&list[i]))
	}
	return out, total, nil
}

func (s *TenantService) Create(companyID uint64, req dto.CreateTenantRequest) (*dto.TenantDTO, error) {
	if req.CompanyID > 0 {
		companyID = req.CompanyID
	}
	t := &model.Tenant{
		CompanyID: companyID,
		Name:      strings.TrimSpace(req.Name),
		Code:      strings.TrimSpace(req.Code),
		Status:    1,
		Remark:    strings.TrimSpace(req.Remark),
	}
	if err := s.repos.Tenant.Create(t); err != nil {
		return nil, err
	}
	if err := SeedBuiltinRoles(s.repos, t.ID); err != nil {
		return nil, err
	}
	if err := EnsurePlatformUsersInTenant(s.repos, t.ID); err != nil {
		return nil, err
	}
	d := s.toDTO(t)
	return &d, nil
}

func (s *TenantService) Update(id uint64, req dto.UpdateTenantRequest) (*dto.TenantDTO, error) {
	t, err := s.repos.Tenant.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if req.Name != "" {
		t.Name = strings.TrimSpace(req.Name)
	}
	if req.Code != "" {
		t.Code = strings.TrimSpace(req.Code)
	}
	if req.Status != nil {
		t.Status = *req.Status
	}
	if req.Remark != "" {
		t.Remark = strings.TrimSpace(req.Remark)
	}
	if err := s.repos.Tenant.Save(t); err != nil {
		return nil, err
	}
	d := s.toDTO(t)
	return &d, nil
}

func (s *TenantService) Get(id uint64) (*dto.TenantDTO, error) {
	t, err := s.repos.Tenant.GetByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	d := s.toDTO(t)
	return &d, nil
}

func (s *TenantService) toDTO(t *model.Tenant) dto.TenantDTO {
	d := dto.TenantDTO{
		ID: t.ID, CompanyID: t.CompanyID, Name: t.Name,
		Code: t.Code, Status: t.Status, Remark: t.Remark,
	}
	if c, err := s.repos.Company.GetByID(t.CompanyID); err == nil {
		d.CompanyName = c.Name
	}
	return d
}

func toTenantDTO(t *model.Tenant) dto.TenantDTO {
	return dto.TenantDTO{
		ID: t.ID, CompanyID: t.CompanyID, Name: t.Name,
		Code: t.Code, Status: t.Status, Remark: t.Remark,
	}
}

