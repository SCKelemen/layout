package layout

import (
	"fmt"
	"strings"
	"testing"
)

// enumSpec drives the round-trip test for one enum: it knows how many values
// the enum has, formats a value, and parses a keyword back to an int.
type enumSpec struct {
	name   string
	count  int
	str    func(int) string
	parse  func(string) (int, error)
	unkStr string // expected String() of value 99
}

func spec[T ~int](name string, count int, parse func(string) (T, error), str func(T) string) enumSpec {
	return enumSpec{
		name:  name,
		count: count,
		str:   func(i int) string { return str(T(i)) },
		parse: func(s string) (int, error) {
			v, err := parse(s)
			return int(v), err
		},
		unkStr: name + "(99)",
	}
}

var enumSpecs = []enumSpec{
	spec("Display", len(displayNames), ParseDisplay, Display.String),
	spec("FlexDirection", len(flexDirectionNames), ParseFlexDirection, FlexDirection.String),
	spec("FlexWrap", len(flexWrapNames), ParseFlexWrap, FlexWrap.String),
	spec("JustifyContent", len(justifyContentNames), ParseJustifyContent, JustifyContent.String),
	spec("AlignItems", len(alignItemsNames), ParseAlignItems, AlignItems.String),
	spec("JustifyItems", len(justifyItemsNames), ParseJustifyItems, JustifyItems.String),
	spec("AlignContent", len(alignContentNames), ParseAlignContent, AlignContent.String),
	spec("GridAutoFlow", len(gridAutoFlowNames), ParseGridAutoFlow, GridAutoFlow.String),
	spec("BoxSizing", len(boxSizingNames), ParseBoxSizing, BoxSizing.String),
	spec("Position", len(positionNames), ParsePosition, Position.String),
	spec("TextAlign", len(textAlignNames), ParseTextAlign, TextAlign.String),
	spec("TextAlignLast", len(textAlignLastNames), ParseTextAlignLast, TextAlignLast.String),
	spec("TextJustify", len(textJustifyNames), ParseTextJustify, TextJustify.String),
	spec("WhiteSpace", len(whiteSpaceNames), ParseWhiteSpace, WhiteSpace.String),
	spec("TextOverflow", len(textOverflowNames), ParseTextOverflow, TextOverflow.String),
	spec("OverflowWrap", len(overflowWrapNames), ParseOverflowWrap, OverflowWrap.String),
	spec("WordBreak", len(wordBreakNames), ParseWordBreak, WordBreak.String),
	spec("TextTransform", len(textTransformNames), ParseTextTransform, TextTransform.String),
	spec("Hyphens", len(hyphensNames), ParseHyphens, Hyphens.String),
	spec("HangingPunctuation", len(hangingPunctuationNames), ParseHangingPunctuation, HangingPunctuation.String),
	spec("Direction", len(directionNames), ParseDirection, Direction.String),
	spec("WritingMode", len(writingModeNames), ParseWritingMode, WritingMode.String),
	spec("FontStyle", len(fontStyleNames), ParseFontStyle, FontStyle.String),
	spec("TextDecorationStyle", len(textDecorationStyleNames), ParseTextDecorationStyle, TextDecorationStyle.String),
	spec("VerticalAlign", len(verticalAlignNames), ParseVerticalAlign, VerticalAlign.String),
	spec("IntrinsicSize", len(intrinsicSizeNames), ParseIntrinsicSize, IntrinsicSize.String),
	spec("InlineBoxKind", len(inlineBoxKindNames), ParseInlineBoxKind, InlineBoxKind.String),
}

