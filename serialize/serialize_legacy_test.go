package serialize

import (
	"encoding/json"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/SCKelemen/layout"
)

// Regression tests for files written by the previous encoder, which stored
// every length as a bare pixel number and unbounded lengths (the MaxSize of
// FractionTrack and AutoTrack, and occasionally Rect fields) as
// math.MaxFloat64. Such files must keep loading.

// legacyMaxFloat64 is how encoding/json and yaml.v3 print math.MaxFloat64.
const legacyMaxFloat64 = "1.7976931348623157e+308"

func TestLegacyGridWithUnboundedMaxSizeLoads(t *testing.T) {
	legacy := `{"style":{"display":"grid","gridTemplateColumns":[{"minSize":0,"maxSize":1.7976931348623157e+308,"fraction":1}],"width":100}}`
	out, err := FromJSON([]byte(legacy))
	if err != nil {
		t.Fatalf("FromJSON(legacy grid): %v", err)
	}
	if out.Style.Display != layout.DisplayGrid {
		t.Errorf("display = %v, want grid", out.Style.Display)
	}
	cols := out.Style.GridTemplateColumns
	if len(cols) != 1 {
		t.Fatalf("gridTemplateColumns = %+v, want one track", cols)
	}
	if cols[0].MaxSize != layout.PxUnbounded {
		t.Errorf("MaxSize = %+v, want PxUnbounded", cols[0].MaxSize)
	}
	if cols[0] != layout.FractionTrack(1) {
		t.Errorf("track = %+v, want FractionTrack(1)", cols[0])
	}
	if out.Style.Width != layout.Px(100) {
		t.Errorf("width = %+v, want Px(100)", out.Style.Width)
	}
}

func TestLegacyBareNumberForms(t *testing.T) {
	cases := []struct {
		in   string
		want layout.Length
	}{
		{"100", layout.Px(100)},
		{"0", layout.Px(0)},
		{"-1", layout.Px(-1)},
		{legacyMaxFloat64, layout.PxUnbounded},
	}
	for _, tc := range cases {
		// The same bare number as a style length and as a track bound.
		jsonIn := `{"style":{"width":` + tc.in + `,"gridAutoRows":{"minSize":0,"maxSize":` + tc.in + `}}}`
		out, err := FromJSON([]byte(jsonIn))
		if err != nil {
			t.Errorf("FromJSON(%s): %v", tc.in, err)
			continue
		}
		if out.Style.Width != tc.want || out.Style.GridAutoRows.MaxSize != tc.want {
			t.Errorf("JSON %s: width %+v, gridAutoRows.maxSize %+v, want %+v", tc.in, out.Style.Width, out.Style.GridAutoRows.MaxSize, tc.want)
		}

		yamlIn := "style:\n  width: " + tc.in + "\n  gridAutoRows:\n    minSize: 0\n    maxSize: " + tc.in + "\n"
		out, err = FromYAML([]byte(yamlIn))
		if err != nil {
			t.Errorf("FromYAML(%s): %v", tc.in, err)
			continue
		}
		if out.Style.Width != tc.want || out.Style.GridAutoRows.MaxSize != tc.want {
			t.Errorf("YAML %s: width %+v, gridAutoRows.maxSize %+v, want %+v", tc.in, out.Style.Width, out.Style.GridAutoRows.MaxSize, tc.want)
		}
	}

	// The sentinel must resolve to Unbounded, the value the grid engine checks.
	out, err := FromJSON([]byte(`{"style":{"maxWidth":` + legacyMaxFloat64 + `}}`))
	if err != nil {
		t.Fatal(err)
	}
	if layout.ResolveLength(out.Style.MaxWidth, nil, 16) != layout.Unbounded {
		t.Errorf("legacy MaxFloat64 resolved to %v, want Unbounded", layout.ResolveLength(out.Style.MaxWidth, nil, 16))
	}

	// A YAML literal that overflows float64 parses to +Inf and is unbounded too.
	out, err = FromYAML([]byte("style:\n  maxWidth: 1e309\n"))
	if err != nil {
		t.Fatalf("FromYAML(1e309): %v", err)
	}
	if out.Style.MaxWidth != layout.PxUnbounded {
		t.Errorf("YAML 1e309 read as %+v, want PxUnbounded", out.Style.MaxWidth)
	}
}

