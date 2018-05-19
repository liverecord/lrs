package lrs

import "encoding/json"

const (
	PingFrame           = "Ping"
	AuthFrame           = "Auth"
	AuthErrorFrame      = "AuthError"
	JWTFrame            = "JWT"
	UserListFrame       = "UserList"
	UserInfoFrame       = "UserInfo"
	UserUpdateFrame     = "UserUpdate"
	UserDeleteFrame     = "UserDelete"
	CategoryFrame       = "Category"
	CategoryListFrame   = "CategoryList"
	CategorySaveFrame   = "CategorySave"
	CategoryUpdateFrame = "CategoryUpdate"
	CategoryDeleteFrame = "CategoryDelete"
	CategoryErrorFrame  = "CategoryError"
	TopicFrame          = "Topic"
	TopicSaveFrame      = "TopicSave"
	TopicListFrame      = "TopicList"
	CommentFrame        = "Comment"
	CommentListFrame    = "CommentList"
	CommentSaveFrame    = "CommentSave"
	UserFrame           = "User"
	ResetPasswordFrame  = "ResetPassword"
	FileUploadFrame     = "Upload"
	CancelUploadFrame   = "CancelUpload"
)

type Frame struct {
	Type      string `json:"type"`
	Data      string `json:"data"`
	RequestID string `json:"requestId"`
}

func (frame Frame) BindJSON(obj interface{}) error {
	return json.Unmarshal([]byte(frame.Data), obj)
}

func NewFrame(t string, obj interface{}, requestId string) Frame {
	r, _ := json.Marshal(obj)
	return Frame{
		Type:      t,
		Data:      string(r),
		RequestID: requestId,
	}
}
