package main

import (
	"fmt"

	"github.com/SCKelemen/layout"
)

func main() {
	fmt.Println("=== Aspect Ratio Examples ===")
	fmt.Println()

	// Example 1: Image with fixed width, aspect ratio calculates height
	fmt.Println("Example 1: Image with 16:9 aspect ratio")
	image := &layout.Node{
		Style: layout.Style{
			Width: 800,
			Height: -1, // Explicitly set to auto
		},
	}
	image = layout.AspectRatio(image, 16.0/9.0)

	root1 := layout.VStack(image)
	root1.Style.Width = 1000

	constraints := layout.Loose(1000, layout.Unbounded)
	layout.Layout(root1, constraints)

	fmt.Printf("Image: %.2f x %.2f (aspect ratio: %.2f)\n",
		image.Rect.Width, image.Rect.Height, image.Rect.Width/image.Rect.Height)
	fmt.Printf("Expected: 800 x 450 (16:9 = 1.777...)\n")
	fmt.Println()

	// Example 2: Video that fills available width
	fmt.Println("Example 2: Video that fills container width")
	video := &layout.Node{
		Style: layout.Style{
			Width: -1,  // Auto - will use container width
			Height: -1, // Auto - will be calculated from aspect ratio
		},
	}
	video = layout.AspectRatio(video, 16.0/9.0)

	root2 := layout.VStack(video)
	root2.Style.Width = 1200

	layout.Layout(root2, constraints)

	fmt.Printf("Video: %.2f x %.2f (aspect ratio: %.2f)\n",
		video.Rect.Width, video.Rect.Height, video.Rect.Width/video.Rect.Height)
	fmt.Printf("Expected: 1200 x 675 (16:9 = 1.777...)\n")
	fmt.Println()

	// Example 3: Square cards in a grid
	fmt.Println("Example 3: Square cards in a grid")
	cards := []*layout.Node{
		layout.AspectRatio(&layout.Node{Style: layout.Style{Width: 200, Height: -1}}, 1.0),
		layout.AspectRatio(&layout.Node{Style: layout.Style{Width: 200, Height: -1}}, 1.0),
		layout.AspectRatio(&layout.Node{Style: layout.Style{Width: 200, Height: -1}}, 1.0),
	}

	root3 := layout.HStack(cards...)
	root3.Style.Width = 700

	layout.Layout(root3, constraints)

	for i, card := range cards {
		fmt.Printf("Card %d: %.2f x %.2f (aspect ratio: %.2f)\n",
			i+1, card.Rect.Width, card.Rect.Height, card.Rect.Width/card.Rect.Height)
	}
	fmt.Println()

	// Example 4: Aspect ratio with height set
	fmt.Println("Example 4: Element with height set, aspect ratio calculates width")
	element := &layout.Node{
		Style: layout.Style{
			Width: -1,  // Auto - will be calculated from aspect ratio
			Height: 300,
		},
	}
	element = layout.AspectRatio(element, 4.0/3.0)

	root4 := layout.VStack(element)
	root4.Style.Width = 1000

	layout.Layout(root4, constraints)

	fmt.Printf("Element: %.2f x %.2f (aspect ratio: %.2f)\n",
		element.Rect.Width, element.Rect.Height, element.Rect.Width/element.Rect.Height)
	fmt.Printf("Expected: 400 x 300 (4:3 = 1.333...)\n")
}

