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
	Secret           string `mapstructure:"secret"`
	AccessTTLMinutes int    `mapstructure:"access_ttl_minutes"`
	RefreshTTLHours  int    `mapstructure:"refresh_ttl_hours"`
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
		cfg.JWT.AccessTTLMinutes = 120
	}
	if cfg.JWT.RefreshTTLHours == 0 {
		cfg.JWT.RefreshTTLHours = 168
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
			"http://localhost:5179",
			"http://127.0.0.1:5179",
			"http://localhost:5180",
			"http://127.0.0.1:5180",
			"http://localhost:5181",
			"http://127.0.0.1:5181",
		}
	}
	return &cfg, nil
}
