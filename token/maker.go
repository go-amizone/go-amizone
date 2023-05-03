package token 

type Maker interface {
	CreateToken(email string, password string) (string, error)
	VerifyToken(token string) (*Payload, error)
}