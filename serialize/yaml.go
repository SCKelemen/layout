//go:build !no_yaml
// +build !no_yaml

package serialize

import (
	"gopkg.in/yaml.v3"

	"github.com/SCKelemen/layout"
)

// ToYAML converts a layout.Node to YAML bytes
// Requires: go get gopkg.in/yaml.v3
// To disable YAML support, build with: go build -tags no_yaml
func ToYAML(node *layout.Node) ([]byte, error) {
	// First convert to JSON structure
	nodeJSON := nodeToJSON(node)
	// Then convert to YAML
	return yaml.Marshal(nodeJSON)
}

// FromYAML converts YAML bytes to a layout.Node
// Requires: go get gopkg.in/yaml.v3
// To disable YAML support, build with: go build -tags no_yaml
func FromYAML(data []byte) (*layout.Node, error) {
	var nodeJSON NodeJSON
	if err := yaml.Unmarshal(data, &nodeJSON); err != nil {
		return nil, err
	}
	return jsonToNode(&nodeJSON), nil
}
