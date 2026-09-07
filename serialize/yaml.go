//go:build !no_yaml
// +build !no_yaml

package serialize

import (
	"gopkg.in/yaml.v3"

	"github.com/SCKelemen/layout"
)

// ToYAML converts a layout.Node to YAML bytes.
// Keys match the JSON field names (camelCase) and lengths use the same
// CSS-like string form ("10px", "2em", "unbounded").
// Requires: go get gopkg.in/yaml.v3
// To disable YAML support, build with: go build -tags no_yaml
func ToYAML(node *layout.Node) ([]byte, error) {
	// First convert to the shared wire structure
	nodeJSON, err := nodeToJSON(node)
	if err != nil {
		return nil, err
	}
	// Then convert to YAML
	return yaml.Marshal(nodeJSON)
}

// FromYAML converts YAML bytes to a layout.Node.
// It applies the same validation as FromJSON (finite numbers, known enum
// keywords, MaxTreeDepth/MaxChildren). A bare number is accepted for any
// length and interpreted as pixels. An empty document or a bare null
// returns ErrNullInput.
// Requires: go get gopkg.in/yaml.v3
// To disable YAML support, build with: go build -tags no_yaml
func FromYAML(data []byte) (*layout.Node, error) {
	var nodeJSON *NodeJSON
	if err := yaml.Unmarshal(data, &nodeJSON); err != nil {
		return nil, err
	}
	if nodeJSON == nil {
		return nil, ErrNullInput
	}
	return jsonToNode(nodeJSON)
}
