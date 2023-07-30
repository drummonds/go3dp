// This is seperator to go between to stacks of plate or bowls.
// It consists of a foot plate and a seperator
// The Stacks of plates are modelled as two cylinders. The bottom one being the
// clearance under the plate
package main

import (
	"log"

	"github.com/soypat/sdf"
	"gonum.org/v1/gonum/spatial/r3"

	"github.com/soypat/sdf/form3"
	"github.com/soypat/sdf/render"
)

type PlateStack struct {
	Height       float64 // overall
	Diameter     float64
	FootDiameter float64
	FootHeight   float64
}

func NewPlateStack(dia float64) (s PlateStack) {
	return PlateStack{
		Height:       100,
		FootHeight:   8,
		Diameter:     dia,
		FootDiameter: dia - 20,
	}
}

const (
	bowlDia     = 240
	medPlateDia = 220
	gap         = 40
)

// Create this centered XY and bottom Z=0
func StackToSdf(dia float64) (s sdf.SDF3, err error) {
	ps := NewPlateStack(dia)
	h1 := ps.Height - ps.FootHeight
	s, err = form3.Cylinder(h1, ps.Diameter, 0)
	if err != nil {
		return
	}
	s = sdf.Transform3D(s, sdf.Translate3D(r3.Vec{Z: ps.FootHeight + h1/2.0}))
	foot, err1 := form3.Cylinder(ps.FootHeight, ps.FootDiameter, 0)
	if err1 != nil {
		return
	}
	foot = sdf.Transform3D(foot, sdf.Translate3D(r3.Vec{Z: ps.FootHeight / 2.0}))
	s = sdf.Union3D(s, foot)
	return
}

func Seperator() (s sdf.SDF3, err error) {
	s1, _ := StackToSdf(bowlDia)
	s2, _ := StackToSdf(medPlateDia)
	s2 = sdf.Transform3D(s2, sdf.Translate3D((r3.Vec{Y: gap + (bowlDia+medPlateDia)/2.0})))
	s = sdf.Union3D(s1, s2)
	return s1, err
}

func main() {
	var err error
	const quality = 20
	b, _ := Seperator()
	b, _ = form3.Cylinder(20, 240, 0)
	// b, _ = form3.Cone(100, 240, 240, 0.1)

	// err = uirender.EncodeRenderer(os.Stdout, render.NewOctreeRenderer(b, quality))
	if err != nil {
		log.Fatal(err)
	}
	err = render.CreateSTL("/mnt/c/t/separator.stl", render.NewOctreeRenderer(b, quality))
	if err != nil {
		log.Fatal(err)
	}
}
