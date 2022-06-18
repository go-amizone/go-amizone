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

	// StudentUUID is the UUID associated with the student used across testdata in fileSystem.
	StudentUUID = "98RFGK88-A01C-1JJO-N73D-4BJR42B33J51"
)
