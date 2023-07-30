//-----------------------------------------------------------------------------
/*

Wood screw plug

This is like a plastiplug to fit into a piece of furntiture to take a wood screw.

No8 1/2"



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

const bottomThickness = 0.2 // Thickness of bottom of plug
// Just sufficient for one layer so that glue doesn't leak into body of plug
// if glued in
const innerDiameter = 3    // Thickness of body of screw at entry point actuall 2.85
const outerDiameter = 4    // Out diameter of hole in which to fit
const knurlThickness = 0.8 // Out diameter of hole in which to fit
const depth = 8.0          // length of hole in which to insert plug

//-----------------------------------------------------------------------------

// material shrinkage
const shrink = 1.0 / 0.999 // PLA ~0.1%
//const shrink = 1.0/0.995; // ABS ~0.5%

//-----------------------------------------------------------------------------

func hex() (sdf.SDF2, error) {
	return sdf.Polygon2D(sdf.Nagon(6, outerDiameter*0.5))
}

func knurledCylinder() (sdf.SDF3, error) {

	h, err := hex()
	if err != nil {
		return nil, err
	}

	// make the extrusions
	n := 1.0
	sFwd := sdf.TwistExtrude3D(sdf.Offset2D(h, 0), depth, sdf.Tau/n)
	sRev := sdf.TwistExtrude3D(sdf.Offset2D(h, 0), depth, -sdf.Tau/n)
	sCombo := sdf.Union3D(sFwd, sRev)

	// return a union of them all
	return sCombo, nil
}

func screwPlug() (sdf.SDF3, error) {

	h := depth
	r := outerDiameter*0.5 - knurlThickness
	outer, err := sdf.Cylinder3D(h, r, 1.0)
	if err != nil {
		return nil, err
	}
	knurl, err := knurledCylinder()
	if err != nil {
		return nil, err
	}
	outer = sdf.Union3D(outer, knurl)

	h = depth - bottomThickness
	r = innerDiameter * 0.5
	screw, err := screwModel()
	if err != nil {
		return nil, err
	}
	div, err := divider()
	if err != nil {
		return nil, err
	}

	outer = sdf.Difference3D(outer, div)

	return sdf.Difference3D(outer, screw), nil
}

// returns cone model of screw
func screwModel() (sdf.SDF3, error) {

	h := depth - bottomThickness
	r := innerDiameter * 0.5
	screw, err := sdf.Cone3D(h, 0.5, r, 0)
		screw = sdf.Transform3D(screw, sdf.Translate3d(sdf.V3{0, 0, 0.5}))
	if err != nil {
		return nil, err
	}

	return screw, nil
}

func split() (sdf.SDF3, error) {

	h := depth - bottomThickness
	r := innerDiameter * 0.5
	screw, err := sdf.Cone3D(h, 0.5, r, 0)
	screw = sdf.Transform3D(screw, sdf.Translate3d(sdf.V3{0, 0, 0.5}))
	if err != nil {
		return nil, err
	}

	return screw, nil
}

func knurledFinger() (sdf.SDF3, error) {

	h, err := sdf.Polygon2D(sdf.Nagon(15, 15*0.5))
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	depth := 20.0
	// make the extrusions
	n := 1.0
	sFwd := sdf.TwistExtrude3D(sdf.Offset2D(h, 0), depth, sdf.Tau/n)
	sRev := sdf.TwistExtrude3D(sdf.Offset2D(h, 0), depth, -sdf.Tau/n)
	sCombo := sdf.Union3D(sFwd, sRev)

	// return a union of them all
	return sCombo, nil
}

//Tool to make sure the holes are deep enough
func depthGauge() (sdf.SDF3, error) {
	head, _ := sdf.Cylinder3D(10, 8*0.5, 0.0)
	head = sdf.Transform3D(head, sdf.Translate3d(sdf.V3{0, 0, 15}))
	tip, _ := sdf.Cylinder3D(depth, outerDiameter*0.5, 0)
	tip = sdf.Transform3D(tip, sdf.Translate3d(sdf.V3{0, 0, 2 + 30 - depth}))
	finger, _ := knurledFinger()
	sCombo := sdf.Union3D(tip, head, finger)

	// return a union of them all
	return sCombo, nil
}

// Tool to push the plastic plugs into the holes
func pusher() (sdf.SDF3, error) {
	head, _ := sdf.Cylinder3D(10, 8*0.5, 0.0)
	head = sdf.Transform3D(head, sdf.Translate3d(sdf.V3{0, 0, 15}))
	h := depth - bottomThickness
	r := innerDiameter * 0.5
	tip, _ := sdf.Cone3D(h, r, 0.5, 0)
	tip = sdf.Transform3D(tip, sdf.Translate3d(sdf.V3{0, 0, 30 - h}))
	finger, _ := knurledFinger()
	sCombo := sdf.Union3D(tip, head, finger)

	// return a union of them all
	return sCombo, nil
}

// This is a splitter to allow screw to break apart plug as it bites it
func divider() (sdf.SDF3, error) {
	box, err := sdf.Box3D(sdf.V3{0.2, 10, 10}, 0)
	if err != nil {
		return box, err
	}
	box = sdf.Transform3D(box, sdf.Translate3d(sdf.V3{-0.1, 0, bottomThickness + depth/2}))
	return box, nil
}

//Tool to around drill to prevent drilling to deep
func collet() (sdf.SDF3, error) {
	length := 50 - depth
	finger, _ := knurledFinger()
	head, _ := sdf.Cylinder3D(length-20, 8*0.5, 0.0)
	head = sdf.Transform3D(head, sdf.Translate3d(sdf.V3{0, 0, 15}))
	drill, _ := sdf.Cylinder3D(length+10, 0.15+outerDiameter*0.5, 0)
	sCombo := sdf.Union3D(head, finger)
	sCombo = sdf.Difference3D(sCombo, drill)

	// return a union of them all
	return sCombo, nil
}

//-----------------------------------------------------------------------------

func main() {
	var (
		fn string
		f  func() (sdf.SDF3, error)
	)
	i := 5
	switch i {
	case 0:
		f = screwPlug
		fn = "screwPlug"
	case 1:
		f = knurledFinger
		fn = "knurledFinder"
	case 2:
		f = depthGauge
		fn = "depthGauge"
	case 3:
		f = pusher
		fn = "pusher"
	case 4:
		f = divider
		fn = "divider"
	case 5:
		f = collet
		fn = "collet"
	}
	c, err := f()
	if err != nil {
		log.Fatalf("error: %s", err)
	}
	render.RenderSTL(sdf.ScaleUniform3D(c, shrink), 240,
		fmt.Sprintf("/mnt/c/t/%s.stl", fn))
}

// func main() {
// 	// c, err := screwPlug()
// 	// c, err := cone()
// 	if err != nil {
// 		log.Fatalf("error: %s", err)
// 	}
// 	render.RenderSTL(sdf.ScaleUniform3D(c, shrink), 240, "/mnt/c/t/screwPlug.stl")
// }

//-----------------------------------------------------------------------------
