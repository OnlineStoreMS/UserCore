package jwt

import (
	"crypto/rand"
	"encoding/hex"
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

// RefreshClaims is a long-lived token used only to obtain new access tokens.
type RefreshClaims struct {
	UserID   uint64 `json:"uid"`
	TenantID uint64 `json:"tid"`
	TokenUse string `json:"use"` // must be "refresh"
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

func (m *Manager) AccessTTL() time.Duration  { return m.accessTTL }
func (m *Manager) RefreshTTL() time.Duration { return m.refreshTTL }

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

func (m *Manager) IssueRefresh(userID, tenantID uint64) (tokenStr string, jti string, exp time.Time, err error) {
	now := time.Now()
	exp = now.Add(m.refreshTTL)
	jti, err = randomJTI()
	if err != nil {
		return "", "", time.Time{}, err
	}
	claims := RefreshClaims{
		UserID:   userID,
		TenantID: tenantID,
		TokenUse: "refresh",
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti,
			Subject:   "refresh",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err = token.SignedString(m.secret)
	return tokenStr, jti, exp, err
}

func randomJTI() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
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
	if claims.Subject == "refresh" {
		return nil, ErrInvalidToken
	}
	return claims, nil
}

func (m *Manager) ParseRefresh(tokenStr string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &RefreshClaims{}, func(t *jwt.Token) (interface{}, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, ErrInvalidToken
	}
	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}
	if claims.TokenUse != "refresh" || claims.UserID == 0 || claims.TenantID == 0 {
		return nil, ErrInvalidToken
	}
	return claims, nil
}
