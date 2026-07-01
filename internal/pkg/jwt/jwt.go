package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	UserID      uint64   `json:"uid"`
	CompanyID   uint64   `json:"cid"`
	TenantID    uint64   `json:"tid"`
	Email       string   `json:"email"`
	DisplayName string   `json:"name"`
	Permissions []string `json:"perms"`
	IsPlatform  bool     `json:"platform"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewManager(secret string, accessMinutes, refreshHours int) *Manager {
	return &Manager{
		secret:     []byte(secret),
		accessTTL:  time.Duration(accessMinutes) * time.Minute,
		refreshTTL: time.Duration(refreshHours) * time.Hour,
	}
}

func (m *Manager) IssueAccess(claims Claims) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.accessTTL)
	claims.RegisteredClaims = jwt.RegisteredClaims{
		Subject:   claims.Email,
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(m.secret)
	return s, exp, err
}

func (m *Manager) IssueRefresh(userID, tenantID uint64) (string, time.Time, error) {
	now := time.Now()
	exp := now.Add(m.refreshTTL)
	claims := jwt.RegisteredClaims{
		Subject:   "refresh",
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(exp),
		ID:        formatRefreshID(userID, tenantID),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString(m.secret)
	return s, exp, err
}

func (m *Manager) ParseAccess(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func formatRefreshID(userID, tenantID uint64) string {
	return jwt.NewNumericDate(time.Unix(int64(userID), 0)).String() + ":" + jwt.NewNumericDate(time.Unix(int64(tenantID), 0)).String()
}
