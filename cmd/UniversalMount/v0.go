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
	Name          string              // e.g. "v0-XS"
	Wi            float32             // small/wall face across flats, mm
	Wo            float32             // large/outer face across flats, mm
	H             float32             // frustum height (must equal (Wo-Wi)/2 for 45°)
	Screw         fasteners.WoodScrew // wall screw used for all hole positions
	Holes         []ms2.Vec           // (X,Y) screw centres on the outer face
	Recess        float32             // 90° cone widening: head sits this far below the cone-top opening, mm
	HeadClearance float32             // extra flat-topped slab above the frustum (legacy; usually 0), mm
	Counterbore   float32             // cylindrical recess above the screw cone for sub-flush head, mm depth
}

// V0Sizes is the canonical size series. Smallest first.
var V0Sizes = []V0Size{V0_XS, V0_S, V0_M, V0_L}

var (
	V0_XS = V0Size{
		Name: "v0-XS", Wi: 10, Wo: 15, H: 2.5,
		Screw:       fasteners.Spax3_5x16,
		Holes:       []ms2.Vec{{X: 0, Y: 0}},
		Recess:      0.25,
		Counterbore: 0.5,
	}
	V0_S = V0Size{
		Name: "v0-S", Wi: 22, Wo: 27, H: 2.5,
		Screw:  fasteners.Spax3_5x16,
		Holes:  []ms2.Vec{{X: 0, Y: 6}, {X: 0, Y: -6}},
		Recess: 0.25,
	}
	V0_M = V0Size{
		Name: "v0-M", Wi: 30, Wo: 35, H: 2.5,
		Screw:  fasteners.Spax3_5x16,
		Holes:  triangleHoles(10),
		Recess: 0.25,
	}
	V0_L = V0Size{
		Name: "v0-L", Wi: 40, Wo: 45, H: 2.5,
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

// V0Block builds the wall block: octagonal frustum (Wi at z=0, Wo at
// z=H, 45° outward) capped by a 45°-inward-tapering octagonal frustum
// of HeadClearance height (Wo at z=H, Wo-2·HeadClearance at z=H+HeadClearance)
// so the screw head can sit sub-flush inside the cone-in cap. The screw
// cutout punches through the full new stack height.
func V0Block(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	if sz.Wo-sz.Wi != 2*sz.H {
		return nil, fmt.Errorf("V0Block %s: 45° rule violated (Wo=%v, Wi=%v, H=%v)",
			sz.Name, sz.Wo, sz.Wi, sz.H)
	}
	if len(sz.Holes) == 0 {
		return nil, fmt.Errorf("V0Block %s: no screw holes specified", sz.Name)
	}
	if sz.HeadClearance < 0 {
		return nil, fmt.Errorf("V0Block %s: HeadClearance must be non-negative", sz.Name)
	}

	block, err := OctagonalFrustum(bld, sz.Wi, sz.Wo, sz.H)
	if err != nil {
		return nil, err
	}
	if sz.HeadClearance > 0 {
		// 45° inward taper: across-flats decreases by 2·HeadClearance
		// over HeadClearance of height. wTop must remain positive — fail
		// loudly if HeadClearance is too aggressive for this size.
		wTop := sz.Wo - 2*sz.HeadClearance
		if wTop <= 0 {
			return nil, fmt.Errorf("V0Block %s: HeadClearance %v consumes the Wo=%v top face",
				sz.Name, sz.HeadClearance, sz.Wo)
		}
		// OctagonalFrustum gives small@z=0, large@z=h. Flip via 180° rotation
		// about X (negates Z), then translate so large face sits at z=H and
		// small face at z=H+HeadClearance.
		coneIn, err := OctagonalFrustum(bld, wTop, sz.Wo, sz.HeadClearance)
		if err != nil {
			return nil, fmt.Errorf("V0Block %s cone-in: %w", sz.Name, err)
		}
		coneIn = bld.Rotate(coneIn, float32(math.Pi), ms3.Vec{X: 1})
		coneIn = bld.Translate(coneIn, 0, 0, sz.H+sz.HeadClearance)
		block = bld.Union(block, coneIn)
	}

	// Screw cutout punches from the new top down through the bottom,
	// with a 0.5 mm overshoot so the Difference is clean. When
	// Counterbore > 0, the WallCutout (cone + shank) is shifted down by
	// Counterbore so the head opening lands at z=topZ-Counterbore, and a
	// matching cylindrical bore sits above for the head recess.
	topZ := sz.H + sz.HeadClearance
	throughLen := topZ + 0.5
	cutTemplate, err := sz.Screw.WallCutout(bld, throughLen, sz.Recess)
	if err != nil {
		return nil, err
	}
	headOpenZ := topZ - sz.Counterbore
	cutTemplate = bld.Translate(cutTemplate, 0, 0, headOpenZ)
	if sz.Counterbore > 0 {
		// rCB matches the cone's top-opening radius (rHeadCutout =
		// DHead/2 + Recess) so the cylinder merges seamlessly with the
		// head cone below.
		rCB := sz.Screw.DHead/2 + sz.Recess
		cb := bld.NewCylinder(rCB, sz.Counterbore, 0)
		cb = bld.Translate(cb, 0, 0, headOpenZ+sz.Counterbore/2)
		cutTemplate = bld.Union(cutTemplate, cb)
	}

	for _, h := range sz.Holes {
		c := bld.Translate(cutTemplate, h.X, h.Y, 0)
		block = bld.Difference(block, c)
	}

	return block, nil
}
