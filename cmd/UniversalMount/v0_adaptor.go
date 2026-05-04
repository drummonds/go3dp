package main

import (
	"fmt"

	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)

// V0AdaptorParams holds tolerances and wall thicknesses common to the
// V0 adaptor family — slide-on parts that mate with a V0Block via its
// outer octagonal-frustum geometry.
//
// Defaults (V0AdaptorPLA) target FDM PLA at 0.4 mm nozzle / 0.2 mm
// layer height. Other materials need different numbers — see ROADMAP.
type V0AdaptorParams struct {
	Tolerance       float32 // clearance per face between v0 mount and adaptor cavity, mm
	WallThickness   float32 // material thickness around the cavity, mm
	LengthExtension float32 // adaptor length past the v0 footprint along the slide axis, mm
	CavityLip       float32 // flat lip at the cavity's wall-side entrance (z=0..CavityLip at Wo+2·Tol cross-section), mm
}

// V0AdaptorPLA is the default preset for FDM PLA.
var V0AdaptorPLA = V0AdaptorParams{
	Tolerance:       0.15,
	WallThickness:   1.5,
	LengthExtension: 5.0,
	CavityLip:       0.5,
}

// V0SlideOn is the first member of the V0 adaptor family: a lateral
// slide-on sleeve.
//
// Outer profile: rectangular cuboid for most of the length, with a
// half-octagonal cap at the closed (-X) end whose external surfaces
// follow the v0 mount's frustum at +WallThickness offset. Inside is a
// pocket matching the v0 mount + Tolerance per face, plus a rectangular
// slot that opens the cavity to the +X face so the v0 mount can slide
// in laterally.
//
// Coordinates: v0 mount centred on origin, axis along +Z. After full
// engagement the v0 mount sits centred at origin within the adaptor's
// pocket. The adaptor's z=0 face is open (against the wall, butts
// against it just like the v0 mount). The adaptor's +Z face is closed
// by p.WallThickness of material.
//
// Slide axis: +X. The user pushes the adaptor in -X over a wall-fixed
// v0 mount.
func V0SlideOn(bld *gsdf.Builder, sz V0Size, p V0AdaptorParams) (glbuild.Shader3D, error) {
	if p.Tolerance < 0 {
		return nil, fmt.Errorf("V0SlideOn: Tolerance must be non-negative")
	}
	if p.WallThickness <= 0 {
		return nil, fmt.Errorf("V0SlideOn: WallThickness must be positive")
	}
	if p.LengthExtension < 0 {
		return nil, fmt.Errorf("V0SlideOn: LengthExtension must be non-negative")
	}
	if p.CavityLip < 0 {
		return nil, fmt.Errorf("V0SlideOn: CavityLip must be non-negative")
	}

	// Cavity dimensions: matches v0 mount + tolerance per face.
	wic := sz.Wi + 2*p.Tolerance
	woc := sz.Wo + 2*p.Tolerance
	houter := sz.H + p.WallThickness

	// 1) Outer body — rectangular plan view, WallThickness uniform on
	//    the -X, +Y, and -Y faces; +X is the slide-in opening (slot cut
	//    from the cavity below). Drops the previous chamfered "flange"
	//    corners now that retention is intended to come from a tongue-
	//    and-groove against the block (TBD).
	xLeft := -woc/2 - p.WallThickness
	xRight := woc/2 + p.LengthExtension
	yHalf := woc/2 + p.WallThickness
	outerLenX := xRight - xLeft
	outerWidthY := 2 * yHalf
	outer := bld.NewBox(outerLenX, outerWidthY, houter, 0)
	outer = bld.Translate(outer, (xLeft+xRight)/2, 0, houter/2)

	// 3) Cavity: octagonal frustum at origin matching v0 + tolerance,
	//    plus an optional flat lip at the wall-side entrance (z=0..lip)
	//    at the largest cross-section (Wo+2·Tol). The lip leaves a
	//    sliding-clearance gap around the v0 mount's narrow end so the
	//    adaptor doesn't bind on the wall during lateral install.
	cavity, err := OctagonalFrustum(bld, wic, woc, sz.H)
	if err != nil {
		return nil, fmt.Errorf("V0SlideOn cavity: %w", err)
	}
	if p.CavityLip > 0 {
		lip, err := OctagonalPrism(bld, woc, p.CavityLip)
		if err != nil {
			return nil, fmt.Errorf("V0SlideOn cavity lip: %w", err)
		}
		cavity = bld.Union(cavity, lip)
	}

	// 4) Slot: a Woc × H rectangular prism extending the cavity from
	//    the centre out past the +X open face. Wider than the v0
	//    mount's YZ silhouette at every X, so the v0 fits during the
	//    slide. The 1 mm overshoot past the open face guarantees a
	//    clean Difference (no numerical sliver).
	slotLen := woc/2 + p.LengthExtension + 1
	slot := bld.NewBox(slotLen, woc, sz.H, 0)
	slot = bld.Translate(slot, slotLen/2, 0, sz.H/2)

	cavityFull := bld.Union(cavity, slot)
	return bld.Difference(outer, cavityFull), nil
}

// V0SlideOnPLA wraps V0SlideOn with V0AdaptorPLA defaults so it matches
// the partBuilder signature used by the CLI.
func V0SlideOnPLA(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	return V0SlideOn(bld, sz, V0AdaptorPLA)
}

// V0Assembly returns the boolean union of the V0 block and its slide-on
// adaptor in the assembled position (both centred on origin). Useful as
// a debugging cutaway: cross-sections show the cavity-vs-block clearance,
// the wall-side lip gap, the slot opening, and the top wall — all in one
// SVG. The 3MF/STL of the union is rarely what you want to print.
func V0Assembly(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	block, err := V0Block(bld, sz)
	if err != nil {
		return nil, fmt.Errorf("V0Assembly block: %w", err)
	}
	adaptor, err := V0SlideOnPLA(bld, sz)
	if err != nil {
		return nil, fmt.Errorf("V0Assembly adaptor: %w", err)
	}
	return bld.Union(block, adaptor), nil
}
