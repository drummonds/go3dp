//-----------------------------------------------------------------------------
/*

Boundary wire inserter

*/
//-----------------------------------------------------------------------------

package main

import (
	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
	v2 "github.com/deadsy/sdfx/vec/v2"
	v3 "github.com/deadsy/sdfx/vec/v3"
)

//-----------------------------------------------------------------------------

const recessHeight = 2.0 // Top of srew below floor board recess within cover

//-----------------------------------------------------------------------------

func peg(l, h float64) (sdf.SDF3, error) {
	// Create simple peg shape
	points := []v2.Vec{
		{X: 0, Y: 0},
		{X: 4, Y: 0},
		{X: 3, Y: l - 10},
		{X: 1.5, Y: l},
		{X: 0.5, Y: l - 3},
		{X: -0.5, Y: l - 3},
		{X: -1.5, Y: l},
		{X: -3, Y: l - 10},
		{X: -4, Y: 0},
	}

	// Create a polygon from the points
	poly, err := sdf.Polygon2D(points)
	if err != nil {
		return nil, err
	}

	// Extrude to 3D
	peg, err := sdf.Extrude3D(poly, h), nil
	if err != nil {
		return nil, err
	}
	// peg = sdf.Transform3D(peg, sdf.Translate3d(v3.Vec{0, 0, h / 2}))
	return peg, nil
}

func boundaryInserter(length float64) sdf.SDF3 {
	peg, err := peg(length, 3)
	if err != nil {
		panic(err)
	}

	return peg
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
	s := boundaryInserter(70)
	// s = xArray(s, 7, 10)
	// s = yArray(s, 7, 10)
	// un-comment for a cut-away view
	//s = sdf.Cut3D(s, v3.Vec{0, 0, 0}, v3.Vec{1, 0, 0})
	render.ToSTL(s, "peg.stl", render.NewMarchingCubesOctree(500))
}

//-----------------------------------------------------------------------------
