package docs

import (
	"strings"
	"testing"
)

func TestSwaggerDocumentContainsFoundationEndpoint(t *testing.T) {
	document := swaggerDoc{}.ReadDoc()
	for _, expected := range []string{"FIAP Video API", "/ping", "swagger"} {
		if !strings.Contains(document, expected) {
			t.Fatalf("document missing %q", expected)
		}
	}
}
