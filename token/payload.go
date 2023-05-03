package token

type Payload struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewPayload(userID string, email string, password string) *Payload {
	return &Payload{
		UserID:   userID,
		Email:    email,
		Password: password,
	}
}
