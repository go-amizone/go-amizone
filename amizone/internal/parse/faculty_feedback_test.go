package parse_test

import (
	"testing"

	"github.com/ditsuke/go-amizone/amizone/internal/mock"
	"github.com/ditsuke/go-amizone/amizone/internal/parse"
	. "github.com/onsi/gomega"
)

func TestFacultyFeedback(t *testing.T) {
	g := NewWithT(t)
	r, err := mock.FacultyPage.Open()
	g.Expect(err).ToNot(HaveOccurred())

	spec, err := parse.FacultyFeedback(r)
	g.Expect(err).ToNot(HaveOccurred())

	expected := ReadExpectedFile(mock.ExpectedFacultyFeedbackSpec, g)
	g.Expect(toJSON(spec, g)).To(MatchJSON(expected))
}
