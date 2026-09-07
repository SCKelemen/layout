package serialize

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/SCKelemen/layout"
)

// A bare null document has no layout meaning and must not decode to an empty
// node with a nil error.
func TestFromJSONRejectsNull(t *testing.T) {
	for _, input := range []string{"null", " null ", "\nnull\n"} {
		n, err := FromJSON([]byte(input))
		if !errors.Is(err, ErrNullInput) {
			t.Errorf("FromJSON(%q) = %v, %v; want ErrNullInput", input, n, err)
		}
	}
	// A null child is still skipped (documented behavior).
	n, err := FromJSON([]byte(`{"style":{},"children":[null,{"style":{}}]}`))
	if err != nil || len(n.Children) != 1 {
		t.Fatalf("null child should be skipped: %v, %v", n, err)
	}
}

func TestFromYAMLRejectsNull(t *testing.T) {
	skipIfNoYAML(t)
	for _, input := range []string{"", "null", "~", "---\n", "# only a comment\n"} {
		n, err := yamlDecode([]byte(input))
		if !errors.Is(err, ErrNullInput) {
			t.Errorf("FromYAML(%q) = %v, %v; want ErrNullInput", input, n, err)
		}
	}
}

// ToJSON applies the MaxChildren cap that FromJSON enforces so an encoded
// tree always reads back.
func TestToJSONEnforcesChildLimit(t *testing.T) {
	root := &layout.Node{Children: make([]*layout.Node, MaxChildren+1)}
	for i := range root.Children {
		root.Children[i] = &layout.Node{}
	}
	if _, err := ToJSON(root); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}
	// Exactly at the limit is fine.
	root.Children = root.Children[:MaxChildren]
	if _, err := ToJSON(root); err != nil {
		t.Fatalf("%d children should be accepted: %v", MaxChildren, err)
	}

	// Track lists are capped the same way on encode.
	wide := &layout.Node{Style: layout.Style{GridTemplateColumns: make([]layout.GridTrack, MaxChildren+1)}}
	if _, err := ToJSON(wide); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded for %d tracks, got %v", MaxChildren+1, err)
	}
}

func TestToYAMLEnforcesChildLimit(t *testing.T) {
	skipIfNoYAML(t)
	root := &layout.Node{Children: make([]*layout.Node, MaxChildren+1)}
	for i := range root.Children {
		root.Children[i] = &layout.Node{}
	}
	if _, err := yamlEncode(root); !errors.Is(err, ErrLimitExceeded) {
		t.Fatalf("expected ErrLimitExceeded, got %v", err)
	}
}

