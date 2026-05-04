package main

import (
	"fmt"

	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)

// V0CoverParams configures the slide-on cover. The cover envelops a
// wall-mounted v0 block: outer is an octagonal prism on the closed
// (-X) end joined to a rectangular slot box extending +X, with a
// V0BlockCut-shaped cavity on the inside. The +X face is open (the
// slide-in opening); the closed end's outline is octagonal so the
// cover echoes the block's plan-view shape with chamfered corners
// rather than sharp 90° corners.
type V0CoverParams struct {
	Tolerance       float32 // per-face cavity clearance, mm
	WallThickness   float32 // wall material on -X, +Y, -Y, +Z faces, mm
	LengthExtension float32 // slide-path overshoot past the block centre on +X, mm
}

// V0CoverPLA is the default preset for FDM PLA.
var V0CoverPLA = V0CoverParams{
	Tolerance:       0.15,
	WallThickness:   2.5,
	LengthExtension: 5.0,
}

// V0Cover returns the slide-on cover for a given block size. Outer
// shape = octagonal prism (across-flats Wo + 2·Tolerance + 2·Wall
// Thickness, vertical sides) at the closed end ∪ rectangular slot box
// extending +X to slideLen. Cavity = V0BlockCut with SlideLen =
// woc/2 + LengthExtension so the slot opens flush with the cover's
// +X face. Closed-end corners follow the cavity's octagonal outline;
// the slot-end corners are square.
//
// Coordinates: z=0 sits against the wall (open face), z=H_cover at
// the room-side top. Slide axis +X (the open face).
func V0Cover(bld *gsdf.Builder, sz V0Size, p V0CoverParams) (glbuild.Shader3D, error) {
	if p.Tolerance < 0 {
		return nil, fmt.Errorf("V0Cover %s: Tolerance must be non-negative", sz.Name)
	}
	if p.WallThickness <= 0 {
		return nil, fmt.Errorf("V0Cover %s: WallThickness must be positive", sz.Name)
	}
	if p.LengthExtension < 0 {
		return nil, fmt.Errorf("V0Cover %s: LengthExtension must be non-negative", sz.Name)
	}

	woc := sz.Wo + 2*p.Tolerance
	slideLen := woc/2 + p.LengthExtension
	zTop := sz.H + p.Tolerance + p.WallThickness
	outerOctW := woc + 2*p.WallThickness

	cavity, err := V0BlockCut(bld, sz, V0BlockCutParams{
		Tolerance: p.Tolerance,
		SlideLen:  slideLen,
	})
	if err != nil {
		return nil, fmt.Errorf("V0Cover %s cavity: %w", sz.Name, err)
	}

	octEnd, err := OctagonalPrism(bld, outerOctW, zTop)
	if err != nil {
		return nil, fmt.Errorf("V0Cover %s closed end: %w", sz.Name, err)
	}

	// Slot box covers X ∈ [0, slideLen]; Y matches the octagon's
	// across-flats so the slot box's ±Y faces are flush with the
	// octagon's ±Y flats.
	slotBox := bld.NewBox(slideLen, outerOctW, zTop, 0)
	slotBox = bld.Translate(slotBox, slideLen/2, 0, zTop/2)

	outer := bld.Union(octEnd, slotBox)

	return bld.Difference(outer, cavity), nil
}

// V0CoverDefault wraps V0Cover with V0CoverPLA defaults so it matches
// the partBuilder signature.
func V0CoverDefault(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	return V0Cover(bld, sz, V0CoverPLA)
}
