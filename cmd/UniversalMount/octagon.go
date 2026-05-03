package main

import (
	"errors"
	"math"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)


// OctagonalFrustum builds an octagonal frustum aligned along Z.
// The small face (across-flats = wInner) sits at z=0; the large face
// (across-flats = wOuter) sits at z=h. The taper angle is set implicitly
// by the dimensions — for a 45° taper choose h = (wOuter-wInner)/2.
//
// Construction: intersection of 4 trapezoidal prisms rotated 0/45/90/135°
// about Z. Each prism's trapezoidal cross-section in its local XZ plane
// constrains one pair of opposing flats.
func OctagonalFrustum(bld *gsdf.Builder, wInner, wOuter, h float32) (glbuild.Shader3D, error) {
	if wInner <= 0 || wOuter <= 0 || h <= 0 {
		return nil, errors.New("OctagonalFrustum: positive dimensions required")
	}
	if wOuter <= wInner {
		return nil, errors.New("OctagonalFrustum: wOuter must exceed wInner")
	}

	prism := trapezoidalPrism(bld, wInner, wOuter, h)
	rotZ := func(s glbuild.Shader3D, rad float32) glbuild.Shader3D {
		return bld.Rotate(s, rad, ms3.Vec{Z: 1})
	}
	p0 := prism
	p45 := rotZ(prism, float32(math.Pi/4))
	p90 := rotZ(prism, float32(math.Pi/2))
	p135 := rotZ(prism, float32(3*math.Pi/4))

	out := bld.Intersection(p0, p45)
	out = bld.Intersection(out, p90)
	out = bld.Intersection(out, p135)
	// trapezoidalPrism centres the cross-section on z=0 (vertices at ±h/2),
	// so shift up by h/2 to put the small face at z=0 and the large face at z=h.
	out = bld.Translate(out, 0, 0, h/2)
	return out, nil
}

// trapezoidalPrism builds a prism whose XZ cross-section is the trapezoid
//
//	(-wOuter/2, +h/2), (wOuter/2, +h/2), (wInner/2, -h/2), (-wInner/2, -h/2)
//
// extruded along the world Y axis (long enough to span the largest octagon
// diagonal without hitting the prism end caps).
//
// Caller still has to translate the result by +h/2 in Z to place the small
// face at world z=0 and the large face at world z=h.
func trapezoidalPrism(bld *gsdf.Builder, wInner, wOuter, h float32) glbuild.Shader3D {
	verts := []ms2.Vec{
		{X: -wOuter / 2, Y: +h / 2},
		{X: +wOuter / 2, Y: +h / 2},
		{X: +wInner / 2, Y: -h / 2},
		{X: -wInner / 2, Y: -h / 2},
	}
	trap := bld.NewPolygon(verts) // 2D in builder's XY plane

	// Extrude the polygon along the builder's Z axis. The extrude length
	// must exceed the largest octagonal diagonal (wOuter / cos(22.5°))
	// after the prism is rotated about world Z and intersected with its
	// siblings; double wOuter is comfortably enough.
	extrudeLen := wOuter * 2
	prism := bld.Extrude(trap, extrudeLen)

	// gsdf's Extrude lays the 2D polygon in the builder's XY plane and
	// extrudes along Z. We want the trapezoid lying in the world XZ plane
	// and the prism extending along world Y. Rotate 90° about X so the
	// builder's Y axis (the trapezoid's vertical edge) maps to world Z,
	// and the extrude axis (builder Z) maps to world Y.
	prism = bld.Rotate(prism, float32(math.Pi/2), ms3.Vec{X: 1})
	return prism
}