func TestLengthJSONUnmarshalMapsMaxFloat64ToUnbounded(t *testing.T) {
	var got struct{ A, B LengthJSON }
	if err := json.Unmarshal([]byte(`{"A":`+legacyMaxFloat64+`,"B":100}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.A != "unbounded" || got.B != "100px" {
		t.Errorf("got %+v", got)
	}
}

func TestLegacyBareNumbersOutOfRangeStillRejected(t *testing.T) {
	// Only the exact MaxFloat64 sentinel escapes the magnitude limit.
	for _, n := range []string{"1e13", "-1e13", "1e300", "-" + legacyMaxFloat64} {
		if _, err := FromJSON([]byte(`{"style":{"width":` + n + `}}`)); err == nil {
			t.Errorf("JSON width %s: expected error", n)
		}
		if _, err := FromJSON([]byte(`{"style":{"gridTemplateRows":[{"maxSize":` + n + `}]}}`)); err == nil {
			t.Errorf("JSON maxSize %s: expected error", n)
		}
		if _, err := FromYAML([]byte("style:\n  width: " + n + "\n")); err == nil {
			t.Errorf("YAML width %s: expected error", n)
		}
	}
	// Non-numeric words are not bare numbers and must not sneak in as lengths.
	for _, n := range []string{".nan", "nan", ".inf", "inf", "-.inf", "Infinity"} {
		if _, err := FromYAML([]byte("style:\n  width: " + n + "\n")); err == nil {
			t.Errorf("YAML width %s: expected error", n)
		}
		if _, err := FromJSON([]byte(`{"style":{"width":"` + n + `"}}`)); err == nil {
			t.Errorf("JSON width %q: expected error", n)
		}
	}
}

func TestRectUnboundedRoundTrip(t *testing.T) {
	in := &layout.Node{Rect: layout.Rect{X: 0, Y: 10, Width: math.MaxFloat64, Height: math.MaxFloat64}}

	data, err := ToJSON(in)
	if err != nil {
		t.Fatalf("ToJSON(unbounded rect): %v", err)
	}
	if !strings.Contains(string(data), `"width": `+legacyMaxFloat64) {
		t.Errorf("unbounded rect width should be written as %s:\n%s", legacyMaxFloat64, data)
	}
	out, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON(unbounded rect): %v", err)
	}
	if out.Rect != in.Rect {
		t.Errorf("rect round-tripped as %+v, want %+v", out.Rect, in.Rect)
	}

	ydata, err := ToYAML(in)
	if err != nil {
		t.Fatalf("ToYAML(unbounded rect): %v", err)
	}
	out, err = FromYAML(ydata)
	if err != nil {
		t.Fatalf("FromYAML(unbounded rect): %v\n%s", err, ydata)
	}
	if out.Rect != in.Rect {
		t.Errorf("YAML rect round-tripped as %+v, want %+v", out.Rect, in.Rect)
	}

	// A legacy file with the sentinel in rect.
	legacy := `{"style":{},"rect":{"x":0,"y":0,"width":` + legacyMaxFloat64 + `,"height":0}}`
	out, err = FromJSON([]byte(legacy))
	if err != nil {
		t.Fatalf("FromJSON(legacy rect): %v", err)
	}
	if out.Rect.Width != layout.Unbounded {
		t.Errorf("legacy rect width = %v, want Unbounded", out.Rect.Width)
	}

	// Everything else above the limit is still rejected, on both sides.
	for _, r := range []layout.Rect{{Width: 1e15}, {X: -1e13}, {Height: math.Inf(1)}, {Y: math.NaN()}} {
		if _, err := ToJSON(&layout.Node{Rect: r}); err == nil {
			t.Errorf("ToJSON(rect %+v): expected error", r)
		}
	}
	for _, input := range []string{
		`{"style":{},"rect":{"x":0,"y":0,"width":1e15,"height":0}}`,
		`{"style":{},"rect":{"x":0,"y":0,"width":-` + legacyMaxFloat64 + `,"height":0}}`,
	} {
		if _, err := FromJSON([]byte(input)); err == nil {
			t.Errorf("FromJSON(%s): expected error", input)
		}
	}
	if _, err := FromYAML([]byte("style: {}\nrect:\n  x: 0\n  y: 0\n  width: .inf\n  height: 0\n")); err == nil {
		t.Error("FromYAML(rect width .inf): expected error")
	}
}

func TestToJSONAcceptsLayoutOutputUnderUnboundedConstraints(t *testing.T) {
	tracks := []layout.GridTrack{layout.FractionTrack(1), layout.AutoTrack(), layout.FitContentTrack(300)}
	root := &layout.Node{
		Style: layout.Style{
			Display:             layout.DisplayGrid,
			GridTemplateColumns: tracks,
			GridAutoRows:        layout.AutoTrack(),
		},
		Children: []*layout.Node{
			{Style: layout.Style{Width: layout.Px(50), Height: layout.Px(20)}},
			{Style: layout.Style{Width: layout.Px(50), Height: layout.Px(20)}},
			{Style: layout.Style{Width: layout.Px(50), Height: layout.Px(20)}},
		},
	}
	ctx := layout.NewLayoutContext(800, 600, 16)
	layout.Layout(root, layout.Loose(layout.Unbounded, layout.Unbounded), ctx)

	data, err := ToJSON(root)
	if err != nil {
		t.Fatalf("ToJSON after Layout under unbounded constraints: %v", err)
	}
	out, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON: %v\n%s", err, data)
	}
	if !reflect.DeepEqual(out.Style.GridTemplateColumns, tracks) {
		t.Errorf("tracks round-tripped as %+v", out.Style.GridTemplateColumns)
	}
	if out.Rect != root.Rect {
		t.Errorf("root rect round-tripped as %+v, want %+v", out.Rect, root.Rect)
	}
	for i, c := range out.Children {
		if c.Rect != root.Children[i].Rect {
			t.Errorf("child %d rect round-tripped as %+v, want %+v", i, c.Rect, root.Children[i].Rect)
		}
	}
	if _, err := ToYAML(root); err != nil {
		t.Errorf("ToYAML after Layout: %v", err)
	}
}
