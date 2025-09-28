//-----------------------------------------------------------------------------
/*

torx screw cover

*/
//-----------------------------------------------------------------------------

package main

import (
	"math"

	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	v2 "github.com/deadsy/sdfx/vec/v2"
	v3 "github.com/deadsy/sdfx/vec/v3"
)

//-----------------------------------------------------------------------------

const recessHeight = 2.0 // Top of srew below floor board recess within cover

// TorxSize represents the different Torx drive sizes
type TorxSize int

const (
	T15 TorxSize = iota
	T20
	T20Spax
	T25
)

// TorxSpec contains the specifications for a Torx drive
type TorxSpec struct {
	width float64 // major diameter
	depth float64 // recommended depth
	plug  float64 // plug size
}

var torxSpecs = map[TorxSize]TorxSpec{
	T15:     {width: 2.30, depth: 1.5, plug: 8.0},
	T20:     {width: 2.68, depth: 2.0, plug: 8.0},
	T20Spax: {width: 2.68, depth: 1.92, plug: 8.0},
	T25:     {width: 2.95, depth: 2.2, plug: 12.0},
}

//-----------------------------------------------------------------------------

// Torx shapes
// https://www.printables.com/model/680108-parametric-torx-template/files
// M3.5 #6 T10
// M4   #8 T10
// M5   #10 T10
//  Plug cutters at 3/8" 8mm  and 1/2" 12.5mm
//  Screw sizes #4

func cover(plugDiameter float64) (sdf.SDF3, error) {
	r := plugDiameter / 2.0
	h := recessHeight
	cover, err := sdf.Cylinder3D(h, r, 0.1*r)
	if err != nil {
		return nil, err
	}
	cover = sdf.Transform3D(cover, sdf.Translate3d(v3.Vec{0, 0, h / 2}))
	return cover, nil
}

// Allow the cover to be removed easily
func extractor(recessHeight float64) (sdf.SDF3, error) {
	r := 0.6
	h := 10.0
	extractor, err := sdf.Cylinder3D(h, r, 0.1*r)
	if err != nil {
		return nil, err
	}
	extractor = sdf.Transform3D(extractor, sdf.Translate3d(v3.Vec{-1, 0, recessHeight - h/2}))
	extractor = sdf.Transform3D(extractor, sdf.Rotate3d(v3.Vec{0, 1, 0}, math.Pi/4))
	return extractor, nil
}

func torx(A, h float64) (sdf.SDF3, error) {
	// A is the major diameter of the Torx pattern
	B := A * 0.72 // Minor diameter

	// Create the basic star pattern
	points := make([]v2.Vec, 12) // 12 points * 2 coordinates each
	for i := 0; i < 6; i++ {
		angle := float64(i) * math.Pi / 3.0
		// Outer point
		points[i*2].X = math.Cos(angle) * (A / 2) // x
		points[i*2].Y = math.Sin(angle) * (A / 2) // y
		// Inner point
		points[i*2+1].X = math.Cos(angle+math.Pi/6) * (B / 2) // x
		points[i*2+1].Y = math.Sin(angle+math.Pi/6) * (B / 2) // y
	}

	// Create a polygon from the points
	poly, err := sdf.Polygon2D(points)
	if err != nil {
		return nil, err
	}

	// Round the corners
	roundR := A * 0.15 // radius for rounding corners
	torxProfile := sdf.Offset2D(poly, roundR)

	// Extrude to 3D
	torx, err := sdf.Extrude3D(torxProfile, h), nil
	if err != nil {
		return nil, err
	}
	torx = sdf.Transform3D(torx, sdf.Translate3d(v3.Vec{0, 0, h / 2}))
	return torx, nil
}

func screwCover(size TorxSize) sdf.SDF3 {
	// Using the new torxSpecs map
	spec := torxSpecs[size] // Change T15 to desired size
	cover, err := cover(spec.plug)
	if err != nil {
		panic(err)
	}
	torx, err := torx(spec.width, recessHeight+spec.depth)
	if err != nil {
		panic(err)
	}
	extractor, err := extractor(recessHeight) //	h := recessHeight
	if err != nil {
		panic(err)
	}
	cover = sdf.Union3D(cover, torx)
	cover = sdf.Difference3D(cover, extractor)
	// return sdf.Cut3D(cover, v3.Vec{0, 0, 0}, v3.Vec{0, 0, 1}), nil
	return cover
}

func sizeArray() sdf.SDF3 {
	list := []TorxSize{T20, T20, T20, T20, T20, T20, T20}
	var result sdf.SDF3
	offset := 0.0
	for i, size := range list {
		spec := torxSpecs[size] // Change T15 to desired size
		sc := screwCover(size)
		if i != 0 {
			offset += spec.plug + 2.0
		}
		sc = sdf.Transform3D(sc, sdf.Translate3d(v3.Vec{offset, 0, 0}))
		if i == 0 {
			result = sc
		} else {
			result = sdf.Union3D(result, sc)
		}
	}
	return result
}

func xArray(s sdf.SDF3, reps int, spacing float64) sdf.SDF3 {
	var result sdf.SDF3
	for i := 0; i < reps; i++ {
		newS := sdf.Transform3D(s, sdf.Translate3d(v3.Vec{float64(i) * spacing, 0, 0}))
		if i == 0 {
			result = newS
		} else {
			result = sdf.Union3D(result, newS)
		}
	}
	return result
}

func yArray(s sdf.SDF3, reps int, spacing float64) sdf.SDF3 {
	var result sdf.SDF3
	for i := 0; i < reps; i++ {
		newS := sdf.Transform3D(s, sdf.Translate3d(v3.Vec{0, float64(i) * spacing, 0}))
		if i == 0 {
			result = newS
		} else {
			result = sdf.Union3D(result, newS)
		}
	}
	return result
}

func main() {
	// s := sizeArray()
	s := screwCover(T20)
	s = xArray(s, 7, 10)
	s = yArray(s, 7, 10)
	// un-comment for a cut-away view
	//s = sdf.Cut3D(s, v3.Vec{0, 0, 0}, v3.Vec{1, 0, 0})
	render.ToSTL(s, "cover.stl", render.NewMarchingCubesOctree(1500))
}

//-----------------------------------------------------------------------------
