package frame

const (
	PingFrame = iota
	AuthFrame
	AuthErrorFrame
	JWTFrame
	CategoryFrame
	TopicFrame
	CommentFrame
	UserFrame
)

type Frame struct {
	Type int    `json:"type"`
	Data string `json:"data"`
}
