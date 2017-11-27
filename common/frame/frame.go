package frame

const (
	PingFrame = iota
	AuthFrame
	AuthErrorFrame
	JWTFrame
	UserListFrame
	UserInfoFrame
	UserUpdateFrame
	UserDeleteFrame
	CategoryFrame
	CategoryErrorFrame
	TopicFrame
	CommentFrame
	UserFrame
)

type Frame struct {
	Type      int    `json:"type"`
	Data      string `json:"data"`
	RequestID string `json:"ri"`
}
