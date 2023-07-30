package main

import (
	"log"
	"os"

	"github.com/soypat/sdf"
	"github.com/soypat/sdf3ui/uirender"
	"gonum.org/v1/gonum/spatial/r3"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/render"
)

const (
	// Bowl
	height          = 54.5 //
	lip_dia         = 110
	lip_height      = 6
	top_cone_dia    = 106.6 //
	bottom_cone_dia = 103.0 //
	wall            = 3.9   // Wall thickness
	// stacker
	stacker_height    = 10
	clearance         = 0.5
	stackerWall       = 1.2
	stackerBufferWall = 10 // This protects stack from res
	bufferHeight      = 10
	joinerHeight      = 4
)

// This is a hollow truncate cone and moved so the base is at the origin.
// It uses diameter rather than radius as that is what you measure
// d0 is the bottom
func Collet(height, d0, d1, thickness, round float64) (s sdf.SDF3, err error) {
	var insert sdf.SDF3
	r0 := (d0 / 2)
	r1 := (d1 / 2)
	s, err = form3.Cone(height, r0, r1, round)
	if err != nil {
		return
	}
	insert, err = form3.Cone(height, r0-thickness, r1-thickness, round)
	if err != nil {
		return
	}
	s = sdf.Difference3D(s, insert)
	s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{Z: height / 2.0}))
	return
}

func bowlStacker() (s sdf.SDF3, err error) {
	var (
		d0, d1 float64
	)
	origin, _ := form3.Box(r3.Vec{X: 2, Y: 2, Z: 2}, 0.0)
	// Model bowl
	bowl, _ := Collet(height, bottom_cone_dia, top_cone_dia, wall, 0)
	d0 = top_cone_dia - (top_cone_dia-bottom_cone_dia)*(lip_height/height)
	lip, _ := Collet(lip_height, d0, lip_dia, wall, 0.5)
	lip = sdf.Transform3D(lip, sdf.Translate3D(r3.Vec{Z: -lip_height + height}))
	bowl = sdf.Union3D(bowl, lip)
	// Create stacker
	// First create the bit that sits inside the top of the bowl
	stacker_top_dia := top_cone_dia - 2*wall - 2*clearance
	d1 = stacker_top_dia
	d0 = d1 - (top_cone_dia-bottom_cone_dia)*(stacker_height/height)
	stacker, _ := Collet(stacker_height, d0, d1, stackerWall, 0)
	// add buffer
	d1 = bottom_cone_dia + (top_cone_dia-bottom_cone_dia)*(bufferHeight/height) + 2*stackerBufferWall
	d0 = bottom_cone_dia + 2*stackerBufferWall
	buffer, _ := Collet(bufferHeight, d0, d1, stackerBufferWall, 0)
	buffer = sdf.Transform3D(buffer, sdf.Translate3D(r3.Vec{Z: stacker_height}))
	// add Joiner going at 45deg for printability
	d1 = stacker_top_dia + 2*joinerHeight
	d0 = stacker_top_dia
	joiner, _ := Collet(joinerHeight, d0, d1, stackerWall, 0)
	joiner = sdf.Transform3D(joiner, sdf.Translate3D(r3.Vec{Z: stacker_height}))

	stacker = sdf.Union3D(stacker, buffer, joiner)
	// using top of base model as base line for stacker
	stacker = sdf.Transform3D(stacker, sdf.Translate3D(r3.Vec{Z: height}))

	// Expand drawing
	stacker = sdf.Transform3D(stacker, sdf.Translate3D(r3.Vec{Z: 20}))

	bowl2 := sdf.Transform3D(bowl, sdf.Translate3D(r3.Vec{Z: height + 20 + 30}))

	// return sdf.Union3D(bowl, origin, stacker, bowl2), err
	// Part
	_ = sdf.Union3D(bowl, origin, stacker, bowl2)
	return stacker, err
}

func main() {
	const quality = 200
	// b, _ := obj3.Bolt(obj3.BoltParms{
	// 	Thread:      "M16x2",
	// 	Style:       obj3.CylinderHex,
	// 	Tolerance:   0.1,
	// 	TotalLength: 101.,
	// 	ShankLength: 10.0,
	// })
	b, _ := bowlStacker()
	err := uirender.EncodeRenderer(os.Stdout, render.NewOctreeRenderer(b, quality))
	if err != nil {
		log.Fatal(err)
	}
	err = render.CreateSTL("/mnt/c/t/bowlstacker.stl", render.NewOctreeRenderer(b, 200))
	if err != nil {
		log.Fatal(err)
	}
}
