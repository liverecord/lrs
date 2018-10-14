package lrs

// ErrorResponse defines the basic error response structure
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// Predefined objects identifiers
// 01: User
// 02: Topic
// 03: Comment
// 04: Category
const (
	// GeneralError test
	GeneralError = 0
	// WrongPassword tells about wrong password
	WrongPassword = 1
	// NoPasswordReset can't reset password
	NoPasswordReset = 2

	// TopicNotFound lets us know that topic cannot be located
	TopicNotFound = 40402
	TopicNotAccessible = 40302
)