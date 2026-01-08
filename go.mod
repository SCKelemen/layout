module github.com/SCKelemen/layout

go 1.24.0

require (
	github.com/SCKelemen/text v0.0.0-00010101000000-000000000000
	github.com/SCKelemen/unicode v1.0.1-0.20251225190048-233be2b0d647
	gopkg.in/yaml.v3 v3.0.1
)

require github.com/SCKelemen/units v0.0.0-20251215145938-c61f55703fef // indirect

require (
	cel.dev/expr v0.24.0 // indirect
	github.com/SCKelemen/wpt-test-gen v0.0.0-00010101000000-000000000000
	github.com/antlr4-go/antlr/v4 v4.13.0 // indirect
	github.com/google/cel-go v0.26.1 // indirect
	github.com/stoewer/go-strcase v1.2.0 // indirect
	golang.org/x/exp v0.0.0-20230515195305-f3d0a9c9a5cc // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20240826202546-f6391c0de4c7 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240826202546-f6391c0de4c7 // indirect
	google.golang.org/protobuf v1.34.2 // indirect
)

replace github.com/SCKelemen/wpt-test-gen => ../wpt-test-gen

replace github.com/SCKelemen/text => ../text

replace github.com/SCKelemen/unicode => ../unicode
