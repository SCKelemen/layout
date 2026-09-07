//go:build !no_yaml

package serialize

func init() {
	yamlAvailable = true
	yamlEncode = ToYAML
	yamlDecode = FromYAML
}
