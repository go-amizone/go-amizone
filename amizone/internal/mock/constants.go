package mock

// Constants for use in tests using the mock package to create Gock requests, etc.
const (
	ValidUser = "fakeUsername"
	ValidPass = "fakePassword"

	InvalidUser = "this-user-does-not-exist"
	InvalidPass = "this-password-does-not-either"

	AuthCookie        = "fakeAuthCookie"
	VerificationToken = "fakeRequestVerificationToken"
	SessionID         = "fakeSessionId"
)