// TestEnumRoundTrip checks Parse(String(v)) == v for every value of every
// enum, that String never returns an empty or Go-style name for a valid
// value, and that out-of-range values format as Kind(n) and fail to parse
// with an error listing the valid keywords.
func TestEnumRoundTrip(t *testing.T) {
	for _, e := range enumSpecs {
		t.Run(e.name, func(t *testing.T) {
			if e.count == 0 {
				t.Fatal("enum has no values")
			}
			for v := 0; v < e.count; v++ {
				s := e.str(v)
				if s == "" || strings.Contains(s, "(") || s != strings.ToLower(s) {
					t.Errorf("%s(%d).String() = %q, want a lowercase CSS keyword", e.name, v, s)
				}
				got, err := e.parse(s)
				if err != nil {
					t.Errorf("Parse%s(%q): %v", e.name, s, err)
					continue
				}
				if got != v {
					t.Errorf("Parse%s(%q) = %d, want %d", e.name, s, got, v)
				}
			}

			if s := e.str(99); s != e.unkStr {
				t.Errorf("%s(99).String() = %q, want %q", e.name, s, e.unkStr)
			}
			if s := e.str(-1); s != e.name+"(-1)" {
				t.Errorf("%s(-1).String() = %q", e.name, s)
			}

			for _, bad := range []string{"", "bogus", e.unkStr, strings.ToUpper(e.str(0))} {
				_, err := e.parse(bad)
				if err == nil {
					t.Errorf("Parse%s(%q) should fail", e.name, bad)
					continue
				}
				// The error lists every canonical keyword.
				for v := 0; v < e.count; v++ {
					if !strings.Contains(err.Error(), e.str(v)) {
						t.Errorf("Parse%s(%q) error %q does not list %q", e.name, bad, err, e.str(v))
					}
				}
			}
		})
	}
}

// TestEnumSpecsCoverAllValues pins the value counts so a new constant added
// to types.go without a keyword is caught here.
func TestEnumSpecsCoverAllValues(t *testing.T) {
	checks := []struct {
		name string
		got  int
		last int
	}{
		{"Display", len(displayNames), int(DisplayNone)},
		{"FlexDirection", len(flexDirectionNames), int(FlexDirectionColumnReverse)},
		{"FlexWrap", len(flexWrapNames), int(FlexWrapWrapReverse)},
		{"JustifyContent", len(justifyContentNames), int(JustifyContentStretch)},
		{"AlignItems", len(alignItemsNames), int(AlignItemsBaseline)},
		{"JustifyItems", len(justifyItemsNames), int(JustifyItemsCenter)},
		{"AlignContent", len(alignContentNames), int(AlignContentSpaceEvenly)},
		{"GridAutoFlow", len(gridAutoFlowNames), int(GridAutoFlowColumnDense)},
		{"BoxSizing", len(boxSizingNames), int(BoxSizingBorderBox)},
		{"Position", len(positionNames), int(PositionSticky)},
		{"TextAlign", len(textAlignNames), int(TextAlignJustify)},
		{"TextAlignLast", len(textAlignLastNames), int(TextAlignLastJustify)},
		{"TextJustify", len(textJustifyNames), int(TextJustifyNone)},
		{"WhiteSpace", len(whiteSpaceNames), int(WhiteSpacePreLine)},
		{"TextOverflow", len(textOverflowNames), int(TextOverflowEllipsis)},
		{"OverflowWrap", len(overflowWrapNames), int(OverflowWrapAnywhere)},
		{"WordBreak", len(wordBreakNames), int(WordBreakKeepAll)},
		{"TextTransform", len(textTransformNames), int(TextTransformFullSizeKana)},
		{"Hyphens", len(hyphensNames), int(HyphensAuto)},
		{"HangingPunctuation", len(hangingPunctuationNames), int(HangingPunctuationAllowEnd)},
		{"Direction", len(directionNames), int(DirectionRTL)},
		{"WritingMode", len(writingModeNames), int(WritingModeSidewaysLR)},
		{"FontStyle", len(fontStyleNames), int(FontStyleOblique)},
		{"TextDecorationStyle", len(textDecorationStyleNames), int(TextDecorationStyleWavy)},
		{"VerticalAlign", len(verticalAlignNames), int(VerticalAlignBottom)},
		{"IntrinsicSize", len(intrinsicSizeNames), int(IntrinsicSizeFitContent)},
		{"InlineBoxKind", len(inlineBoxKindNames), int(InlineBoxText)},
	}
	for _, c := range checks {
		if c.got != c.last+1 {
			t.Errorf("%s: %d keywords for %d values", c.name, c.got, c.last+1)
		}
	}
}

