package frame

import "encoding/json"

const (
	PingFrame           = "Ping"
	AuthFrame           = "Auth"
	AuthErrorFrame      = "AuthError"
	LoginFrame          = "Login"
	JWTFrame            = "JWT"
	UserListFrame       = "UserList"
	UserInfoFrame       = "UserInfo"
	UserUpdateFrame     = "UserUpdate"
	UserDeleteFrame     = "UserDelete"
	CategoryFrame       = "Category"
	CategoryListFrame   = "CategoryList"
	CategoryUpdateFrame = "CategoryUpdate"
	CategoryDeleteFrame = "CategoryDelete"
	CategoryErrorFrame  = "CategoryError"
	TopicFrame          = "Topic"
	TopicSaveFrame      = "TopicSave"
	TopicListFrame      = "TopicList"
	TopicDeleteFrame    = "TopicDelete"
	CommentFrame        = "Comment"
	UserFrame           = "User"
)

type Frame struct {
	Type      string `json:"type"`
	Data      string `json:"data"`
	RequestID string `json:"requestId"`
}

func (frame Frame) BindJSON(obj interface{}) error {
	return json.Unmarshal([]byte(frame.Data), obj)
}
