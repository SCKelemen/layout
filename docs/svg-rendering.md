# SVG Rendering

Layout produces rectangles; turning them into SVG is a loop over the tree. `examples/cards/main.go` is a complete program.

## Basic rendering

```go
package main

import (
	"fmt"
	"strings"

	"github.com/SCKelemen/layout"
)

func main() {
	root := &layout.Node{
		Style: layout.Style{
			Display: layout.DisplayGrid,
			GridTemplateColumns: []layout.GridTrack{
				layout.FixedTrack(layout.Px(150)),
				layout.FixedTrack(layout.Px(150)),
			},
			GridAutoRows: layout.FixedTrack(layout.Px(100)),
			GridGap:      layout.Px(10),
		},
		Children: []*layout.Node{{}, {}, {}, {}},
	}
	size := layout.LayoutSimple(root, layout.Loose(400, 300))

	var svg strings.Builder
	fmt.Fprintf(&svg, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f">`+"\n", size.Width, size.Height)
	for _, child := range root.Children {
		r := child.Rect
		fmt.Fprintf(&svg, `  <rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="#e0e0e0" stroke="#333"/>`+"\n",
			r.X, r.Y, r.Width, r.Height)
	}
	svg.WriteString("</svg>\n")
	fmt.Print(svg.String())
}
```

`Rect` is relative to the parent. For nested trees accumulate the parent offsets while walking (see the pattern below) or wrap each subtree in `<g transform="translate(x, y)">`.

## Transforms

`Style.Transform` is a 2D affine matrix that layout ignores and renderers apply:

```go
node := &layout.Node{
	Style: layout.Style{
		Width:     layout.Px(100),
		Height:    layout.Px(100),
		Transform: layout.RotateDegrees(45),
	},
}

attr := node.Style.Transform.ToSVGString() // matrix(a,b,c,d,e,f); "" for the identity
bounds := layout.GetFinalRect(node)        // bounding box after the transform
_, _ = attr, bounds
```

Constructors: `Translate(x, y)`, `Scale(sx, sy)`, `Rotate(radians)`, `RotateDegrees(deg)`, `SkewX(radians)`, `SkewY(radians)`, `Matrix(a, b, c, d, e, f)`, `IdentityTransform()`. `t1.Multiply(t2)` applies `t2` first. The zero value `Transform{}` is the identity, so untouched nodes render without a `transform` attribute.

Rotation is about the origin of the node's coordinate system; to rotate about the center, compose `Translate(cx, cy).Multiply(Rotate(a)).Multiply(Translate(-cx, -cy))`.

`GetSVGTransform(node)` and `CollectNodesForSVG(root, &nodes)` still exist but are deprecated in favor of `node.Style.Transform.ToSVGString()` and `root.DescendantsAndSelf()`.

## Rendering pattern with nesting and text

```go
func RenderToSVG(root *layout.Node, constraints layout.Constraints) string {
	size := layout.LayoutSimple(root, constraints)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f">`+"\n", size.Width, size.Height)

	var walk func(n *layout.Node, offsetX, offsetY float64)
	walk = func(n *layout.Node, offsetX, offsetY float64) {
		if n.Style.Display == layout.DisplayNone {
			return
		}
		x, y := offsetX+n.Rect.X, offsetY+n.Rect.Y
		attrs := fmt.Sprintf(`x="%.2f" y="%.2f" width="%.2f" height="%.2f"`, x, y, n.Rect.Width, n.Rect.Height)
		if t := n.Style.Transform.ToSVGString(); t != "" {
			attrs += fmt.Sprintf(` transform="%s"`, t)
		}
		fmt.Fprintf(&b, `  <rect %s fill="none" stroke="#333"/>`+"\n", attrs)

		if n.TextLayout != nil {
			for _, line := range n.TextLayout.Lines {
				lx := x + line.OffsetX
				ly := y + line.OffsetY
				for _, box := range line.Boxes {
					fmt.Fprintf(&b, `  <text x="%.2f" y="%.2f" font-size="%.0f">%s</text>`+"\n",
						lx, ly+box.Ascent, n.Style.TextStyle.FontSize, html.EscapeString(box.Text))
					lx += box.Width
					if box.SpaceAfter {
						lx += line.SpaceAdjustment + n.Style.TextStyle.FontSize*0.6 // space width from your metrics provider
					}
				}
			}
		}
		for _, child := range n.Children {
			walk(child, x, y)
		}
	}
	walk(root, 0, 0)

	b.WriteString("</svg>\n")
	return b.String()
}
```

(`html` is `html` from the standard library, `strings`/`fmt` likewise.) Measure the inter-word space with the same `TextMetricsProvider` you laid out with instead of the `0.6 x FontSize` approximation shown here.

## Tips

1. Use `LayoutWithPositioning` when the tree contains positioned nodes; `Rect` is still parent-relative afterwards.
2. `GetFinalRect` gives the bounding box after a transform; use it to size the canvas when transforms extend outside the layout.
3. Padding and borders are inside `Rect`; draw the content box at `Rect` inset by the resolved padding/border if your design distinguishes them (`layout.ResolveLength(n.Style.Padding.Left, ctx, fontSize)`).
4. `AlignNodes`, `DistributeNodes`, and `SnapNodes` adjust `Rect` after layout for design-tool operations on absolutely positioned or block boxes.
