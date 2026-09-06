package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
	CORS     CORSConfig
	Apps     AppsConfig
}

type ServerConfig struct {
	Port int
	Mode string
}

type DatabaseConfig struct {
	Driver      string
	SQLitePath  string
	PostgresDSN string `mapstructure:"postgres_dsn"`
}

type JWTConfig struct {
	Secret            string `mapstructure:"secret"`
	AccessTTLMinutes  int    `mapstructure:"access_ttl_minutes"`
	RefreshTTLHours   int    `mapstructure:"refresh_ttl_hours"`
	SSOCodeTTLSeconds int    `mapstructure:"sso_code_ttl_seconds"`
	CookieDomain      string `mapstructure:"cookie_domain"`
	CookieSecure      bool   `mapstructure:"cookie_secure"`
	CookieSameSite    string `mapstructure:"cookie_samesite"`
}

type CORSConfig struct {
	AllowOrigins []string `mapstructure:"allow_origins"`
}

type AppsConfig struct {
	ProductCoreURL      string `mapstructure:"productcore_url"`
	SupplyCoreURL       string `mapstructure:"supplycore_url"`
	AfterSalesCoreURL   string `mapstructure:"aftersalescore_url"`
	StoreCoreURL        string `mapstructure:"storecore_url"`
	StoreSyncAgentURL   string `mapstructure:"storesyncagent_url"`
	WarehouseCoreURL    string `mapstructure:"warehousecore_url"`
	ShippingCoreURL     string `mapstructure:"shippingcore_url"`
	OrderCoreURL        string `mapstructure:"ordercore_url"`
	CustomerCoreURL     string `mapstructure:"customercore_url"`
	MallCoreURL         string `mapstructure:"mallcore_url"`
	MaterialCoreURL     string `mapstructure:"materialcore_url"`
	CatalogCoreURL      string `mapstructure:"catalogcore_url"`
	QuoteCoreURL        string `mapstructure:"quotecore_url"`
	TodoCenterURL       string `mapstructure:"todocenter_url"`
	SelfCoreURL         string `mapstructure:"selfcore_url"`
	OpsMobileURL        string `mapstructure:"opsmobile_url"`
	OsmsBackupURL       string `mapstructure:"osmsbackup_url"`
	AgentsCenterURL     string `mapstructure:"agentscenter_url"`
}

func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8091
	}
	if cfg.Database.Driver == "" {
		cfg.Database.Driver = "postgres"
	}
	if cfg.Database.SQLitePath == "" {
		cfg.Database.SQLitePath = "./data/usercore.db"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "dev-jwt-secret-change-in-production"
	}
	if cfg.JWT.AccessTTLMinutes == 0 {
		cfg.JWT.AccessTTLMinutes = 30
	}
	if cfg.JWT.RefreshTTLHours == 0 {
		cfg.JWT.RefreshTTLHours = 168
	}
	if cfg.JWT.SSOCodeTTLSeconds == 0 {
		cfg.JWT.SSOCodeTTLSeconds = 60
	}
	if cfg.JWT.CookieSameSite == "" {
		cfg.JWT.CookieSameSite = "Lax"
	}
	if cfg.Apps.ProductCoreURL == "" {
		cfg.Apps.ProductCoreURL = "http://localhost:5173"
	}
	if cfg.Apps.SupplyCoreURL == "" {
		cfg.Apps.SupplyCoreURL = "http://localhost:5175"
	}
	if cfg.Apps.AfterSalesCoreURL == "" {
		cfg.Apps.AfterSalesCoreURL = "http://localhost:5176"
	}
	if cfg.Apps.StoreCoreURL == "" {
		cfg.Apps.StoreCoreURL = "http://localhost:5179"
	}
	if cfg.Apps.StoreSyncAgentURL == "" {
		cfg.Apps.StoreSyncAgentURL = "http://localhost:5178"
	}
	if cfg.Apps.WarehouseCoreURL == "" {
		cfg.Apps.WarehouseCoreURL = "http://localhost:5180"
	}
	if cfg.Apps.ShippingCoreURL == "" {
		cfg.Apps.ShippingCoreURL = "http://localhost:5181"
	}
	if cfg.Apps.OrderCoreURL == "" {
		cfg.Apps.OrderCoreURL = "http://localhost:5182"
	}
	if cfg.Apps.CustomerCoreURL == "" {
		cfg.Apps.CustomerCoreURL = "http://localhost:5183"
	}
	if cfg.Apps.MallCoreURL == "" {
		cfg.Apps.MallCoreURL = "http://localhost:5184"
	}
	if cfg.Apps.MaterialCoreURL == "" {
		cfg.Apps.MaterialCoreURL = "http://localhost:5185"
	}
	if cfg.Apps.CatalogCoreURL == "" {
		cfg.Apps.CatalogCoreURL = "http://localhost:5188"
	}
	if cfg.Apps.QuoteCoreURL == "" {
		cfg.Apps.QuoteCoreURL = "http://localhost:5189"
	}
	if cfg.Apps.TodoCenterURL == "" {
		cfg.Apps.TodoCenterURL = "http://localhost:5186"
	}
	if cfg.Apps.SelfCoreURL == "" {
		cfg.Apps.SelfCoreURL = "http://localhost:5187"
	}
	if cfg.Apps.OpsMobileURL == "" {
		cfg.Apps.OpsMobileURL = "http://localhost:5190"
	}
	if cfg.Apps.OsmsBackupURL == "" {
		cfg.Apps.OsmsBackupURL = "http://localhost:5191"
	}
	if cfg.Apps.AgentsCenterURL == "" {
		cfg.Apps.AgentsCenterURL = "http://localhost:5192"
	}
	if len(cfg.CORS.AllowOrigins) == 0 {
		cfg.CORS.AllowOrigins = []string{
			"http://localhost:5173",
			"http://127.0.0.1:5173",
			"http://localhost:5174",
			"http://127.0.0.1:5174",
			"http://localhost:5175",
			"http://127.0.0.1:5175",
			"http://localhost:5176",
			"http://127.0.0.1:5176",
			"http://localhost:5178",
			"http://127.0.0.1:5178",
			"http://localhost:5179",
			"http://127.0.0.1:5179",
			"http://localhost:5180",
			"http://127.0.0.1:5180",
			"http://localhost:5181",
			"http://127.0.0.1:5181",
			"http://localhost:5182",
			"http://127.0.0.1:5182",
			"http://localhost:5183",
			"http://127.0.0.1:5183",
			"http://localhost:5184",
			"http://127.0.0.1:5184",
			"http://localhost:5185",
			"http://127.0.0.1:5185",
			"http://localhost:5186",
			"http://127.0.0.1:5186",
		}
	}
	return &cfg, nil
}
