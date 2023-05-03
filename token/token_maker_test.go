package token

import "testing"

func TestTokenMaker(t *testing.T) {

	maker, err := NewPasetoMaker("VgD4zPCTPw7HJywbVa7FFgIeW8KWamCX")
	if err != nil {
		t.Fatal(err)
	}

	email := "bruh@gmail.com"
	password := "password"

	// Create token
	token, err := maker.CreateToken(email, password)
	if err != nil {
		t.Fatal(err)
	}

	// Verify token
	payload, err := maker.VerifyToken(token)
	if err != nil {
		t.Fatal(err)
	}

	if payload.Email != email {
		t.Errorf("payload.Email != email")
	}

	if payload.Password != password {
		t.Errorf("payload.Password != password")
	}

}
