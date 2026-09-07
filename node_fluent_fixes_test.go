package layout

import "testing"

// Regression tests for Node.Transform / Node.Map discarding child-list edits
// and panicking on nil results, and for CloneDeep sharing Style internals.

func TestTransformHonorsChildListEdits(t *testing.T) {
	root := HStack(Fixed(1, 1), Fixed(2, 2), Fixed(3, 3))
	isFlex := func(n *Node) bool { return n.Style.Display == DisplayFlex }

	removed := root.Transform(isFlex, func(n *Node) *Node { return n.RemoveChildAt(0) })
	if len(removed.Children) != 2 {
		t.Fatalf("Transform(RemoveChildAt(0)): got %d children, want 2", len(removed.Children))
	}
	if removed.Children[0].Style.Width.Value != 2 || removed.Children[1].Style.Width.Value != 3 {
		t.Fatalf("Transform(RemoveChildAt(0)): unexpected children %v %v",
			removed.Children[0].Style.Width, removed.Children[1].Style.Width)
	}

	added := root.Transform(isFlex, func(n *Node) *Node { return n.AddChild(Fixed(4, 4)) })
	if len(added.Children) != 4 || added.Children[3].Style.Width.Value != 4 {
		t.Fatalf("Transform(AddChild): got %d children, want 4 with last width 4", len(added.Children))
	}

	// The original tree is untouched in both cases.
	if len(root.Children) != 3 {
		t.Fatalf("original mutated: %d children", len(root.Children))
	}
}

func TestTransformAppliesToAddedChildren(t *testing.T) {
	// A child added by the callback is part of the returned node and is
	// therefore visited by the recursion.
	root := HStack(Fixed(1, 1))
	out := root.Transform(
		func(n *Node) bool { return true },
		func(n *Node) *Node {
			if n.Style.Display == DisplayFlex {
				return n.AddChild(Fixed(7, 7))
			}
			return n.WithPadding(1)
		},
	)
	if len(out.Children) != 2 {
		t.Fatalf("got %d children, want 2", len(out.Children))
	}
	for i, c := range out.Children {
		if c.Style.Padding.Top.Value != 1 {
			t.Errorf("child %d not visited by transform: padding %v", i, c.Style.Padding.Top)
		}
	}
}

func TestTransformNilResultRemovesNode(t *testing.T) {
	root := HStack(Fixed(1, 1), Fixed(2, 2), Fixed(3, 3))

	// Removing a child.
	out := root.Transform(
		func(n *Node) bool { return n.Style.Width.Value == 2 },
		func(n *Node) *Node { return nil },
	)
	if out == nil {
		t.Fatalf("root should survive")
	}
	if len(out.Children) != 2 || out.Children[0].Style.Width.Value != 1 || out.Children[1].Style.Width.Value != 3 {
		t.Fatalf("nil result did not remove the node: %d children", len(out.Children))
	}

	// Removing the root returns nil instead of panicking.
	if got := root.Transform(func(*Node) bool { return true }, func(*Node) *Node { return nil }); got != nil {
		t.Fatalf("expected nil when the root is removed, got %+v", got)
	}
}

func TestMapHonorsChildListEditsAndNil(t *testing.T) {
	root := HStack(Fixed(1, 1), Fixed(2, 2), Fixed(3, 3))

	added := root.Map(func(n *Node) *Node {
		if n.Style.Display == DisplayFlex {
			return n.AddChild(Fixed(4, 4))
		}
		return n
	})
	if len(added.Children) != 4 {
		t.Fatalf("Map(AddChild): got %d children, want 4", len(added.Children))
	}

	pruned := root.Map(func(n *Node) *Node {
		if n.Style.Width.Value == 2 {
			return nil
		}
		return n
	})
	if len(pruned.Children) != 2 {
		t.Fatalf("Map(nil): got %d children, want 2", len(pruned.Children))
	}

	if got := root.Map(func(*Node) *Node { return nil }); got != nil {
		t.Fatalf("Map removing the root should return nil, got %+v", got)
	}
	if len(root.Children) != 3 {
		t.Fatalf("original mutated: %d children", len(root.Children))
	}
}

func TestTransformDoesNotMutateSharedChildSlice(t *testing.T) {
	// Clone shares the Children backing array; Transform must allocate a new
	// slice rather than writing transformed nodes into the original's array.
	root := HStack(Fixed(1, 1), Fixed(2, 2))
	orig0 := root.Children[0]
	_ = root.Transform(func(*Node) bool { return true }, func(n *Node) *Node { return n.WithWidth(99) })
	if root.Children[0] != orig0 || root.Children[0].Style.Width.Value != 1 {
		t.Fatalf("original child slice was mutated")
	}
}

