package config

import "time"

type Config struct {
	Server   ServerConfig
	App      AppConfig
	Email    EmailConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port int `env:"PORT" default:"8080"`
}

type JWTConfig struct {
	Secret   string        `env:"JWT_SECRET, required"`
	Expiry   time.Duration `env:"JWT_EXPIRY" default:"3600"`
	Issuer   string        `env:"JWT_ISSUER"`
	Audience string        `env:"JWT_AUDIENCE"`
}
type AppConfig struct {
	Environment string `env:"ENVIRONMENT" default:"development"`
}

type EmailConfig struct {
	SMTPHost        string        `env:"SMTP_HOST" default:"smtp.gmail.com"`
	SMTPPort        int           `env:"SMTP_PORT" default:"465"`
	SMTPUsername    string        `env:"SMTP_USERNAME"`
	SMTPPassword    string        `env:"SMTP_PASSWORD"`
	FromEmail       string        `env:"SMTP_FROM_EMAIL"`
	FromName        string        `env:"SMTP_FROM_NAME" default:"Inga"`
	TemplatesDir    string        `env:"SMTP_TEMPLATES_DIR" default:"internal/notification/emails/templates"`
	UseTLS          bool          `env:"SMTP_USE_TLS" default:"true"`
	Timeout         time.Duration `env:"SMTP_TIMEOUT" default:"15s"`
	VerificationURL string        `env:"EMAIL_VERIFICATION_URL" default:"http://localhost:8080/api/auth/verify?token="`
}

type DatabaseConfig struct {
	Host     string `env:"DATABASE_HOST"`
	Port     int    `env:"DATABASE_PORT"`
	Username string `env:"DATABASE_USERNAME"`
	Password string `env:"DATABASE_PASSWORD"`
	Database string `env:"DATABASE_NAME"`
	DSN      string `env:"DATABASE_DSN"`
}
