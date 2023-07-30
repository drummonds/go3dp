//-----------------------------------------------------------------------------
/*
Button Cleaner

This is for cleaning silver buttons on a Tunic where you can't remove them.






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

// material shrinkage
const shrink = 1.0 / 0.999 // PLA ~0.1%
//const shrink = 1.0/0.995; // ABS ~0.5%

//-----------------------------------------------------------------------------

const (
	thickness  = 1.2 //
	diameter   = 70  //
	holedia    = 7
	notchWidth = 5
)

//-----------------------------------------------------------------------------

// This is the button cleaner
func buttonCleaner() (sdf.SDF3, error) {
	bc, _ := sdf.Cylinder3D(thickness, diameter*0.5, 0.0)
	hole, _ := sdf.Cylinder3D(thickness*1.1, holedia*0.5, 0.0)
	notch, _ := sdf.Box3D(sdf.V3{notchWidth, diameter * 0.5, thickness * 1.1}, 0)
	notch = sdf.Transform3D(notch, sdf.Translate3d(sdf.V3{0, diameter * 0.25, 0}))
	bc = sdf.Difference3D(bc, notch)
	bc = sdf.Difference3D(bc, hole)

	return bc, nil
}

//-----------------------------------------------------------------------------

func main() {
	var (
		fn string                   // Output file name
		f  func() (sdf.SDF3, error) // Render function
	)
	i := 0
	switch i {
	case 0:
		f = buttonCleaner
		fn = "buttonCleaner"
	}
	c, err := f()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	render.RenderSTL(sdf.ScaleUniform3D(c, shrink), 240,
		fmt.Sprintf("/mnt/c/t/%s.stl", fn))
}
