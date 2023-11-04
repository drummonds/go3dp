package main

import (
	"log"

	"github.com/soypat/sdf"
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
	// adaptor
	adaptor_height    = 10
	clearance         = 0.5
	adaptorWall       = 1.2
	adaptorBufferWall = 10 // This protects stack from res
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

func Pipe(height, d0, thickness, round float64) (s sdf.SDF3, err error) {
	var insert sdf.SDF3
	r0 := (d0 / 2)
	s, err = form3.Cone(height, r0, r0, round) // Cylinder returns ball
	if err != nil {
		return
	}
	r_inner := r0 - thickness
	insert, err = form3.Cone(height, r_inner, r_inner, round)
	if err != nil {
		return
	}
	s = sdf.Difference3D(s, insert)
	s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{Z: height / 2.0}))
	return
}

func mieleAdaptor() (s sdf.SDF3, err error) {
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
	// Create adaptor
	// First create the bit that sits inside the top of the bowl
	adaptor_top_dia := top_cone_dia - 2*wall - 2*clearance
	d1 = adaptor_top_dia
	d0 = d1 - (top_cone_dia-bottom_cone_dia)*(adaptor_height/height)
	adaptor, _ := Collet(adaptor_height, d0, d1, adaptorWall, 0)
	// add buffer
	d1 = bottom_cone_dia + (top_cone_dia-bottom_cone_dia)*(bufferHeight/height) + 2*adaptorBufferWall
	d0 = bottom_cone_dia + 2*adaptorBufferWall
	buffer, _ := Collet(bufferHeight, d0, d1, adaptorBufferWall, 0)
	buffer = sdf.Transform3D(buffer, sdf.Translate3D(r3.Vec{Z: adaptor_height}))
	// add Joiner going at 45deg for printability
	d1 = adaptor_top_dia + 2*joinerHeight
	d0 = adaptor_top_dia
	joiner, _ := Collet(joinerHeight, d0, d1, adaptorWall, 0)
	joiner = sdf.Transform3D(joiner, sdf.Translate3D(r3.Vec{Z: adaptor_height}))

	adaptor = sdf.Union3D(adaptor, buffer, joiner)
	// using top of base model as base line for adaptor
	adaptor = sdf.Transform3D(adaptor, sdf.Translate3D(r3.Vec{Z: height}))

	// Expand drawing
	adaptor = sdf.Transform3D(adaptor, sdf.Translate3D(r3.Vec{Z: 20}))

	bowl2 := sdf.Transform3D(bowl, sdf.Translate3D(r3.Vec{Z: height + 20 + 30}))

	// return sdf.Union3D(bowl, origin, adaptor, bowl2), err
	// Part
	_ = sdf.Union3D(bowl, origin, adaptor, bowl2)
	// This is the small adapter
	adaptor, _ = Pipe(45, 47.5, 2, 0)
	catch, _ := Pipe(1, 49, 2, 0)
	catch = sdf.Transform3D(catch, sdf.Translate3D(r3.Vec{Z: 15}))
	adaptor = sdf.Union3D(adaptor, catch)
	adaptor = sdf.Transform3D(adaptor, sdf.Translate3D(r3.Vec{Z: 41}))
	// Large adaptor groove on the inside
	big_adaptor, _ := Pipe(45, 56+3*2, 3, 0)
	groove, _ := Pipe(2, 56+1.5*2, 1.5+0.1, 0)
	groove = sdf.Transform3D(groove, sdf.Translate3D(r3.Vec{Z: 4}))
	big_adaptor = sdf.Difference3D(big_adaptor, groove)
	// Large adaptor groove on the inside
	joiner, _ = Pipe(10, 56+3*2, 7.5, 0)
	joiner = sdf.Transform3D(joiner, sdf.Translate3D(r3.Vec{Z: 37}))
	// Add two together
	adaptor = sdf.Union3D(adaptor, big_adaptor, joiner)
	return adaptor, err
}

func main() {
	// const quality = 200
	// b, _ := obj3.Bolt(obj3.BoltParms{
	// 	Thread:      "M16x2",
	// 	Style:       obj3.CylinderHex,
	// 	Tolerance:   0.1,
	// 	TotalLength: 101.,
	// 	ShankLength: 10.0,
	// })
	b, _ := mieleAdaptor()
	// err := uirender.EncodeRenderer(os.Stdout, render.NewOctreeRenderer(b, quality))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	err := render.CreateSTL("/mnt/c/t/mieleadaptor.stl", render.NewOctreeRenderer(b, 200))
	if err != nil {
		log.Fatal(err)
	}
}
