package lrs

import "encoding/json"

// List of all allowed frames
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
	CommentTypingFrame  = "CommentTyping"
	UserFrame           = "User"
	ResetPasswordFrame  = "ResetPassword"
	FileUploadFrame     = "Upload"
	CancelUploadFrame   = "CancelUpload"
	ErrorFrame          = "Error"
)

// Frame is a core envelop used to encapsulate different data structures
type Frame struct {
	Type      string `json:"type"`
	Data      string `json:"data"`
	RequestID string `json:"requestId"`
}

// BindJSON unmarshals JSON to an object
func (frame Frame) BindJSON(obj interface{}) error {
	return json.Unmarshal([]byte(frame.Data), obj)
}

// NewFrame returns a new Frame given a type, object and request identifier
func NewFrame(t string, obj interface{}, requestID string) Frame {
	r, _ := json.Marshal(obj)
	return Frame{
		Type:      t,
		Data:      string(r),
		RequestID: requestID,
	}
}
