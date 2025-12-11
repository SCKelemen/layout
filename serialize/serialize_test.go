package serialize

import (
	"encoding/json"
	"testing"

	"github.com/SCKelemen/layout"
)

func TestToJSON(t *testing.T) {
	// Create a simple layout tree
	root := &layout.Node{
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         200,
			Height:        100,
		},
		Children: []*layout.Node{
			{
				Style: layout.Style{
					Width:  100,
					Height: 50,
				},
				Rect: layout.Rect{
					X:      0,
					Y:      0,
					Width:  100,
					Height: 50,
				},
			},
		},
		Rect: layout.Rect{
			X:      0,
			Y:      0,
			Width:  200,
			Height: 100,
		},
	}

	jsonBytes, err := ToJSON(root)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	// Verify it's valid JSON
	var jsonData interface{}
	if err := json.Unmarshal(jsonBytes, &jsonData); err != nil {
		t.Fatalf("Generated JSON is invalid: %v", err)
	}

	// Verify we can deserialize it back
	deserialized, err := FromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	// Verify basic properties
	if deserialized.Style.Display != root.Style.Display {
		t.Errorf("Display mismatch: got %v, want %v", deserialized.Style.Display, root.Style.Display)
	}
	if deserialized.Style.FlexDirection != root.Style.FlexDirection {
		t.Errorf("FlexDirection mismatch: got %v, want %v", deserialized.Style.FlexDirection, root.Style.FlexDirection)
	}
	if deserialized.Style.Width != root.Style.Width {
		t.Errorf("Width mismatch: got %v, want %v", deserialized.Style.Width, root.Style.Width)
	}
	if deserialized.Rect.Width != root.Rect.Width {
		t.Errorf("Rect.Width mismatch: got %v, want %v", deserialized.Rect.Width, root.Rect.Width)
	}
	if len(deserialized.Children) != len(root.Children) {
		t.Errorf("Children count mismatch: got %d, want %d", len(deserialized.Children), len(root.Children))
	}
}

func TestGridSerialization(t *testing.T) {
	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateRows: []layout.GridTrack{
				layout.FixedTrack(100),
				layout.FixedTrack(200),
			},
			GridTemplateColumns: []layout.GridTrack{
				layout.FractionTrack(1),
				layout.FractionTrack(2),
			},
			GridGap: 10,
		},
	}

	jsonBytes, err := ToJSON(root)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	deserialized, err := FromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if deserialized.Style.Display != layout.DisplayGrid {
		t.Errorf("Display mismatch: got %v, want %v", deserialized.Style.Display, layout.DisplayGrid)
	}
	if len(deserialized.Style.GridTemplateRows) != 2 {
		t.Errorf("GridTemplateRows count mismatch: got %d, want 2", len(deserialized.Style.GridTemplateRows))
	}
	if len(deserialized.Style.GridTemplateColumns) != 2 {
		t.Errorf("GridTemplateColumns count mismatch: got %d, want 2", len(deserialized.Style.GridTemplateColumns))
	}
	if deserialized.Style.GridGap != 10 {
		t.Errorf("GridGap mismatch: got %v, want 10", deserialized.Style.GridGap)
	}
}

func TestAspectRatioSerialization(t *testing.T) {
	root := &layout.Node{
		Style: layout.Style{
			Width:       800,
			Height:      -1, // auto
			AspectRatio: 16.0 / 9.0,
		},
	}

	jsonBytes, err := ToJSON(root)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	deserialized, err := FromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if deserialized.Style.AspectRatio != root.Style.AspectRatio {
		t.Errorf("AspectRatio mismatch: got %v, want %v", deserialized.Style.AspectRatio, root.Style.AspectRatio)
	}
}

func TestTransformSerialization(t *testing.T) {
	root := &layout.Node{
		Style: layout.Style{
			Transform: layout.Translate(10, 20),
		},
	}

	jsonBytes, err := ToJSON(root)
	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	deserialized, err := FromJSON(jsonBytes)
	if err != nil {
		t.Fatalf("FromJSON failed: %v", err)
	}

	if deserialized.Style.Transform.A != root.Style.Transform.A {
		t.Errorf("Transform.A mismatch: got %v, want %v", deserialized.Style.Transform.A, root.Style.Transform.A)
	}
}

