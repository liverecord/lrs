package lrs

// SMTP used to define SMTP configuration
type SMTP struct {
	From        string
	Host        string
	Port        int
	Username    string
	Password    string
	InsecureTLS bool
	SSL         bool
}

// Config defines app configuration
type Config struct {
	ID           uint   `gorm:"primary_key" json:"id"`
	JwtSignature []byte `gorm:"type:varbinary(256)"`
	Protocol     string
	Port         uint
	Name         string
	Domain       string
	LogoPath     string
	UploadDir    string
	DocumentRoot string
	SMTP         SMTP `gorm:"embedded;embedded_prefix:smtp_"`
	Debug        bool
}

// SiteURL Returns site URL
func (cfg *Config) SiteURL() string {
	return cfg.Protocol + "://" + cfg.Domain + "/"
}

// LogoURL returns link to the logo
func (cfg *Config) LogoURL() string {
	return cfg.Protocol + "://" + cfg.Domain + "/" + cfg.LogoPath
}
