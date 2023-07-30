//-----------------------------------------------------------------------------
/*

Breadboard insert

This is an insert for our olive wood breadboard that doesn't glue nicelu
To space the bars.



*/
//-----------------------------------------------------------------------------

package main

import (
	"fmt"
	"log"

	"github.com/deadsy/sdfx/render"
	"github.com/deadsy/sdfx/sdf"
)

//-----------------------------------------------------------------------------

const (
	length     = 201 //
	width      = 20  //
	baseHeight = 16
	barWidth   = 14
	barHeight  = 3 // Actually 9 but don't want plastic space flush with top
	numBars    = 8
	firstSpace = 3 // Actually 8 but a inside a lip
)

//-----------------------------------------------------------------------------

// material shrinkage
const shrink = 1.0 / 0.999 // PLA ~0.1%
//const shrink = 1.0/0.995; // ABS ~0.5%

//-----------------------------------------------------------------------------

// box function which returns a box with corner at 0,0,0 rather than x/2,y/2,z/2
func SDFBox(x, y, z float64) sdf.SDF3 {
	box, err := sdf.Box3D(sdf.V3{x, y, z}, 0)
	if err != nil {
		panic("calc error")
	}
	box = sdf.Transform3D(box, sdf.Translate3d(sdf.V3{x / 2, y / 2, z / 2}))
	return box
}

// Going to shift spacer to end of bar
func endSpacer() sdf.SDF3 {
	box := SDFBox(width/2, barHeight, firstSpace)
	box = sdf.Transform3D(box, sdf.Translate3d(sdf.V3{0, baseHeight, 0}))
	return box
}

// This is the breadboard spacer
func completeSpacer() (sdf.SDF3, error) {
	box := SDFBox(width, baseHeight, length)
	box = sdf.Transform3D(box, sdf.Translate3d(sdf.V3{0, 0, 0}))
	// Add space at each end
	spacer := sdf.Transform3D(endSpacer(), sdf.Translate3d(sdf.V3{0, 0, 0}))
	box = sdf.Union3D(box, spacer)
	spacer = sdf.Transform3D(endSpacer(), sdf.Translate3d(sdf.V3{0, 0, length - firstSpace}))
	box = sdf.Union3D(box, spacer)
	// Add all the other spacers
	space := ((length - 2.0*firstSpace) - (numBars * barWidth)) / (numBars - 1)
	for i := 0; i < (numBars - 1); i++ {
		spacer := SDFBox(width/2, barHeight, space)
		offset := float64(firstSpace+barWidth) + float64(i)*(space+float64(barWidth))
		spacer = sdf.Transform3D(spacer, sdf.Translate3d(sdf.V3{0, baseHeight, offset}))
		box = sdf.Union3D(box, spacer)
	}

	return box, nil
}

//-----------------------------------------------------------------------------

func main() {
	var (
		fn string
		f  func() (sdf.SDF3, error)
	)
	i := 0
	switch i {
	case 0:
		f = completeSpacer
		fn = "spacer"
	}
	c, err := f()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	render.RenderSTL(sdf.ScaleUniform3D(c, shrink), 240,
		fmt.Sprintf("/mnt/c/t/%s.stl", fn))
}
