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
	ID                  uint   `gorm:"primary_key" json:"id"`
	JwtSignature        []byte `gorm:"type:varbinary(256)" json:"-"`
	Protocol            string
	Port                uint 	`gorm:"default:80"`
	Name                string
	Domain              string
	LogoPath            string
	UploadDir           string `json:"-" gorm:"default:'files'"`
	DocumentRoot        string `json:"-"`
	SMTP                SMTP   `gorm:"embedded;embedded_prefix:smtp_" json:"-"`
	Debug               bool   `json:"debug" gorm:"default:0"`
	GAID                string `json:""`
	FacebookClientID    string
	FacebookSecret      string `json:"-"`
	TwitterClientID     string
	TwitterSecret       string `json:"-"`
	WindowsLiveClientID string
	WindowsLiveSecret   string `json:"-"`
	VkClientID          string
	VkSecret            string `json:"-"`
	GitHubClientID      string
	GitHubSecret        string `json:"-"`
	GoogleClientID      string
	GoogleSecret        string `json:"-"`
	GoogleCMAPIKey      string
	VapidPublicKey      string
	VapidPrivateKey     string `json:"-"`
	CommentsPerPage     uint `json:"commentsPerPage" gorm:"default:100"`
	TopicsPerPage     	uint `json:"topicsPerPage" gorm:"default:100"`
}

// SiteURL Returns site URL
func (cfg *Config) SiteURL() string {
	return cfg.Protocol + "://" + cfg.Domain + "/"
}

// LogoURL returns link to the logo
func (cfg *Config) LogoURL() string {
	return cfg.Protocol + "://" + cfg.Domain + "/" + cfg.LogoPath
}
