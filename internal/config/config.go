package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Env      string         `mapstructure:"env"`
	LogLevel string         `mapstructure:"log_level"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Encrypt  EncryptConfig  `mapstructure:"encrypt"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Addr         string `mapstructure:"addr"`
	ReadTimeout  int    `mapstructure:"read_timeout"`
	WriteTimeout int    `mapstructure:"write_timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"` // mysql or postgres
	DSN             string `mapstructure:"dsn"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	MaxOpenConns    int    `mapstructure:"max_open_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	TLS      bool   `mapstructure:"tls"` // 是否启用 TLS
}

// JWTConfig JWT 配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

// EncryptConfig 加密配置
type EncryptConfig struct {
	Key string `mapstructure:"key"` // AES-256 密钥 (32 bytes)
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/confighub")

	// 环境变量支持
	viper.SetEnvPrefix("CONFIGHUB")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// 默认值
	setDefaults()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
		// 配置文件不存在时使用默认值和环境变量
	}

	// 从标准环境变量构建 DSN (支持 Railway/Render/Fly.io 等平台)
	overrideFromEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// overrideFromEnv 从标准环境变量覆盖配置 (支持云平台)
func overrideFromEnv() {
	// 数据库配置 - 支持 DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME
	dbHost := getEnv("DB_HOST", "MYSQL_HOST", "DATABASE_HOST")
	dbPort := getEnv("DB_PORT", "MYSQL_PORT", "DATABASE_PORT")
	dbUser := getEnv("DB_USER", "MYSQL_USER", "DATABASE_USER")
	dbPassword := getEnv("DB_PASSWORD", "MYSQL_PASSWORD", "DATABASE_PASSWORD")
	dbName := getEnv("DB_NAME", "MYSQL_DATABASE", "DATABASE_NAME")
	dbDriver := getEnv("DB_DRIVER", "DATABASE_DRIVER")

	if dbHost != "" && dbUser != "" && dbPassword != "" && dbName != "" {
		if dbPort == "" {
			dbPort = "3306"
		}
		if dbDriver == "" {
			dbDriver = "mysql"
		}

		// 检查是否需要 TLS (通过 DB_TLS 环境变量或自动检测 TiDB Cloud)
		dbTLS := os.Getenv("DB_TLS")
		if dbTLS == "" {
			// 自动检测 TiDB Cloud (域名包含 tidbcloud.com)
			if strings.Contains(dbHost, "tidbcloud.com") {
				dbTLS = "true"
			}
		}

		var dsn string
		if dbDriver == "postgres" {
			sslMode := "disable"
			if dbTLS == "true" {
				sslMode = "require"
			}
			dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
				dbHost, dbPort, dbUser, dbPassword, dbName, sslMode)
		} else {
			// MySQL DSN
			tlsParam := ""
			if dbTLS == "true" {
				tlsParam = "&tls=true"
			}
			dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local%s",
				dbUser, dbPassword, dbHost, dbPort, dbName, tlsParam)
		}
		viper.Set("database.dsn", dsn)
		viper.Set("database.driver", dbDriver)
	}

	// Redis 配置 - 支持 REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_URL
	redisURL := os.Getenv("REDIS_URL")
	if redisURL != "" {
		// Railway/Render 风格的 Redis URL
		viper.Set("redis.addr", redisURL)
	} else {
		redisHost := getEnv("REDIS_HOST", "REDIS_HOSTNAME")
		redisPort := getEnv("REDIS_PORT")
		if redisHost != "" {
			if redisPort == "" {
				redisPort = "6379"
			}
			viper.Set("redis.addr", fmt.Sprintf("%s:%s", redisHost, redisPort))
			
			// 自动检测 Upstash Redis (域名包含 upstash.io)
			if strings.Contains(redisHost, "upstash.io") {
				viper.Set("redis.tls", true)
			}
		}
	}
	if redisPassword := os.Getenv("REDIS_PASSWORD"); redisPassword != "" {
		viper.Set("redis.password", redisPassword)
	}
	// 支持 REDIS_TLS 环境变量
	if redisTLS := os.Getenv("REDIS_TLS"); redisTLS == "true" {
		viper.Set("redis.tls", true)
	}

	// JWT 配置
	if jwtSecret := getEnv("JWT_SECRET", "JWT_KEY"); jwtSecret != "" {
		viper.Set("jwt.secret", jwtSecret)
	}

	// 加密密钥
	if encryptKey := getEnv("ENCRYPT_KEY", "ENCRYPTION_KEY"); encryptKey != "" {
		viper.Set("encrypt.key", encryptKey)
	}

	// 服务器端口 (支持 PORT 环境变量 - 云平台标准)
	if port := os.Getenv("PORT"); port != "" {
		viper.Set("server.addr", ":"+port)
	}
}

// getEnv 从多个环境变量名中获取第一个非空值
func getEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

func setDefaults() {
	viper.SetDefault("env", "development")
	viper.SetDefault("log_level", "info")

	viper.SetDefault("server.addr", ":8080")
	viper.SetDefault("server.read_timeout", 30)
	viper.SetDefault("server.write_timeout", 30)

	viper.SetDefault("database.driver", "mysql")
	viper.SetDefault("database.dsn", "root:password@tcp(localhost:3306)/confighub?charset=utf8mb4&parseTime=True&loc=Local")
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.conn_max_lifetime", 3600)

	viper.SetDefault("redis.addr", "localhost:6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	viper.SetDefault("jwt.secret", "confighub-jwt-secret-key-change-in-production")
	viper.SetDefault("jwt.expire_hour", 24)

	viper.SetDefault("encrypt.key", "confighub-encrypt-key-32bytes!")
}
