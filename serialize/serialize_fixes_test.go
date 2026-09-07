package serialize

import (
	"encoding/json"
	"errors"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/SCKelemen/layout"
)

// Regression tests for round-trip fidelity (length units, missing fields,
// missing enum cases, no-op omitempty) and for input validation.

func roundTrip(t *testing.T, node *layout.Node) *layout.Node {
	t.Helper()
	data, err := ToJSON(node)
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	out, err := FromJSON(data)
	if err != nil {
		t.Fatalf("FromJSON: %v\n%s", err, data)
	}
	return out
}

func TestLengthRoundTripPreservesUnits(t *testing.T) {
	cases := []layout.Length{
		layout.Px(10), layout.Px(-1), layout.Px(0), layout.Px(1.5), layout.Px(0.0000001),
		layout.Em(2), layout.Rem(1.25), layout.Ch(3), layout.Vh(50), layout.Vw(33.5),
		layout.Vmin(1), layout.Vmax(2), layout.Pt(12), layout.Cm(2.54), layout.Mm(10), layout.In(1), layout.Q(4),
		layout.Cqw(10), layout.Cqh(10), layout.Cqi(10), layout.Cqb(10), layout.Cqmin(10), layout.Cqmax(10),
		layout.Px(layout.SizeMinContent), layout.Px(layout.SizeMaxContent), layout.Px(layout.SizeFitContent),
	}
	for _, want := range cases {
		out := roundTrip(t, &layout.Node{Style: layout.Style{Width: want}})
		if out.Style.Width != want {
			t.Errorf("Width %+v round-tripped as %+v", want, out.Style.Width)
		}
	}
}

