package lrs

import (
	"crypto/rand"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/jinzhu/gorm"
	"github.com/liverecord/lrs/common"
	"io"
	"io/ioutil"
	"os"
)

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
	Port                uint `gorm:"default:80"`
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
	CommentsPerPage     uint   `json:"commentsPerPage" gorm:"default:100"`
	TopicsPerPage       uint   `json:"topicsPerPage" gorm:"default:100"`
}

// SiteURL Returns site URL
func (cfg *Config) SiteURL() string {
	return cfg.Protocol + "://" + cfg.Domain + "/"
}

// LogoURL returns link to the logo
func (cfg *Config) LogoURL() string {
	return cfg.Protocol + "://" + cfg.Domain + "/" + cfg.LogoPath
}

// NewConfig instantiates Config
func NewConfig(db *gorm.DB, logger *logrus.Logger) *Config {
	var config Config
	var err error

	db.First(&config)
	if config.ID == 0 {
		// lets set this application with default parameters
		config.JwtSignature = make([]byte, 256)
		if _, err = io.ReadFull(rand.Reader, config.JwtSignature); err != nil {
			logger.WithError(err).Errorln("Unable to generate JWT Signature")
		}
		config.DocumentRoot = common.Env("DOCUMENT_ROOT", "assets")
		config.Domain = common.Env("DOMAIN", "localhost")
		config.Protocol = common.Env("PROTOCOL", "http")
		config.Port = uint(common.IntEnv("PORT", 80))
		config.SMTP.Host = common.Env("SMTP_HOST", "localhost")
		config.SMTP.Port = common.IntEnv("SMTP_PORT", 25)
		config.SMTP.Username = common.Env("SMTP_USERNAME", "")
		config.SMTP.Password = common.Env("SMTP_PASSWORD", "")
		config.SMTP.InsecureTLS = common.BoolEnv("SMTP_INSECURE_TLS", false)
		config.SMTP.SSL = common.BoolEnv("SMTP_SSL", false)
		config.UploadDir, err = ioutil.TempDir("/tmp", "lr_")
		if err != nil {
			logger.WithError(err).Errorln("Unable to create temporary dir. Is '/tmp' writable?")
		}
		db.Save(&config)
	}
	config.Debug = common.BoolEnv("DEBUG", false)

	return &config
}

// Line identifies one prompt line for .env parameters
type Line struct {
	param        string
	tip          string
	defaultValue string
}

func promptOption(option Line) string {
	fmt.Printf("\n%s", option.tip)
	if len(option.defaultValue) > 0 {
		fmt.Printf(" (default: %s)", option.defaultValue)
	}
	fmt.Print(": ")
	var o string
	n, err := fmt.Scanln(&o)
	if err != nil {
		if n == 0 {
			fmt.Printf("Used default value for %s parameter.", option.param)
			return option.defaultValue
		}
		return promptOption(option)
	}
	return o
}

// InteractiveSetup provides user-friendly way to create configuration
func InteractiveSetup(logger *logrus.Logger) {
	fmt.Println("Welcome to LiveRecord interactive setup.")
	cwd, _ := os.Getwd()
	var options = []Line{
		{"DOCUMENT_ROOT", "Document root", cwd},
		{"DOMAIN", "Public App Domain (e.g.: example.com)", "localhost"},
		{"PROTOCOL", "Public Protocol", "http"},
		{"PORT", "Public Port", "80"},
		{"SMTP_HOST", "SMTP Host", "smtp.google.com"},
		{"SMTP_PORT", "SMTP Port", "25"},
		{"SMTP_USERNAME", "SMTP Username (e.g.: someone@gmail.com)", ""},
		{"SMTP_PASSWORD", "SMTP Password", ""},
		{"SMTP_INSECURE_TLS", "Enable insecure TLS for SMTP (true or false, not recommended in production)?", "false"},
		{"SMTP_SSL", "Use SSL for SMTP (true or false)?", "true"},
		{"MYSQL_DSN", "MySQL database source name", "root:123@tcp(127.0.0.1:3306)/liveRecord?charset=utf8&parseTime=True"},
		{"LISTEN_ADDR", "Listen on host:port", "127.0.0.1:8000"},
		{"DEBUG", "Enable debug mode (true or false)?", "false"},
		{"FACEBOOK_APP_ID", "Facebook Application Id (visit https://developers.facebook.com/ to get one)", ""},
		{"FACEBOOK_APP_SECRET", "Facebook Application Secret", ""},
	}

	f, err := os.OpenFile(".env", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		logger.Errorln("Cannot create .env file in this directory")
		return
	}
	for _, v := range options {
		f.Write([]byte(fmt.Sprintf("%s=%s\n", v.param, promptOption(v))))
	}
	f.Close()
	fmt.Println("All set! Reloading configuration...")
}
