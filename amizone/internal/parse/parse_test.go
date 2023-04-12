package parse_test

import (
	"encoding/json"
	"io"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	. "github.com/onsi/gomega"
)

// Constants used across the tests
const ()

// === Test helpers ===

// toJSON converts a struct to a JSON string.
func toJSON[T any](t T, g *WithT) string {
	s, err := json.Marshal(t)
	g.Expect(err).ToNot(HaveOccurred(), "marshall json")
	return string(s)
}

func ReadExpectedFile(file mock.ExpectedJSON, g *WithT) []byte {
	f, err := file.Open()
	g.Expect(err).ToNot(HaveOccurred(), "open expected data file")
	b, err := io.ReadAll(f)
	g.Expect(err).ToNot(HaveOccurred(), "read expected data file")
	return b
}
