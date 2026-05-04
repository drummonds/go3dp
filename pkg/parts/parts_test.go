package parts

import (
	"testing"

	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
	"github.com/soypat/gsdf/gleval"
)

// stubPart is a unit cube centred on the origin. Insert is a slightly
// larger box (Radial mm bigger on each side, Axial mm taller in Z).
// Used to test the assembly helpers without dragging in fasteners.
type stubPart struct{}

func (stubPart) Shape(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	return bld.NewBox(1, 1, 1, 0), bld.Err()
}

func (stubPart) Insert(bld *gsdf.Builder, tol Tolerance) (glbuild.Shader3D, error) {
	return bld.NewBox(1+tol.Radial, 1+tol.Radial, 1+tol.Axial, 0), bld.Err()
}

func evalAt(t *testing.T, shape glbuild.Shader3D, points ...ms3.Vec) []float32 {
	t.Helper()
	cpu, err := gleval.NewCPUSDF3(shape)
	if err != nil {
		t.Fatalf("NewCPUSDF3: %v", err)
	}
	dist := make([]float32, len(points))
	if err := cpu.Evaluate(points, dist, nil); err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return dist
}

func TestComposite_HostMinusInsertPlusPart(t *testing.T) {
	var bld gsdf.Builder
	host := bld.NewBox(10, 10, 10, 0) // big block, centred at origin
	inserts := []Placement{
		{Part: stubPart{}, Offset: ms3.Vec{X: 0, Y: 0, Z: 0}, Tolerance: Tolerance{Radial: 0.5, Axial: 0.5}},
	}
	parts := []Placement{
		{Part: stubPart{}, Offset: ms3.Vec{X: 4, Y: 0, Z: 0}}, // sit a part outside the host
	}
	out, err := Composite(&bld, host, parts, inserts)
	if err != nil {
		t.Fatalf("Composite: %v", err)
	}
	if err := bld.Err(); err != nil {
		t.Fatal(err)
	}
	// Sampling guide:
	//  - host spans [-5,5]
	//  - insert hole spans [-0.75, 0.75] (1.5 mm cube at origin, since
	//    Tolerance grows the unit cube by Radial on each side)
	//  - unioned part spans [3.5, 4.5] in x (unit cube at offset 4)
	d := evalAt(t, out,
		ms3.Vec{X: 0, Y: 0, Z: 0}, // origin: in host AND in insert hole → void
		ms3.Vec{X: 1, Y: 0, Z: 0}, // in host, outside insert hole → solid
		ms3.Vec{X: 4, Y: 0, Z: 0}, // centre of unioned part → solid
		ms3.Vec{X: 6, Y: 0, Z: 0}, // outside host, outside part → void
	)
	if d[0] <= 0 {
		t.Errorf("(0,0,0) inside insert hole: dist=%g, want >0 (void)", d[0])
	}
	if d[1] >= 0 {
		t.Errorf("(1,0,0) host, outside hole: dist=%g, want <0 (solid)", d[1])
	}
	if d[2] >= 0 {
		t.Errorf("(4,0,0) inside unioned part: dist=%g, want <0 (solid)", d[2])
	}
	if d[3] <= 0 {
		t.Errorf("(6,0,0) outside everything: dist=%g, want >0 (void)", d[3])
	}
}

func TestExploded_TranslatesAlongAxis(t *testing.T) {
	var bld gsdf.Builder
	host := bld.NewBox(2, 2, 2, 0) // small reference at origin
	placements := []Placement{
		{Part: stubPart{}}, // i=0 → +5 along axis
		{Part: stubPart{}}, // i=1 → +10 along axis
	}
	out, err := Exploded(&bld, host, placements, ms3.Vec{Z: 1}, 5)
	if err != nil {
		t.Fatalf("Exploded: %v", err)
	}
	if err := bld.Err(); err != nil {
		t.Fatal(err)
	}
	d := evalAt(t, out,
		ms3.Vec{X: 0, Y: 0, Z: 0},  // host
		ms3.Vec{X: 0, Y: 0, Z: 5},  // first part at z=+5
		ms3.Vec{X: 0, Y: 0, Z: 10}, // second part at z=+10
		ms3.Vec{X: 0, Y: 0, Z: 7},  // gap between parts
	)
	if d[0] >= 0 {
		t.Errorf("(0,0,0) host centre: dist=%g, want <0 (solid)", d[0])
	}
	if d[1] >= 0 {
		t.Errorf("(0,0,5) first exploded part: dist=%g, want <0 (solid)", d[1])
	}
	if d[2] >= 0 {
		t.Errorf("(0,0,10) second exploded part: dist=%g, want <0 (solid)", d[2])
	}
	if d[3] <= 0 {
		t.Errorf("(0,0,7) gap between parts: dist=%g, want >0 (void)", d[3])
	}
}
