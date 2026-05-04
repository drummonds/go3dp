package main

import (
	"fmt"
	"math"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)

// V0BlockCutParams configures the negative volume to subtract from a
// host (variant) printed object so a v0 block can slide laterally into
// the host along +X and land centred at the origin.
type V0BlockCutParams struct {
	Tolerance float32 // per-face clearance applied to the cavity, mm
	SlideLen  float32 // slot length along +X past the block centre, mm
}

// V0BlockCutPLA is the default preset for FDM PLA. Tolerance matches
// V0AdaptorPLA. SlideLen of 25 mm clears the wall of most utility
// objects (cups, holders, brackets); increase if the host is thicker
// than this along the slide axis.
var V0BlockCutPLA = V0BlockCutParams{
	Tolerance: 0.15,
	SlideLen:  25,
}

// V0BlockCut returns the cutaway: octagonal frustum (block + tolerance
// per face) at the origin, unioned with a tapered slot that extends
// +X by SlideLen so the block can slide in from outside the host.
//
// Cross-section of the slot in YZ is the block's silhouette plus
// tolerance — a trapezoid (Wi+2·tol at z=0, Wo+2·tol at z=H, 45° sides)
// so every external surface of the eventual void stays at ≤45° and the
// host prints support-free in the same orientation as the block. The
// +X end of the slot is a flat (square) face perpendicular to X.
//
// Coordinates: same Z convention as V0Block (small Wi face at z=0
// against the wall, large Wo face at z=H room-side). Slide axis +X.
func V0BlockCut(bld *gsdf.Builder, sz V0Size, p V0BlockCutParams) (glbuild.Shader3D, error) {
	if sz.Wo-sz.Wi != 2*sz.H {
		return nil, fmt.Errorf("V0BlockCut %s: 45° rule violated (Wo=%v, Wi=%v, H=%v)",
			sz.Name, sz.Wo, sz.Wi, sz.H)
	}
	if p.Tolerance < 0 {
		return nil, fmt.Errorf("V0BlockCut %s: Tolerance must be non-negative", sz.Name)
	}
	if p.SlideLen <= 0 {
		return nil, fmt.Errorf("V0BlockCut %s: SlideLen must be positive", sz.Name)
	}

	wic := sz.Wi + 2*p.Tolerance
	woc := sz.Wo + 2*p.Tolerance

	cavity, err := OctagonalFrustum(bld, wic, woc, sz.H)
	if err != nil {
		return nil, fmt.Errorf("V0BlockCut %s cavity: %w", sz.Name, err)
	}

	// Slot: trapezoidal YZ cross-section (Y width = wic at z=0, woc at
	// z=H, 45° sides) extruded along +X for SlideLen. The polygon lives
	// in the builder's XY plane (builder X = world Y, builder Y = world
	// Z); after extrude+rotate, builder Z maps to world X.
	profile := []ms2.Vec{
		{X: -wic / 2, Y: 0},
		{X: +wic / 2, Y: 0},
		{X: +woc / 2, Y: sz.H},
		{X: -woc / 2, Y: sz.H},
	}
	poly := bld.NewPolygon(profile)
	prism := bld.Extrude(poly, p.SlideLen)
	// builder X→world Y, builder Y→world Z, builder Z→world X.
	prism = bld.Rotate(prism, float32(math.Pi/2), ms3.Vec{X: 1})
	prism = bld.Rotate(prism, float32(math.Pi/2), ms3.Vec{Z: 1})
	// Extrude is centred on builder Z=0 (→ world X=0); shift so the
	// slot spans X ∈ [0, SlideLen].
	prism = bld.Translate(prism, p.SlideLen/2, 0, 0)

	return bld.Union(cavity, prism), nil
}

// V0BlockCutDefault wraps V0BlockCut with V0BlockCutPLA defaults so it
// matches the partBuilder signature used by the CLI.
func V0BlockCutDefault(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	return V0BlockCut(bld, sz, V0BlockCutPLA)
}
