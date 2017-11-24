package common

func S2BA(value string) []byte {
	return []byte(value)
}

type ServerConfig struct {
	ID           uint `gorm:"primary_key" json:"id"`
	JwtSignature []byte
}
