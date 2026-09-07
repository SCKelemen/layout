package serialize

import (
	"errors"
	"testing"

	"github.com/SCKelemen/layout"
)

// The CI matrix builds this package with -tags no_yaml, which removes
// ToYAML/FromYAML. Tests reach YAML through these indirections so the test
// binary still compiles without the YAML dependency; yaml_helpers_enabled_test.go
// wires them to the real functions when YAML is built in.
var errNoYAML = errors.New("serialize: YAML support disabled (no_yaml build tag)")

var yamlAvailable bool

var yamlEncode = func(*layout.Node) ([]byte, error) { return nil, errNoYAML }

var yamlDecode = func([]byte) (*layout.Node, error) { return nil, errNoYAML }

func skipIfNoYAML(t *testing.T) {
	t.Helper()
	if !yamlAvailable {
		t.Skip("YAML support disabled (no_yaml build tag)")
	}
}
