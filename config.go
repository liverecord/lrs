package server

type Config struct {
	ID           uint   `gorm:"primary_key" json:"id"`
	JwtSignature []byte `gorm:"type:varbinary(256)"`
}
