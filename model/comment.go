package model

type Comment struct {
	Model
	Topic       Topic
	User        User
	Body        string `json:"body"`
	Attachments []Attachment
}

type Attachment struct {
	Model
	Type        string `json:"type"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Thumbnail   string `json:"thumbnail"`
	Html        string `json:"html"`
}