func TestLengthZeroValueIsOmittedAndDistinctFromZeroPx(t *testing.T) {
	// The engine treats Length{} (no unit) as "not set" and Px(0) as an
	// explicit zero, so the two must not collapse into each other.
	data, err := ToJSON(&layout.Node{Style: layout.Style{Width: layout.Px(0)}})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"width": "0px"`) {
		t.Errorf("Px(0) should serialize as \"0px\":\n%s", data)
	}
	out, err := FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if out.Style.Width != layout.Px(0) {
		t.Errorf("Px(0) round-tripped as %+v", out.Style.Width)
	}

	data, err = ToJSON(&layout.Node{})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "width") {
		t.Errorf("zero-value Length must be omitted:\n%s", data)
	}
	out, err = FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if out.Style.Width != (layout.Length{}) {
		t.Errorf("omitted width should read back as Length{}, got %+v", out.Style.Width)
	}
}

func TestUnboundedLengthRoundTrip(t *testing.T) {
	for _, in := range []layout.Length{layout.UnboundedLength(), layout.PxUnbounded, layout.Px(math.Inf(1))} {
		data, err := ToJSON(&layout.Node{Style: layout.Style{MaxWidth: in}})
		if err != nil {
			t.Fatalf("ToJSON(%+v): %v", in, err)
		}
		if !strings.Contains(string(data), `"maxWidth": "unbounded"`) {
			t.Errorf("%+v should serialize as \"unbounded\":\n%s", in, data)
		}
		out, err := FromJSON(data)
		if err != nil {
			t.Fatal(err)
		}
		// Both forms normalize to PxUnbounded, which is what the grid engine checks.
		if out.Style.MaxWidth != layout.PxUnbounded {
			t.Errorf("%+v read back as %+v, want PxUnbounded", in, out.Style.MaxWidth)
		}
		if layout.ResolveLength(out.Style.MaxWidth, nil, 16) != layout.Unbounded {
			t.Errorf("unbounded did not resolve to Unbounded")
		}
	}

	// FractionTrack / AutoTrack use PxUnbounded as MaxSize.
	out := roundTrip(t, &layout.Node{Style: layout.Style{
		Display:             layout.DisplayGrid,
		GridTemplateColumns: []layout.GridTrack{layout.FractionTrack(1), layout.AutoTrack(), layout.FitContentTrack(300)},
	}})
	if !reflect.DeepEqual(out.Style.GridTemplateColumns, []layout.GridTrack{layout.FractionTrack(1), layout.AutoTrack(), layout.FitContentTrack(300)}) {
		t.Errorf("tracks round-tripped as %+v", out.Style.GridTemplateColumns)
	}
}

func TestLegacyBareNumberLengthsAreReadAsPx(t *testing.T) {
	legacy := `{
	  "style": {
	    "display": "flex",
	    "width": 200,
	    "height": -1,
	    "padding": {"top": 10, "right": 10, "bottom": 10, "left": 10},
	    "gridTemplateRows": [{"minSize": 100, "maxSize": 100}],
	    "gridAutoRows": {"minSize": 0, "maxSize": 50}
	  },
	  "children": [{"style": {"width": 100.5, "height": 50}}]
	}`
	out, err := FromJSON([]byte(legacy))
	if err != nil {
		t.Fatalf("FromJSON(legacy): %v", err)
	}
	if out.Style.Width != layout.Px(200) || out.Style.Height != layout.Px(-1) {
		t.Errorf("width/height = %+v / %+v", out.Style.Width, out.Style.Height)
	}
	if out.Style.Padding != layout.Uniform(layout.Px(10)) {
		t.Errorf("padding = %+v", out.Style.Padding)
	}
	if len(out.Style.GridTemplateRows) != 1 || out.Style.GridTemplateRows[0] != layout.FixedTrack(layout.Px(100)) {
		t.Errorf("gridTemplateRows = %+v", out.Style.GridTemplateRows)
	}
	if out.Style.GridAutoRows.MaxSize != layout.Px(50) {
		t.Errorf("gridAutoRows = %+v", out.Style.GridAutoRows)
	}
	if out.Children[0].Style.Width != layout.Px(100.5) {
		t.Errorf("child width = %+v", out.Children[0].Style.Width)
	}
}

func TestLegacyBareNumberLengthsAreReadAsPxYAML(t *testing.T) {
	legacy := "style:\n  display: flex\n  width: 200\n  height: 1.5\n  padding:\n    top: 10\n"
	out, err := FromYAML([]byte(legacy))
	if err != nil {
		t.Fatalf("FromYAML(legacy): %v", err)
	}
	if out.Style.Width != layout.Px(200) || out.Style.Height != layout.Px(1.5) || out.Style.Padding.Top != layout.Px(10) {
		t.Errorf("got width %+v height %+v padding.top %+v", out.Style.Width, out.Style.Height, out.Style.Padding.Top)
	}
}

func TestDisplayNoneAndInlineTextRoundTrip(t *testing.T) {
	for _, d := range []layout.Display{layout.DisplayBlock, layout.DisplayFlex, layout.DisplayGrid, layout.DisplayInlineText, layout.DisplayNone} {
		out := roundTrip(t, &layout.Node{Style: layout.Style{Display: d}})
		if out.Style.Display != d {
			t.Errorf("Display %d round-tripped as %d", d, out.Style.Display)
		}
	}
}

func TestDefaultAlignmentIsOmitted(t *testing.T) {
	data, err := ToJSON(&layout.Node{})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"alignItems", "justifyItems", "padding", "margin", "border", "transform", "gridAutoRows", "gridAutoColumns", "rect"} {
		if strings.Contains(string(data), key) {
			t.Errorf("default node must not emit %q:\n%s", key, data)
		}
	}
	// A default node is a very small document.
	if string(data) != "{\n  \"style\": {}\n}" {
		t.Errorf("unexpected default encoding:\n%s", data)
	}

	// Non-default alignment is still written and read.
	out := roundTrip(t, &layout.Node{Style: layout.Style{AlignItems: layout.AlignItemsCenter, JustifyItems: layout.JustifyItemsEnd}})
	if out.Style.AlignItems != layout.AlignItemsCenter || out.Style.JustifyItems != layout.JustifyItemsEnd {
		t.Errorf("alignment round trip: %+v", out.Style)
	}
	// "stretch" is still accepted explicitly.
	out, err = FromJSON([]byte(`{"style":{"alignItems":"stretch","justifyItems":"stretch"}}`))
	if err != nil || out.Style.AlignItems != layout.AlignItemsStretch || out.Style.JustifyItems != layout.JustifyItemsStretch {
		t.Errorf("explicit stretch: %v %+v", err, out)
	}
}

func TestIdentityTransformIsOmitted(t *testing.T) {
	for _, tr := range []layout.Transform{{}, layout.IdentityTransform()} {
		data, err := ToJSON(&layout.Node{Style: layout.Style{Transform: tr}})
		if err != nil {
			t.Fatal(err)
		}
		if strings.Contains(string(data), "transform") {
			t.Errorf("identity transform %+v must be omitted:\n%s", tr, data)
		}
	}
	out := roundTrip(t, &layout.Node{Style: layout.Style{Transform: layout.RotateDegrees(10)}})
	if out.Style.Transform != layout.RotateDegrees(10) {
		t.Errorf("transform round trip: %+v", out.Style.Transform)
	}
}

// TestFullStyleRoundTrip covers every previously unserialized field.
func TestFullStyleRoundTrip(t *testing.T) {
	areas := layout.NewGridTemplateAreas(3, 3)
	areas.DefineArea("header", 0, 1, 0, 3)
	areas.DefineArea("sidebar", 1, 3, 0, 1)

	in := &layout.Node{
		Text:     "Hello, world",
		Baseline: 12.5,
		Rect:     layout.Rect{X: 1, Y: 2, Width: 3, Height: 4},
		Style: layout.Style{
			Display:        layout.DisplayGrid,
			FlexDirection:  layout.FlexDirectionColumnReverse,
			FlexWrap:       layout.FlexWrapWrapReverse,
			JustifyContent: layout.JustifyContentSpaceEvenly,
			AlignItems:     layout.AlignItemsBaseline,
			AlignContent:   layout.AlignContentSpaceAround,
			AlignSelf:      layout.AlignItemsFlexEnd,
			JustifyItems:   layout.JustifyItemsCenter,
			JustifySelf:    layout.JustifyItemsStart,
			FlexGrow:       1.5,
			FlexShrink:     0.5,
			FlexBasis:      layout.Em(10),
			FlexGap:        layout.Px(4),
			FlexRowGap:     layout.Px(5),
			FlexColumnGap:  layout.Px(6),
			Order:          -3,

			GridTemplateRows:    []layout.GridTrack{layout.FixedTrack(layout.Rem(2)), layout.MinMaxTrack(layout.Px(10), layout.Px(100))},
			GridTemplateColumns: []layout.GridTrack{layout.FractionTrack(2), layout.MinContentTrack(), layout.MaxContentTrack()},
			GridAutoRows:        layout.FixedTrack(layout.Px(40)),
			GridAutoColumns:     layout.AutoTrack(),
			GridAutoFlow:        layout.GridAutoFlowColumnDense,
			GridGap:             layout.Px(8),
			GridRowGap:          layout.Px(9),
			GridColumnGap:       layout.Px(10),
			GridRowStart:        -1,
			GridRowEnd:          2,
			GridColumnStart:     1,
			GridColumnEnd:       3,
			GridTemplateAreas:   areas,
			GridArea:            "header",

			Width:            layout.Vw(50),
			Height:           layout.Px(200),
			MinWidth:         layout.Px(10),
			MinHeight:        layout.Px(20),
			MaxWidth:         layout.PxUnbounded,
			MaxHeight:        layout.Px(400),
			AspectRatio:      16.0 / 9.0,
			WidthSizing:      layout.IntrinsicSizeFitContent,
			HeightSizing:     layout.IntrinsicSizeMaxContent,
			FitContentWidth:  layout.Px(300),
			FitContentHeight: layout.Px(150),

			Padding:   layout.Uniform(layout.Px(10)),
			Margin:    layout.Horizontal(layout.Em(1)),
			Border:    layout.Vertical(layout.Px(2)),
			BoxSizing: layout.BoxSizingBorderBox,

			Position: layout.PositionSticky,
			Top:      layout.Px(1),
			Right:    layout.Px(2),
			Bottom:   layout.Px(3),
			Left:     layout.Px(4),
			ZIndex:   7,

			Transform: layout.Translate(3, 4).Multiply(layout.Scale(2, 2)),

			WritingMode:   layout.WritingModeSidewaysLR,
			ContainerType: layout.ContainerTypeInlineSize,
			ContainerName: layout.ContainerName{"card", "sidebar"},

			TextStyle: &layout.TextStyle{
				TextAlign:           layout.TextAlignJustify,
				TextAlignLast:       layout.TextAlignLastCenter,
				TextJustify:         layout.TextJustifyDistribute,
				LineHeight:          1.4,
				WordSpacing:         -1,
				LetterSpacing:       0.5,
				TextIndent:          -8,
				WhiteSpace:          layout.WhiteSpacePreLine,
				OverflowWrap:        layout.OverflowWrapAnywhere,
				WordBreak:           layout.WordBreakKeepAll,
				TextOverflow:        layout.TextOverflowEllipsis,
				TextTransform:       layout.TextTransformFullSizeKana,
				Hyphens:             layout.HyphensAuto,
				HangingPunctuation:  layout.HangingPunctuationAllowEnd,
				TabSize:             4,
				FontSize:            14,
				FontFamily:          "Inter",
				FontWeight:          layout.FontWeightBold,
				FontStyle:           layout.FontStyleOblique,
				TextDecoration:      layout.TextDecorationUnderline | layout.TextDecorationLineThrough,
				TextDecorationStyle: layout.TextDecorationStyleWavy,
				TextDecorationColor: "#ff0000",
				VerticalAlign:       layout.VerticalAlignSuper,
				WritingMode:         layout.WritingModeVerticalRL,
				Direction:           layout.DirectionRTL,
			},
		},
		Children: []*layout.Node{
			{Text: "child", Style: layout.Style{Display: layout.DisplayInlineText, Width: layout.Ch(20)}},
		},
	}

	out := roundTrip(t, in)
	if !reflect.DeepEqual(in, out) {
		data, _ := ToJSON(in)
		t.Fatalf("round trip mismatch\n got: %+v\nwant: %+v\njson:\n%s", out, in, data)
	}

	// TextLayout is derived output and intentionally not serialized.
	in.TextLayout = &layout.TextLayout{LineHeight: 10}
	out = roundTrip(t, in)
	if out.TextLayout != nil {
		t.Errorf("TextLayout must not be serialized")
	}
}

func TestYAMLRoundTrip(t *testing.T) {
	in := &layout.Node{
		Text: "yaml",
		Style: layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Em(12.5),
			MaxWidth:      layout.PxUnbounded,
			Padding:       layout.Uniform(layout.Px(10)),
			AlignItems:    layout.AlignItemsCenter,
			TextStyle:     &layout.TextStyle{FontSize: 14, TextAlign: layout.TextAlignCenter},
		},
		Children: []*layout.Node{{Style: layout.Style{Height: layout.Px(50)}}},
	}
	data, err := ToYAML(in)
	if err != nil {
		t.Fatalf("ToYAML: %v", err)
	}
	// YAML keys match the JSON keys and omit defaults.
	if !strings.Contains(string(data), "flexDirection: column") || strings.Contains(string(data), "margin") {
		t.Errorf("unexpected YAML:\n%s", data)
	}
	out, err := FromYAML(data)
	if err != nil {
		t.Fatalf("FromYAML: %v\n%s", err, data)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("YAML round trip mismatch\n got: %+v\nwant: %+v\nyaml:\n%s", out, in, data)
	}
}

// --- Validation -------------------------------------------------------------

func TestFromJSONRejectsBadLengths(t *testing.T) {
	cases := map[string]string{
		"unknown unit":   `{"style":{"width":"10parsecs"}}`,
		"garbage":        `{"style":{"width":"abc"}}`,
		"exponent":       `{"style":{"width":"1e3px"}}`,
		"too large":      `{"style":{"width":"1e13"}}`,
		"too large num":  `{"style":{"width":1e13}}`,
		"too small":      `{"style":{"width":"-1000000000000000px"}}`,
		"nested padding": `{"style":{"padding":{"top":"NaNpx"}}}`,
		"track":          `{"style":{"gridTemplateRows":[{"minSize":"1e300px"}]}}`,
		"wrong type":     `{"style":{"width":true}}`,
		"object":         `{"style":{"width":{"value":10}}}`,
		"child":          `{"style":{},"children":[{"style":{"height":"10furlongs"}}]}`,
	}
	for name, input := range cases {
		if _, err := FromJSON([]byte(input)); err == nil {
			t.Errorf("%s: expected error for %s", name, input)
		}
	}
}

func TestFromJSONRejectsNonFiniteAndAbsurdFloats(t *testing.T) {
	// encoding/json already rejects NaN/Infinity literals; magnitude is ours.
	cases := []string{
		`{"style":{"flexGrow":1e13}}`,
		`{"style":{"aspectRatio":-1e13}}`,
		`{"style":{"transform":{"a":1,"b":0,"c":0,"d":1,"e":1e300,"f":0}}}`,
		`{"style":{},"rect":{"x":0,"y":0,"width":1e15,"height":0}}`,
		`{"style":{},"baseline":1e13}`,
		`{"style":{"textStyle":{"fontSize":1e13}}}`,
		`{"style":{"gridTemplateColumns":[{"fraction":1e13}]}}`,
	}
	for _, input := range cases {
		if _, err := FromJSON([]byte(input)); err == nil {
			t.Errorf("expected error for %s", input)
		}
	}
	// YAML can express non-finite floats directly.
	for _, input := range []string{"style:\n  flexGrow: .nan\n", "style:\n  aspectRatio: .inf\n", "style:\n  width: .inf\n"} {
		if _, err := FromYAML([]byte(input)); err == nil {
			t.Errorf("expected error for YAML %q", input)
		}
	}
}

func TestFromJSONRejectsUnknownEnums(t *testing.T) {
	cases := []string{
		`{"style":{"display":"table"}}`,
		`{"style":{"flexDirection":"diagonal"}}`,
		`{"style":{"flexWrap":"maybe"}}`,
		`{"style":{"justifyContent":"left"}}`,
		`{"style":{"alignItems":"start"}}`,
		`{"style":{"alignSelf":"middle"}}`,
		`{"style":{"alignContent":"space-evenly"}}`,
		`{"style":{"justifyItems":"flex-start"}}`,
		`{"style":{"justifySelf":"baseline"}}`,
		`{"style":{"gridAutoFlow":"dense"}}`,
		`{"style":{"boxSizing":"padding-box"}}`,
		`{"style":{"position":"floating"}}`,
		`{"style":{"widthSizing":"auto"}}`,
		`{"style":{"writingMode":"diagonal"}}`,
		`{"style":{"containerType":"block-size"}}`,
		`{"style":{"textStyle":{"textAlign":"middle"}}}`,
		`{"style":{"textStyle":{"whiteSpace":"collapse"}}}`,
		`{"style":{"textStyle":{"direction":"ttb"}}}`,
		`{"style":{"textStyle":{"fontWeight":1001}}}`,
		`{"style":{"textStyle":{"fontWeight":-1}}}`,
		`{"style":{"textStyle":{"textDecoration":8}}}`,
		`{"style":{"display":"Flex"}}`, // case-sensitive
	}
	for _, input := range cases {
		if _, err := FromJSON([]byte(input)); err == nil {
			t.Errorf("expected error for %s", input)
		}
	}
	// Errors name the offending field.
	_, err := FromJSON([]byte(`{"style":{},"children":[{"style":{"display":"table"}}]}`))
	if err == nil || !strings.Contains(err.Error(), "children[0].style.display") || !strings.Contains(err.Error(), `"table"`) {
		t.Errorf("error should name the field: %v", err)
	}
}

func TestFromJSONRejectsFractionalOrOverflowingInts(t *testing.T) {
	for _, input := range []string{
		`{"style":{"gridRowStart":1.5}}`,
		`{"style":{"zIndex":1e3}}`,
		`{"style":{"order":99999999999999999999}}`,
		`{"style":{"gridTemplateAreas":{"rows":-1,"cols":1}}}`,
		`{"style":{"gridTemplateAreas":{"rows":1,"cols":1,"areas":[{"name":"a","rowStart":1,"rowEnd":0,"columnStart":0,"columnEnd":1}]}}}`,
	} {
		if _, err := FromJSON([]byte(input)); err == nil {
			t.Errorf("expected error for %s", input)
		}
	}
}

func TestFromJSONEnforcesDepthLimit(t *testing.T) {
	deep := strings.Repeat(`{"style":{},"children":[`, MaxTreeDepth+2) + `{"style":{}}` + strings.Repeat(`]}`, MaxTreeDepth+2)
	_, err := FromJSON([]byte(deep))
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}

	// Exactly at the limit is fine.
	ok := strings.Repeat(`{"style":{},"children":[`, MaxTreeDepth) + `{"style":{}}` + strings.Repeat(`]}`, MaxTreeDepth)
	if _, err := FromJSON([]byte(ok)); err != nil {
		t.Fatalf("depth %d should be accepted: %v", MaxTreeDepth, err)
	}

	// ToJSON applies the same cap so a cyclic tree fails instead of overflowing.
	cyc := &layout.Node{}
	cyc.Children = []*layout.Node{cyc}
	if _, err := ToJSON(cyc); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded for cyclic tree, got %v", err)
	}
}

func TestFromJSONEnforcesChildLimit(t *testing.T) {
	var b strings.Builder
	b.WriteString(`{"style":{},"children":[`)
	for i := 0; i <= MaxChildren; i++ {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(`{"style":{}}`)
	}
	b.WriteString(`]}`)
	_, err := FromJSON([]byte(b.String()))
	if !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}
}

func TestToJSONRejectsNonFiniteValues(t *testing.T) {
	for _, n := range []*layout.Node{
		{Style: layout.Style{Width: layout.Px(math.NaN())}},
		{Style: layout.Style{Width: layout.Px(math.Inf(-1))}},
		{Style: layout.Style{Width: layout.Px(1e13)}},
		{Style: layout.Style{FlexGrow: math.NaN()}},
		{Style: layout.Style{Display: layout.Display(42)}},
		{Baseline: math.Inf(1)},
	} {
		if _, err := ToJSON(n); err == nil {
			t.Errorf("expected ToJSON error for %+v", n)
		}
	}
}

func TestLengthJSONUnmarshal(t *testing.T) {
	var got struct {
		A, B, C, D LengthJSON
	}
	if err := json.Unmarshal([]byte(`{"A":"10px","B":10,"C":null,"D":-1.5}`), &got); err != nil {
		t.Fatal(err)
	}
	if got.A != "10px" || got.B != "10px" || got.C != "" || got.D != "-1.5px" {
		t.Errorf("got %+v", got)
	}
	for _, bad := range []string{`{"A":true}`, `{"A":[1]}`, `{"A":{}}`} {
		if err := json.Unmarshal([]byte(bad), &got); err == nil {
			t.Errorf("expected error for %s", bad)
		}
	}
}