func TestCloneDeepDoesNotShareStyleInternals(t *testing.T) {
	areas := NewGridTemplateAreas(2, 2)
	areas.DefineArea("header", 0, 1, 0, 2)
	orig := &Node{
		Style: Style{
			GridTemplateRows:    []GridTrack{FixedTrack(Px(10))},
			GridTemplateColumns: []GridTrack{FixedTrack(Px(20))},
			GridTemplateAreas:   areas,
			ContainerName:       ContainerName{"card"},
			TextStyle:           &TextStyle{FontSize: 12},
		},
		TextLayout: &TextLayout{
			LineHeight: 14,
			Lines: []TextLine{{
				Width: 5,
				Boxes: []InlineBox{{Text: "ab", Orientations: []bool{true, false}}},
			}},
		},
	}

	cp := orig.CloneDeep()

	cp.Style.GridTemplateRows[0] = FixedTrack(Px(999))
	cp.Style.GridTemplateColumns[0] = FixedTrack(Px(999))
	cp.Style.GridTemplateAreas.Rows = 42
	cp.Style.GridTemplateAreas.Areas[0].Name = "changed"
	cp.Style.ContainerName[0] = "changed"
	cp.Style.TextStyle.FontSize = 99
	cp.TextLayout.LineHeight = 99
	cp.TextLayout.Lines[0].Width = 99
	cp.TextLayout.Lines[0].Boxes[0].Text = "changed"
	cp.TextLayout.Lines[0].Boxes[0].Orientations[0] = false

	if orig.Style.GridTemplateRows[0].MinSize.Value != 10 {
		t.Errorf("GridTemplateRows shared with clone")
	}
	if orig.Style.GridTemplateColumns[0].MinSize.Value != 20 {
		t.Errorf("GridTemplateColumns shared with clone")
	}
	if orig.Style.GridTemplateAreas.Rows != 2 || orig.Style.GridTemplateAreas.Areas[0].Name != "header" {
		t.Errorf("GridTemplateAreas shared with clone")
	}
	if orig.Style.ContainerName[0] != "card" {
		t.Errorf("ContainerName shared with clone")
	}
	if orig.Style.TextStyle.FontSize != 12 {
		t.Errorf("TextStyle shared with clone")
	}
	if orig.TextLayout.LineHeight != 14 || orig.TextLayout.Lines[0].Width != 5 {
		t.Errorf("TextLayout shared with clone")
	}
	if orig.TextLayout.Lines[0].Boxes[0].Text != "ab" || !orig.TextLayout.Lines[0].Boxes[0].Orientations[0] {
		t.Errorf("TextLayout boxes/orientations shared with clone")
	}
}

func TestCloneDeepPreservesNilFields(t *testing.T) {
	cp := (&Node{}).CloneDeep()
	if cp.Style.TextStyle != nil || cp.Style.GridTemplateAreas != nil || cp.TextLayout != nil ||
		cp.Style.GridTemplateRows != nil || cp.Style.ContainerName != nil {
		t.Fatalf("CloneDeep must keep nil fields nil: %+v", cp)
	}
}

func TestRepeatTracksAutoCountsReturnEmpty(t *testing.T) {
	// Documented behavior: auto-fill/auto-fit cannot be expanded here.
	for _, count := range []int{RepeatCountAutoFill, RepeatCountAutoFit} {
		got := RepeatTracks(count, FixedTrack(Px(100)))
		if got == nil || len(got) != 0 {
			t.Errorf("RepeatTracks(%d) = %v, want empty non-nil slice", count, got)
		}
	}
	// Doc examples compile and behave.
	if got := RepeatTracks(3, FixedTrack(Px(100)), FractionTrack(1)); len(got) != 6 {
		t.Errorf("RepeatTracks(3, fixed, fr) len = %d, want 6", len(got))
	}
	if r := AutoFillTracks(FixedTrack(Px(100))); r.Count != RepeatCountAutoFill || len(r.Tracks) != 1 {
		t.Errorf("AutoFillTracks = %+v", r)
	}
	if r := AutoFitTracks(FixedTrack(Px(100))); r.Count != RepeatCountAutoFit || len(r.Tracks) != 1 {
		t.Errorf("AutoFitTracks = %+v", r)
	}
}
