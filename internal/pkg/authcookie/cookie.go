package authcookie

import (
	"net/http"
	"strings"
	"time"
)

const (
	AccessCookie  = "uc_access"
	RefreshCookie = "uc_refresh"
)

type Config struct {
	Domain   string
	Secure   bool
	SameSite http.SameSite
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func NewConfig(domain string, secure bool, sameSite string, accessTTL, refreshTTL time.Duration) Config {
	ss := http.SameSiteLaxMode
	switch strings.ToLower(strings.TrimSpace(sameSite)) {
	case "strict":
		ss = http.SameSiteStrictMode
	case "none":
		ss = http.SameSiteNoneMode
	case "lax", "":
		ss = http.SameSiteLaxMode
	}
	return Config{
		Domain:     strings.TrimSpace(domain),
		Secure:     secure,
		SameSite:   ss,
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	}
}

func (cfg Config) SetTokens(w http.ResponseWriter, accessToken, refreshToken string) {
	cfg.set(w, AccessCookie, accessToken, cfg.AccessTTL)
	if refreshToken != "" {
		cfg.set(w, RefreshCookie, refreshToken, cfg.RefreshTTL)
	}
}

func (cfg Config) Clear(w http.ResponseWriter) {
	cfg.set(w, AccessCookie, "", -1)
	cfg.set(w, RefreshCookie, "", -1)
}

func (cfg Config) set(w http.ResponseWriter, name, value string, maxAge time.Duration) {
	c := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   cfg.Secure,
		SameSite: cfg.SameSite,
	}
	if cfg.Domain != "" {
		c.Domain = cfg.Domain
	}
	if maxAge < 0 {
		c.MaxAge = -1
		c.Expires = time.Unix(0, 0)
	} else {
		c.MaxAge = int(maxAge.Seconds())
	}
	http.SetCookie(w, c)
}

func AccessToken(r *http.Request) string {
	return cookieValue(r, AccessCookie)
}

func RefreshToken(r *http.Request) string {
	return cookieValue(r, RefreshCookie)
}

func cookieValue(r *http.Request, name string) string {
	c, err := r.Cookie(name)
	if err != nil || c == nil {
		return ""
	}
	return strings.TrimSpace(c.Value)
}
