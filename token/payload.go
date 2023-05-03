package token

type Payload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func NewPayload(email string, password string) *Payload {
	return &Payload{
		Email:    email,
		Password: password,
	}
}
