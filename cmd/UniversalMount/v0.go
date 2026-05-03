package main

import (
	"fmt"
	"math"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"

	"codeberg.org/hum3/go3dp/pkg/fasteners"
)

// V0Size describes a v0 wall block: an octagonal frustum with a fixed
// 45° outer slope, fastened to the wall with countersunk wood screws
// passing axially through the block. No central insert (deferred).
//
// Coordinates: small face at z=0 (against the wall), large face at z=H
// (room-facing). Wall-screw heads recess into the large face.
type V0Size struct {
	Name   string             // e.g. "v0-XS"
	Wi     float32            // small/wall face across flats, mm
	Wo     float32            // large/outer face across flats, mm
	H      float32            // block height (must equal (Wo-Wi)/2 for 45°)
	Screw  fasteners.WoodScrew // wall screw used for all hole positions
	Holes  []ms2.Vec          // (X,Y) screw centres on the outer face
	Recess float32            // depth the head sits below the outer face, mm
}

// V0Sizes is the canonical size series. Smallest first.
var V0Sizes = []V0Size{V0_XS, V0_S, V0_M, V0_L}

var (
	V0_XS = V0Size{
		Name: "v0-XS", Wi: 10, Wo: 14, H: 2,
		Screw:  fasteners.Spax3_5x16,
		Holes:  []ms2.Vec{{X: 0, Y: 0}},
		Recess: 0.25,
	}
	V0_S = V0Size{
		Name: "v0-S", Wi: 22, Wo: 26, H: 2,
		Screw:  fasteners.Spax3_5x16,
		Holes:  []ms2.Vec{{X: 0, Y: 6}, {X: 0, Y: -6}},
		Recess: 0.25,
	}
	V0_M = V0Size{
		Name: "v0-M", Wi: 30, Wo: 34, H: 2,
		Screw:  fasteners.Spax3_5x16,
		Holes:  triangleHoles(10),
		Recess: 0.25,
	}
	V0_L = V0Size{
		Name: "v0-L", Wi: 40, Wo: 44, H: 2,
		Screw:  fasteners.Spax3_5x16,
		Holes:  triangleHoles(15),
		Recess: 0.25,
	}
)

// triangleHoles returns three equally-spaced points on a circle of
// radius r, with the first vertex at +Y (12 o'clock).
func triangleHoles(r float32) []ms2.Vec {
	out := make([]ms2.Vec, 3)
	for i := 0; i < 3; i++ {
		a := math.Pi/2 + float64(i)*2*math.Pi/3
		out[i] = ms2.Vec{
			X: r * float32(math.Cos(a)),
			Y: r * float32(math.Sin(a)),
		}
	}
	return out
}

// V0Cover builds the matching cover cap for the given size: an inverted
// octagonal frustum with the large face at z=0 (sits on the block) and
// the small face at z=H (room-facing, visible). Screw cutouts are
// positioned at the same (X,Y) as the block, with the head recess at
// the visible (top) face and the shaft passing down through the cover
// into the block below.
//
// Coordinates: large face at z=0, small face at z=H. The cover stacks
// directly on top of a V0Block of the same size, so the combined Z
// extent runs from 0 (wall) through H (block top = cover bottom) to 2H
// (cover top, visible). When printed: the cover prints with its large
// face on the build plate (the "z=0 face") so all overhangs stay ≤ 45°.
func V0Cover(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	if sz.Wo-sz.Wi != 2*sz.H {
		return nil, fmt.Errorf("V0Cover %s: 45° rule violated (Wo=%v, Wi=%v, H=%v)",
			sz.Name, sz.Wo, sz.Wi, sz.H)
	}
	if len(sz.Holes) == 0 {
		return nil, fmt.Errorf("V0Cover %s: no screw holes specified", sz.Name)
	}

	// OctagonalFrustum produces small@z=0, large@z=H. We need
	// large@z=0, small@z=H — i.e., flip in Z. Rotating π about X
	// negates Y and Z; the octagon is symmetric so Y-flip is invisible.
	// After rotation: small@z=0 (unchanged) → small@z=0; large@z=H → z=-H.
	// Translate +H to put large@z=0 / small@z=H.
	frustum, err := OctagonalFrustum(bld, sz.Wi, sz.Wo, sz.H)
	if err != nil {
		return nil, err
	}
	frustum = bld.Rotate(frustum, float32(math.Pi), ms3.Vec{X: 1})
	frustum = bld.Translate(frustum, 0, 0, sz.H)

	// Screw cutout: head at the top (z=H), shaft passing down to z=-0.5
	// (overshoots z=0 for clean Difference). WallCutout has head@z=0,
	// shaft → -z; translate +H to put head@z=H.
	throughLen := sz.H + 0.5
	cutTemplate, err := sz.Screw.WallCutout(bld, throughLen, sz.Recess)
	if err != nil {
		return nil, err
	}
	cutTemplate = bld.Translate(cutTemplate, 0, 0, sz.H)

	for _, h := range sz.Holes {
		c := bld.Translate(cutTemplate, h.X, h.Y, 0)
		frustum = bld.Difference(frustum, c)
	}
	return frustum, nil
}

// V0Block builds the wall block for the given size: octagonal frustum
// minus a wall-screw cutout at each of sz.Holes.
func V0Block(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	if sz.Wo-sz.Wi != 2*sz.H {
		return nil, fmt.Errorf("V0Block %s: 45° rule violated (Wo=%v, Wi=%v, H=%v)",
			sz.Name, sz.Wo, sz.Wi, sz.H)
	}
	if len(sz.Holes) == 0 {
		return nil, fmt.Errorf("V0Block %s: no screw holes specified", sz.Name)
	}

	block, err := OctagonalFrustum(bld, sz.Wi, sz.Wo, sz.H)
	if err != nil {
		return nil, err
	}

	// 0.5 mm overshoot so the Difference cleanly punches through the
	// inner face without a numerical-noise sliver.
	throughLen := sz.H + 0.5
	cutTemplate, err := sz.Screw.WallCutout(bld, throughLen, sz.Recess)
	if err != nil {
		return nil, err
	}
	// WallCutout has its head opening at z=0; we want it at z=H (outer face).
	cutTemplate = bld.Translate(cutTemplate, 0, 0, sz.H)

	for _, h := range sz.Holes {
		c := bld.Translate(cutTemplate, h.X, h.Y, 0)
		block = bld.Difference(block, c)
	}
	return block, nil
}
