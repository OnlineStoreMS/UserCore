package service

import (
	"errors"
	"strings"

	"usercore/internal/dto"
	"usercore/internal/model"
	"usercore/internal/pkg/password"
	"usercore/internal/repo"

	"gorm.io/gorm"
)

type UserService struct {
	repos *repo.Repos
}

func NewUserService(repos *repo.Repos) *UserService {
	return &UserService{repos: repos}
}

func (s *UserService) List(tenantID uint64, q dto.PageQuery) ([]dto.UserDTO, int64, error) {
	users, total, err := s.repos.User.ListByTenant(tenantID, strings.TrimSpace(q.Keyword), q.Page, q.PageSize)
	if err != nil {
		return nil, 0, err
	}
	out := make([]dto.UserDTO, 0, len(users))
	for i := range users {
		d, err := s.toDTO(tenantID, &users[i])
		if err != nil {
			return nil, 0, err
		}
		out = append(out, d)
	}
	return out, total, nil
}

func (s *UserService) Create(tenantID uint64, req dto.CreateUserRequest) (*dto.UserDTO, error) {
	email := strings.TrimSpace(strings.ToLower(req.Email))
	if _, err := s.repos.User.FindByEmail(email); err == nil {
		return nil, errors.New("邮箱已存在")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := password.Hash(req.Password)
	if err != nil {
		return nil, err
	}
	u := &model.User{
		Email:       email,
		Password:    hash,
		DisplayName: strings.TrimSpace(req.DisplayName),
		Phone:       strings.TrimSpace(req.Phone),
		Status:      1,
	}
	if err := s.repos.User.Create(u); err != nil {
		return nil, err
	}
	if err := s.repos.User.AddMember(tenantID, u.ID); err != nil {
		return nil, err
	}
	if len(req.RoleIDs) > 0 {
		if err := s.repos.User.SetUserRoles(tenantID, u.ID, req.RoleIDs); err != nil {
			return nil, err
		}
	}
	d, err := s.toDTO(tenantID, u)
	return &d, err
}

func (s *UserService) Update(tenantID, userID uint64, req dto.UpdateUserRequest) (*dto.UserDTO, error) {
	ok, err := s.repos.User.IsMember(userID, tenantID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrForbidden
	}
	u, err := s.repos.User.GetByID(userID)
	if err != nil {
		return nil, err
	}
	if req.DisplayName != "" {
		u.DisplayName = strings.TrimSpace(req.DisplayName)
	}
	if req.Phone != "" {
		u.Phone = strings.TrimSpace(req.Phone)
	}
	if req.Status != nil {
		u.Status = *req.Status
	}
	if req.Password != "" {
		hash, err := password.Hash(req.Password)
		if err != nil {
			return nil, err
		}
		u.Password = hash
	}
	if err := s.repos.User.Update(u); err != nil {
		return nil, err
	}
	if req.RoleIDs != nil {
		if err := s.repos.User.SetUserRoles(tenantID, userID, req.RoleIDs); err != nil {
			return nil, err
		}
	}
	d, err := s.toDTO(tenantID, u)
	return &d, err
}

func (s *UserService) RemoveFromTenant(tenantID, userID, operatorID uint64) error {
	if userID == operatorID {
		return errors.New("不能移除当前登录用户")
	}
	ok, err := s.repos.User.IsMember(userID, tenantID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}
	return s.repos.User.RemoveMember(tenantID, userID)
}

func (s *UserService) toDTO(tenantID uint64, u *model.User) (dto.UserDTO, error) {
	roleIDs, err := s.repos.User.GetRoleIDs(tenantID, u.ID)
	if err != nil {
		return dto.UserDTO{}, err
	}
	roles := make([]dto.RoleDTO, 0, len(roleIDs))
	for _, rid := range roleIDs {
		role, err := s.repos.Role.GetByID(rid)
		if err != nil {
			continue
		}
		perms, _ := s.repos.Role.GetPermissionCodes(role.ID)
		roles = append(roles, dto.RoleDTO{
			ID: role.ID, Code: role.Code, Name: role.Name,
			Description: role.Description, IsBuiltin: role.IsBuiltin == 1,
			Permissions: perms,
		})
	}
	return dto.UserDTO{
		ID: u.ID, Email: u.Email, DisplayName: u.DisplayName,
		Phone: u.Phone, Status: u.Status, Roles: roles,
		CreatedAt: u.CreatedAt.Format("2006-01-02 15:04:05"),
	}, nil
}
