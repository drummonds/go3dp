package main

import (
	"log"
	"math"

	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r3"

	"github.com/soypat/sdf/form2"
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

func kickboard() sdf.SDF3 {
	// Use outside front bottom corner of kickboard as zero mark
	// Just used for design
	x := 16.0
	y := 300.0 // arbitrary
	z := 159.0
	part, _ := form3.Box(r3.Vec{X: x, Y: y, Z: z}, 0.0)
	part = sdf.Transform3D(part, sdf.Translate3D(r3.Vec{X: x / 2.0, Y: y / 2.0, Z: z / 2.0}))
	return part
}

func fridgeBody() sdf.SDF3 {
	// Use outside front bottom corner of kickboard as zero mark
	// Just used for design
	// This is the fridge main frame that leaks
	x := 50.0
	y := 300.0 // arbitrary
	z := 50.0
	part, _ := form3.Box(r3.Vec{X: x, Y: y, Z: z}, 0.0)
	part = sdf.Transform3D(part, sdf.Translate3D(r3.Vec{X: -4.0 + x/2.0, Y: y / 2.0, Z: 180.0 + z/2.0}))
	// This is wall of fridge main frame
	x = 10.0
	y = 300.0 // arbitrary
	z = 180.0
	part2, _ := form3.Box(r3.Vec{X: x, Y: y, Z: z}, 0.0)
	part2 = sdf.Transform3D(part2, sdf.Translate3D(r3.Vec{X: 52.0 + x/2.0, Y: y / 2.0, Z: z / 2.0}))
	part = sdf.Union3D(part, part2)
	// This is outer edge of drawer
	x = 19.0
	y = 300.0 // arbitrary
	z = 80.0
	part2, _ = form3.Box(r3.Vec{X: x, Y: y, Z: z}, 0.0)
	part2 = sdf.Transform3D(part2, sdf.Translate3D(r3.Vec{X: -66.0 + x/2.0, Y: y / 2.0, Z: 163.0 + z/2.0}))
	part = sdf.Union3D(part, part2)
	// plastic mat
	x = 1.0
	y = 300.0 // arbitrary
	z = 30.0
	part2, _ = form3.Box(r3.Vec{X: x, Y: y, Z: z}, 0.0)
	part2 = sdf.Transform3D(part2, sdf.Translate3D(r3.Vec{X: 52.0 - 9.0 + x/2.0, Y: y / 2.0, Z: z / 2.0}))
	part = sdf.Union3D(part, part2)
	return part
}

func drainerOutline(height float64, thickness float64) sdf.SDF3 {
	w := form2.NewPolygon()
	t := thickness
	// ***** Outer side
	// Riser
	w.Add(0, 0)
	w.Add(0, 80.0)
	//diagonal
	w.Add(-52.0+18.0+6.0, 164.0)
	w.Add(-52.0-4.0, 174.0)
	w.Add(-52.0-4.0, 195.0)
	// ***** Inner side
	w.Add(-52.0-4.0+t, 195.0)
	w.Add(-52.0-4.0+t, 174.0+t)
	w.Add(-52.0+18.0+6.0+t, 164.0+t)
	w.Add(0+t, 80.0+t*0.7)
	w.Add(0+t, 0)
	w.Add(0, 0)
	p, _ := form2.Polygon(w.Vertices())
	s := sdf.Extrude3D(p, height)
	s = sdf.Transform3D(s, sdf.RotateX(math.Pi/2))
	s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{X: -t, Y: height / 2}))
	return s
}

func drainerWidget() sdf.SDF3 {
	length := 190.0
	thickness := 3.0
	s := drainerOutline(length, thickness)
	// Add overlap
	overlap := drainerOutline(5.0+thickness, 2.0)
	// Ensure a single object so move X a little less
	overlap = sdf.Transform3D(overlap, sdf.Translate3D(r3.Vec{X: -thickness + 0.01, Y: length - thickness}))
	s = sdf.Union3D(s, overlap)
	// add ribs
	rib := drainerOutline(1.5, thickness+1.0)
	rib = sdf.Transform3D(rib, sdf.Translate3D(r3.Vec{X: 1.0, Y: 10.0}))
	// Make copy of first rib
	ribs := sdf.Transform3D(rib, sdf.Translate3D(r3.Vec{}))
	numRibs := int((length - 10.0) / 15.0)
	for i := 1; i < numRibs; i++ {
		rib = sdf.Transform3D(rib, sdf.Translate3D(r3.Vec{Y: 15.0}))
		ribs = sdf.Union3D(ribs, rib)
	}
	s = sdf.Union3D(s, ribs)
	// move on X axis to position
	s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{X: 51.0}))
	return s
}

func fridgeDrainer() (s sdf.SDF3, err error) {
	part := drainerWidget()
	// scaffolding := sdf.Union3D(kickboard(), fridgeBody())
	// part = sdf.Union3D(part, scaffolding)
	return part, err
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
	b, _ := fridgeDrainer()
	// err := uirender.EncodeRenderer(os.Stdout, render.NewOctreeRenderer(b, quality))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	err := render.CreateSTL("/mnt/c/t/fridgeDrainer.stl", render.NewOctreeRenderer(b, 200))
	if err != nil {
		log.Fatal(err)
	}
}