func TestDirectionRoundTrip(t *testing.T) {
	n := &layout.Node{Style: layout.Style{Direction: layout.DirectionRTL}}
	data, err := ToJSON(n)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"direction": "rtl"`) {
		t.Fatalf("direction not encoded: %s", data)
	}
	back, err := FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if back.Style.Direction != layout.DirectionRTL {
		t.Fatalf("Direction = %v, want rtl", back.Style.Direction)
	}
	// ltr is the zero value and is omitted.
	data, _ = ToJSON(&layout.Node{})
	if strings.Contains(string(data), "direction") {
		t.Fatalf("default direction should be omitted: %s", data)
	}
}

func repeatFixture() *layout.Node {
	return &layout.Node{
		Style: layout.Style{
			Display:             layout.DisplayGrid,
			GridTemplateColumns: []layout.GridTrack{layout.FixedTrack(layout.Px(200))},
			GridTemplateColumnsRepeat: []layout.RepeatTrack{
				layout.AutoFillTracks(layout.FixedTrack(layout.Px(100))),
				{Count: 3, Tracks: []layout.GridTrack{layout.FractionTrack(1), layout.MinMaxTrack(layout.Px(50), layout.Em(10))}},
			},
			GridTemplateRowsRepeat: []layout.RepeatTrack{
				layout.AutoFitTracks(layout.FixedTrack(layout.Px(40))),
			},
		},
	}
}

func TestRepeatTrackRoundTripJSON(t *testing.T) {
	n := repeatFixture()
	data, err := ToJSON(n)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	for _, want := range []string{`"count": "auto-fill"`, `"count": 3`, `"count": "auto-fit"`, `"gridTemplateColumnsRepeat"`, `"gridTemplateRowsRepeat"`} {
		if !strings.Contains(s, want) {
			t.Errorf("missing %s in\n%s", want, s)
		}
	}
	back, err := FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back.Style.GridTemplateColumnsRepeat, n.Style.GridTemplateColumnsRepeat) {
		t.Errorf("columns repeat mismatch:\n got %+v\nwant %+v", back.Style.GridTemplateColumnsRepeat, n.Style.GridTemplateColumnsRepeat)
	}
	if !reflect.DeepEqual(back.Style.GridTemplateRowsRepeat, n.Style.GridTemplateRowsRepeat) {
		t.Errorf("rows repeat mismatch:\n got %+v\nwant %+v", back.Style.GridTemplateRowsRepeat, n.Style.GridTemplateRowsRepeat)
	}
	// Absent repeats stay nil (not an empty slice) so DeepEqual with a
	// hand-built Style keeps working.
	plain, _ := FromJSON([]byte(`{"style":{"display":"grid"}}`))
	if plain.Style.GridTemplateColumnsRepeat != nil || plain.Style.GridTemplateRowsRepeat != nil {
		t.Errorf("absent repeats should decode to nil")
	}
}

func TestRepeatTrackRoundTripYAML(t *testing.T) {
	skipIfNoYAML(t)
	n := repeatFixture()
	data, err := yamlEncode(n)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	// YAML writes the integer count as a bare number and keywords as strings.
	if !strings.Contains(s, "count: 3\n") || !strings.Contains(s, "count: auto-fill") {
		t.Fatalf("unexpected YAML:\n%s", s)
	}
	back, err := yamlDecode(data)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back.Style.GridTemplateColumnsRepeat, n.Style.GridTemplateColumnsRepeat) {
		t.Errorf("columns repeat mismatch:\n got %+v\nwant %+v", back.Style.GridTemplateColumnsRepeat, n.Style.GridTemplateColumnsRepeat)
	}

	// Hand-written YAML with a quoted number and a keyword.
	hand := "style:\n  gridTemplateColumnsRepeat:\n    - count: 2\n      tracks:\n        - minSize: 10px\n          maxSize: 10px\n    - count: auto-fit\n      tracks:\n        - minSize: 1em\n          maxSize: 1em\n"
	back, err = yamlDecode([]byte(hand))
	if err != nil {
		t.Fatal(err)
	}
	got := back.Style.GridTemplateColumnsRepeat
	if len(got) != 2 || got[0].Count != 2 || got[1].Count != layout.RepeatCountAutoFit || got[1].Tracks[0].MinSize != layout.Em(1) {
		t.Errorf("hand-written YAML decoded to %+v", got)
	}
}

func TestRepeatCountValidation(t *testing.T) {
	bad := map[string]string{
		"zero":       `{"style":{"gridTemplateColumnsRepeat":[{"count":0,"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"negative":   `{"style":{"gridTemplateColumnsRepeat":[{"count":-1,"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"too large":  `{"style":{"gridTemplateColumnsRepeat":[{"count":10001,"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"fractional": `{"style":{"gridTemplateColumnsRepeat":[{"count":1.5,"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"keyword":    `{"style":{"gridTemplateColumnsRepeat":[{"count":"auto","tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"missing":    `{"style":{"gridTemplateColumnsRepeat":[{"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"null":       `{"style":{"gridTemplateColumnsRepeat":[{"count":null,"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"object":     `{"style":{"gridTemplateColumnsRepeat":[{"count":{"n":1},"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`,
		"no tracks":  `{"style":{"gridTemplateRowsRepeat":[{"count":2}]}}`,
		"bad length": `{"style":{"gridTemplateRowsRepeat":[{"count":2,"tracks":[{"minSize":"1parsec"}]}]}}`,
	}
	for name, input := range bad {
		_, err := FromJSON([]byte(input))
		if err == nil {
			t.Errorf("%s: expected error for %s", name, input)
			continue
		}
		if !strings.Contains(err.Error(), "Repeat[0]") {
			t.Errorf("%s: error should name the repeat entry: %v", name, err)
		}
	}

	// Boundaries.
	for _, count := range []string{"1", "10000", `"auto-fill"`, `"auto-fit"`} {
		input := `{"style":{"gridTemplateColumnsRepeat":[{"count":` + count + `,"tracks":[{"minSize":"1px","maxSize":"1px"}]}]}}`
		if _, err := FromJSON([]byte(input)); err != nil {
			t.Errorf("count %s should be accepted: %v", count, err)
		}
	}

	// Encoding rejects counts the decoder would refuse.
	for _, count := range []int{0, -3, MaxRepeatCount + 1} {
		n := &layout.Node{Style: layout.Style{GridTemplateRowsRepeat: []layout.RepeatTrack{{Count: count, Tracks: []layout.GridTrack{layout.AutoTrack()}}}}}
		if _, err := ToJSON(n); err == nil || !strings.Contains(err.Error(), "gridTemplateRowsRepeat[0].count") {
			t.Errorf("ToJSON with count %d: %v", count, err)
		}
	}
	empty := &layout.Node{Style: layout.Style{GridTemplateRowsRepeat: []layout.RepeatTrack{layout.AutoFillTracks()}}}
	if _, err := ToJSON(empty); err == nil {
		t.Errorf("ToJSON with an empty repeat pattern should fail")
	}
}

// css-align-3 start/end keywords decode to the flex-* values and the new
// keywords added in v1.4.0 (space-evenly for align-content, stretch for
// justify-content) round-trip.
func TestAlignmentAliasesAndNewKeywords(t *testing.T) {
	n, err := FromJSON([]byte(`{"style":{"justifyContent":"start","alignItems":"end","alignSelf":"start","alignContent":"end"}}`))
	if err != nil {
		t.Fatal(err)
	}
	if n.Style.JustifyContent != layout.JustifyContentFlexStart || n.Style.AlignItems != layout.AlignItemsFlexEnd ||
		n.Style.AlignSelf != layout.AlignItemsFlexStart || n.Style.AlignContent != layout.AlignContentFlexEnd {
		t.Errorf("aliases decoded to %+v", n.Style)
	}
	// Re-encoding writes the canonical keyword. (justifyContent flex-start is
	// the zero value and is omitted.)
	data, _ := ToJSON(n)
	if !strings.Contains(string(data), `"alignSelf": "flex-start"`) || !strings.Contains(string(data), `"alignContent": "flex-end"`) || strings.Contains(string(data), "justifyContent") {
		t.Errorf("aliases should re-encode as flex-*: %s", data)
	}

	full := &layout.Node{Style: layout.Style{JustifyContent: layout.JustifyContentStretch, AlignContent: layout.AlignContentSpaceEvenly}}
	data, err = ToJSON(full)
	if err != nil {
		t.Fatal(err)
	}
	back, err := FromJSON(data)
	if err != nil {
		t.Fatal(err)
	}
	if back.Style.JustifyContent != layout.JustifyContentStretch || back.Style.AlignContent != layout.AlignContentSpaceEvenly {
		t.Errorf("new keywords did not round-trip: %s", data)
	}
}

// An enum value outside its range cannot be represented and must fail on
// encode with a path to the field.
func TestToJSONRejectsUnknownEnumValues(t *testing.T) {
	cases := map[string]*layout.Node{
		"root.style.display":              {Style: layout.Style{Display: layout.Display(42)}},
		"root.style.alignSelf":            {Style: layout.Style{AlignSelf: layout.AlignItems(9)}},
		"root.style.direction":            {Style: layout.Style{Direction: layout.Direction(5)}},
		"root.style.textStyle.whiteSpace": {Style: layout.Style{TextStyle: &layout.TextStyle{WhiteSpace: layout.WhiteSpace(7)}}},
		"root.style.containerType":        {Style: layout.Style{ContainerType: layout.ContainerType(7)}},
	}
	for path, n := range cases {
		_, err := ToJSON(n)
		if err == nil || !strings.Contains(err.Error(), path) {
			t.Errorf("%s: err = %v", path, err)
		}
	}
}
