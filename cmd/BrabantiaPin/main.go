// 0,0,0 = z= 0, centre of tip end
package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/soypat/glgl/math/ms2"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
	"github.com/soypat/gsdf/gleval"
	"github.com/soypat/gsdf/glrender"
)

const (
	// Insert
	insideClip  = 9.0 // Length inside retention clip to surface of holder
	outsideClip = 8.6 // when closed to bottom of lip
	barWidth    = 3.3 // 4.3 new, 3.75 Chinese
	barHeight   = 4.0
	headWidth   = 5.9
	headHeight  = 1.8
	headRear    = 1.0
	headNeck    = 7.5
	baseWidth   = 20.0
	baseHeight  = 4.0
)

// // box function which returns a box with corner at 0,0,0 rather than x/2,y/2,z/2
// func SDFBox(x, y, z float64) sdf.SDF3 {
// 	box, err := sdf..Box3D(sdf.V3{x, y, z}, 0)
// 	if err != nil {
// 		panic("calc error")
// 	}
// 	box = sdf.Transform3D(box, sdf.Translate3d(sdf.V3{x / 2, y / 2, z / 2}))
// 	return box
// }

func BrabantiaPin() (s gleval.SDF3, err error) {
	// Model bar
	// Base plate which goes bolted to joint.
	// bar, err := gsdf.NewBox(barWidth, insideClip+outsideClip, barHeight, 0.1)
	// if err != nil {
	// 	panic("failed NewBox")
	// }
	// // Move bar to side so that can look at it separately
	// bar = gsdf.Translate(bar, -10.0, 0, 0)

	// Build head
	var poly ms2.PolygonBuilder
	poly.AddXY(0, 0)
	poly.AddXY(0.5, 0)
	poly.AddXY(1.0, 0)
	poly.AddXY(1.5, 0)
	poly.AddXY(barWidth/2, 0)
	poly.AddXY(headWidth/2, -headHeight)
	poly.AddXY(headWidth/2, -(headHeight + headRear))
	poly.AddXY(barWidth/2, -(headHeight + headRear))
	poly.AddXY(barWidth/2, -(headHeight + headRear + headNeck))
	poly.AddXY(headWidth/2, -(headHeight + headRear*2 + headNeck))
	poly.AddXY(headWidth/2, -(insideClip + outsideClip - baseHeight - headRear))
	poly.AddXY((headWidth/2)+headRear, -(insideClip + outsideClip - baseHeight))
	poly.AddXY(baseWidth/2, -(insideClip + outsideClip - baseHeight))
	poly.AddXY(baseWidth/2, -(insideClip + outsideClip))
	poly.AddXY(0, -(insideClip + outsideClip))
	poly.AddXY(0, -0.01)

	// poly.AddXY(-baseWidth/2, -(insideClip + outsideClip))
	// poly.AddXY(-baseWidth/2, -(insideClip + outsideClip - baseHeight))
	// poly.AddXY(-barWidth/2, -(insideClip + outsideClip - baseHeight))
	// poly.AddXY(-barWidth/2, -(insideClip + outsideClip))
	// poly.AddXY(-barWidth/2, -(headHeight + headRear))
	// poly.AddXY(-headWidth/2, -(headHeight + headRear))
	// poly.AddXY(-headWidth/2, -headHeight)
	// poly.AddXY(-barWidth/2, 0)
	// poly.AddXY(0, 0)
	vertices, err := poly.AppendVecs(nil)
	if err != nil {
		return nil, err
	}
	halfHeadOutline, err := gsdf.NewPolygon(vertices)
	if err != nil {
		panic("failed new polygon creation")
	}
	obj, err := gsdf.Extrude(halfHeadOutline, barHeight)
	if err != nil {
		panic("failed Extrude")
	}
	mirrorHeadOutline := gsdf.Symmetry(obj, true, false, false)
	// Squash two objects together a fraction
	obj = gsdf.Translate(obj, -0.05, 0, 0)
	mirrorHeadOutline = gsdf.Translate(mirrorHeadOutline, 0.05, 0, 0)
	headOutline := gsdf.Union(mirrorHeadOutline, obj)
	// obj2 := gsdf.Union(obj, bar)
	pin, err := makeSDF(headOutline)

	return pin, err
}

func makeSDF(s glbuild.Shader3D) (gleval.SDF3, error) {
	err := glbuild.RewriteNames3D(&s, 32) // Shorten names to not crash GL tokenizer.
	if err != nil {
		return nil, err
	}
	return gleval.NewCPUSDF3(s)
}
func main() {
	b, _ := BrabantiaPin()

	const resDiv = 200
	const evaluationBufferSize = 1024 * 8
	resolution := b.Bounds().Size().Max() / resDiv
	renderer, err := glrender.NewOctreeRenderer(b, resolution, evaluationBufferSize)
	if err != nil {
		fmt.Println("error creating renderer:", err)
		os.Exit(1)
	}
	start := time.Now()
	triangles, err := glrender.RenderAll(renderer)
	if err != nil {
		fmt.Println("error rendering triangles:", err)
		os.Exit(1)
	}
	elapsed := time.Since(start)
	evals := b.(interface{ Evaluations() uint64 }).Evaluations()

	fp, err := os.Create("/mnt/c/t/brabantia_pin.stl")
	if err != nil {
		fmt.Println("error creating file:", err)
		os.Exit(1)
	}
	defer fp.Close()
	start = time.Now()
	w := bufio.NewWriter(fp)
	_, err = glrender.WriteBinarySTL(w, triangles)
	if err != nil {
		fmt.Println("error writing triangles to file:", err)
		os.Exit(1)
	}
	w.Flush()
	fmt.Println("SDF created in ", elapsed, "evaluated sdf", evals, "times, rendered", len(triangles), "triangles in", elapsed, "wrote file in", time.Since(start))

}
