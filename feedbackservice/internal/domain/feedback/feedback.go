package feedback

import "errors"

type Feedback struct {
	ID     int64  `json:"id"`
	Score  int    `json:"score"`
	Text   string `json:"text,omitempty"`
	RentID string `json:"rent_id"`
}

type FeedbackLocal struct {
	ID     int64  `json:"id"`
	Text   string `json:"text"`
	Type   string `json:"type,omitempty"`
	UserID string `json:"user_id"`
}

type FeedbackPartnership struct {
	ID          int64  `json:"id"`
	UserID      string `json:"user_id"`
	ContactName string `json:"contact_name"`
	CompanyName string `json:"company_name"`
	Email       string `json:"email" binding:"email"`
	PhoneNum    string `json:"phone_num" binding:"e164"`
	Comment     string `json:"comment"`
}

var (
	ErrInvalidRequest = errors.New("invalid request")
	ErrFeedbackExists = errors.New("feedback for this rent already exists")
)
