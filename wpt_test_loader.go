package layout

import (
	"encoding/json"
	"fmt"
	"os"
)

// WPTTest represents a Web Platform Test in JSON format (Schema v1.0.0)
type WPTTest struct {
	Version     string                   `json:"version"`
	ID          string                   `json:"id"`
	Title       string                   `json:"title"`
	Description string                   `json:"description,omitempty"`
	Source      Source                   `json:"source"`
	Generated   Generated                `json:"generated"`
	Spec        Spec                     `json:"spec"`
	Categories  []string                 `json:"categories"`
	Tags        []string                 `json:"tags"`
	Properties  []string                 `json:"properties"`
	Layout      LayoutTree               `json:"layout"`
	Constraints Constraints              `json:"constraints"`
	Results     map[string]BrowserResult `json:"results"`
	Notes       []string                 `json:"notes,omitempty"`
}

type Source struct {
	URL    string  `json:"url"`
	File   string  `json:"file"`
	Commit *string `json:"commit,omitempty"`
}

type Generated struct {
	Timestamp string `json:"timestamp"`
	Tool      string `json:"tool"`
}

type Spec struct {
	Name    string `json:"name"`
	Section string `json:"section"`
	URL     string `json:"url"`
}

type LayoutTree struct {
	Type     string       `json:"type"`
	ID       string       `json:"id,omitempty"`
	Style    Style        `json:"style"`
	Text     string       `json:"text,omitempty"`
	Children []LayoutTree `json:"children,omitempty"`
}

type Style struct {
	Display        string   `json:"display,omitempty"`
	FlexDirection  string   `json:"flexDirection,omitempty"`
	FlexWrap       string   `json:"flexWrap,omitempty"`
	JustifyContent string   `json:"justifyContent,omitempty"`
	AlignItems     string   `json:"alignItems,omitempty"`
	AlignContent   string   `json:"alignContent,omitempty"`
	Width          *float64 `json:"width,omitempty"`
	Height         *float64 `json:"height,omitempty"`
	Padding        *Spacing `json:"padding,omitempty"`
	Margin         *Spacing `json:"margin,omitempty"`
}

type Spacing struct {
	Top    float64 `json:"top"`
	Right  float64 `json:"right"`
	Bottom float64 `json:"bottom"`
	Left   float64 `json:"left"`
}

type Constraints struct {
	Type   string  `json:"type"`
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type BrowserResult struct {
	Browser   Browser         `json:"browser"`
	Rendered  Rendered        `json:"rendered"`
	Elements  []ElementResult `json:"elements"`
	Tolerance *Tolerance      `json:"tolerance,omitempty"`
}

type Browser struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Engine  string `json:"engine"`
}

type Rendered struct {
	Timestamp string   `json:"timestamp"`
	Viewport  Viewport `json:"viewport"`
}

type Viewport struct {
	Width  float64 `json:"width"`
	Height float64 `json:"height"`
}

type ElementResult struct {
	ID       string                 `json:"id,omitempty"`
	Path     string                 `json:"path"`
	Expected map[string]interface{} `json:"expected"`
}

type Tolerance struct {
	Position float64 `json:"position"`
	Size     float64 `json:"size"`
	Numeric  float64 `json:"numeric"`
}

// LoadWPTTest loads a WPT test from a JSON file
func LoadWPTTest(filename string) (*WPTTest, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read test file: %w", err)
	}

	var test WPTTest
	if err := json.Unmarshal(data, &test); err != nil {
		return nil, fmt.Errorf("failed to parse test JSON: %w", err)
	}

	if test.Version != "1.0.0" {
		return nil, fmt.Errorf("unsupported schema version: %s", test.Version)
	}

	return &test, nil
}

// GetTolerance returns tolerance values, or defaults
func (result *BrowserResult) GetTolerance() Tolerance {
	if result.Tolerance != nil {
		return *result.Tolerance
	}
	return Tolerance{
		Position: 1.0,
		Size:     1.0,
		Numeric:  0.01,
	}
}
