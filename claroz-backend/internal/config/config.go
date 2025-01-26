package config

type Config struct {
	Database   DatabaseConfig
	Server     ServerConfig
	Storage    StorageConfig
	Federation FederationConfig
}

type FederationConfig struct {
	PDSHost string // AT Protocol PDS host (e.g. "https://bsky.social")
	Enabled bool   // Whether federation is enabled
}

type StorageConfig struct {
	Provider    string // "local" or "s3"
	LocalPath   string // for local storage
	S3Bucket    string // for S3 storage
	S3Region    string // for S3 storage
	MaxFileSize int64  // maximum file size in bytes
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type ServerConfig struct {
	Port string
	Mode string
}

func NewConfig() *Config {
	return &Config{
		Database: DatabaseConfig{
			Host:     "localhost",
			Port:     "5432",
			User:     "postgres",
			Password: "postgres",
			DBName:   "claroz",
			SSLMode:  "disable",
		},
		Server: ServerConfig{
			Port: "8080",
			Mode: "debug",
		},
		Storage: StorageConfig{
			Provider:    "local",
			LocalPath:   "./uploads",
			S3Bucket:    "",
			S3Region:    "",
			MaxFileSize: 5 * 1024 * 1024, // 5MB
		},
		Federation: FederationConfig{
			PDSHost: "https://bsky.social",
			Enabled: true,
		},
	}
}
