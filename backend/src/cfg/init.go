package cfg

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml/v2"
)

type ServerCfg struct {
	Port      string `toml:"port"`
	UploadDir string `toml:"upload_dir"`
}

type PgsqlCfg struct {
	Host     string `toml:"host"`
	Port     int    `toml:"port"`
	Username string `toml:"username"`
	Password string `toml:"password"`
	Database string `toml:"database"`
	Schema   string `toml:"schema"`
	SSLMode  string `toml:"sslmode"`
	MaxConn  int    `toml:"max_conn"`
	IdleConn int    `toml:"idle_conn"`
}

type RedisCfg struct {
	Host     string `toml:"host"`
	Password string `toml:"password"`
	Database int    `toml:"database"`
	MaxConn  int    `toml:"max_conn"`
	IdleConn int    `toml:"idle_conn"`
}

type AppConfig struct {
	Server ServerCfg `toml:"server"`
	Pgsql  PgsqlCfg  `toml:"pgsql"`
	Redis  RedisCfg  `toml:"redis"`

	Port      string
	PgsqlDSN  string
	UploadDir string
	Env       string
}

var Cfg = new(AppConfig)

func Init() {
	configDir := findConfigDir()

	// 1. 读取 config.toml 或分别读取 pgsql.toml / redis.toml
	configPath := filepath.Join(configDir, "config.toml")
	if data, err := os.ReadFile(configPath); err == nil {
		if err := toml.Unmarshal(data, Cfg); err != nil {
			panic(fmt.Errorf("解析 config.toml 失败: %w", err))
		}
	} else {
		// 分别尝试加载 pgsql.toml
		pgPath := filepath.Join(configDir, "pgsql.toml")
		if pgData, err := os.ReadFile(pgPath); err == nil {
			var wrapper struct {
				Pgsql PgsqlCfg `toml:"pgsql"`
			}
			if err := toml.Unmarshal(pgData, &wrapper); err != nil {
				panic(fmt.Errorf("解析 pgsql.toml 失败: %w", err))
			}
			Cfg.Pgsql = wrapper.Pgsql
		}

		// 分别尝试加载 redis.toml
		redisPath := filepath.Join(configDir, "redis.toml")
		if rdData, err := os.ReadFile(redisPath); err == nil {
			var wrapper struct {
				Redis RedisCfg `toml:"redis"`
			}
			if err := toml.Unmarshal(rdData, &wrapper); err != nil {
				panic(fmt.Errorf("解析 redis.toml 失败: %w", err))
			}
			Cfg.Redis = wrapper.Redis
		}
	}

	// 2. 环境变量覆盖与默认值回退
	if port := os.Getenv("PORT"); port != "" {
		Cfg.Server.Port = port
	}
	if Cfg.Server.Port == "" {
		Cfg.Server.Port = "8080"
	}
	Cfg.Port = Cfg.Server.Port

	if uploadDir := os.Getenv("UPLOAD_DIR"); uploadDir != "" {
		Cfg.Server.UploadDir = uploadDir
	}
	if Cfg.Server.UploadDir == "" {
		Cfg.Server.UploadDir = "./data/uploads"
	}
	Cfg.UploadDir = Cfg.Server.UploadDir

	if redisHost := os.Getenv("REDIS_HOST"); redisHost != "" {
		Cfg.Redis.Host = redisHost
	}

	if env := os.Getenv("APP_ENV"); env != "" {
		Cfg.Env = env
	} else if len(os.Args) > 1 {
		Cfg.Env = os.Args[1]
	} else {
		Cfg.Env = "dev"
	}

	// 3. 构建 PostgreSQL DSN 连接串
	if envDSN := os.Getenv("PG_DSN"); envDSN != "" {
		Cfg.PgsqlDSN = envDSN
	} else {
		Cfg.PgsqlDSN = buildPgsqlDSN(&Cfg.Pgsql)
	}

	if Cfg.PgsqlDSN == "" {
		panic(fmt.Errorf("PostgreSQL 配置不能为空"))
	}
}

func buildPgsqlDSN(c *PgsqlCfg) string {
	if c.Host == "" && c.Database == "" {
		return ""
	}
	if c.Host == "" {
		c.Host = "127.0.0.1"
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.Database == "" {
		c.Database = "hr_forms"
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}

	host := fmt.Sprintf("%s:%d", c.Host, c.Port)
	dsn := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(c.Username, c.Password),
		Host:   host,
		Path:   "/" + c.Database,
	}
	query := url.Values{}
	if c.SSLMode != "" {
		query.Set("sslmode", c.SSLMode)
	}
	if c.Schema != "" {
		query.Set("search_path", c.Schema)
	}
	dsn.RawQuery = query.Encode()
	return dsn.String()
}

func findConfigDir() string {
	var candidates []string

	// 1. 优先检查可执行文件所在目录旁边的 config
	if exePath, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exePath)
		candidates = append(candidates,
			filepath.Join(exeDir, "config"),
			filepath.Join(exeDir, "..", "config"),
			filepath.Join(exeDir, "..", "backend", "config"),
		)
	}

	// 2. 检查当前运行路径候选
	candidates = append(candidates,
		"config",
		"backend/config",
		"./backend/config",
		"../config",
	)

	for _, dir := range candidates {
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			return dir
		}
	}
	return "config"
}