// TestEnumKeywords spot-checks the canonical spellings the serialize wire
// format depends on.
func TestEnumKeywords(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{DisplayInlineText.String(), "inline-text"},
		{DisplayNone.String(), "none"},
		{FlexDirectionRowReverse.String(), "row-reverse"},
		{FlexWrapNoWrap.String(), "nowrap"},
		{JustifyContentSpaceBetween.String(), "space-between"},
		{JustifyContentSpaceEvenly.String(), "space-evenly"},
		{JustifyContentStretch.String(), "stretch"},
		{AlignContentSpaceEvenly.String(), "space-evenly"},
		{AlignItemsBaseline.String(), "baseline"},
		{GridAutoFlowRowDense.String(), "row-dense"},
		{BoxSizingBorderBox.String(), "border-box"},
		{TextAlignDefault.String(), "start"},
		{TextJustifyInterCharacter.String(), "inter-character"},
		{WhiteSpacePreWrap.String(), "pre-wrap"},
		{TextTransformFullSizeKana.String(), "full-size-kana"},
		{HangingPunctuationForceEnd.String(), "force-end"},
		{WritingModeHorizontalTB.String(), "horizontal-tb"},
		{VerticalAlignTextBottom.String(), "text-bottom"},
		{IntrinsicSizeMinContent.String(), "min-content"},
		{IntrinsicSizeNone.String(), "none"},
		{InlineBoxText.String(), "text"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q, want %q", c.got, c.want)
		}
	}
}

// TestEnumAliases: css-align-3 start/end parse to the flex-* values and the
// alias constants format as the canonical flex-* keyword.
func TestEnumAliases(t *testing.T) {
	if v, err := ParseJustifyContent("start"); err != nil || v != JustifyContentFlexStart {
		t.Errorf("ParseJustifyContent(start) = %v, %v", v, err)
	}
	if v, err := ParseJustifyContent("end"); err != nil || v != JustifyContentFlexEnd {
		t.Errorf("ParseJustifyContent(end) = %v, %v", v, err)
	}
	if v, err := ParseAlignItems("start"); err != nil || v != AlignItemsFlexStart {
		t.Errorf("ParseAlignItems(start) = %v, %v", v, err)
	}
	if v, err := ParseAlignItems("end"); err != nil || v != AlignItemsFlexEnd {
		t.Errorf("ParseAlignItems(end) = %v, %v", v, err)
	}
	if v, err := ParseAlignContent("start"); err != nil || v != AlignContentFlexStart {
		t.Errorf("ParseAlignContent(start) = %v, %v", v, err)
	}
	if v, err := ParseAlignContent("end"); err != nil || v != AlignContentFlexEnd {
		t.Errorf("ParseAlignContent(end) = %v, %v", v, err)
	}
	if JustifyContentStart.String() != "flex-start" || AlignItemsEnd.String() != "flex-end" || AlignContentStart.String() != "flex-start" {
		t.Errorf("alias constants must format as the canonical flex-* keyword")
	}

	// The error message lists aliases too.
	_, err := ParseAlignItems("middle")
	if err == nil || !strings.Contains(err.Error(), "start") || !strings.Contains(err.Error(), `"middle"`) {
		t.Errorf("ParseAlignItems(middle) error = %v", err)
	}

	// CSS two-keyword grid-auto-flow spellings.
	for in, want := range map[string]GridAutoFlow{
		"row dense": GridAutoFlowRowDense, "column dense": GridAutoFlowColumnDense,
	} {
		if v, err := ParseGridAutoFlow(in); err != nil || v != want {
			t.Errorf("ParseGridAutoFlow(%q) = %v, %v; want %v", in, v, err, want)
		}
	}
	if _, err := ParseGridAutoFlow("dense"); err == nil {
		t.Errorf("bare \"dense\" should be rejected")
	}
}

// TestEnumStringViaFmt makes sure the Stringer is picked up by %v.
func TestEnumStringViaFmt(t *testing.T) {
	if s := strings.TrimSpace(fmt.Sprintf("%v", DisplayFlex)); s != "flex" {
		t.Errorf("%%v of DisplayFlex = %q", s)
	}
}
