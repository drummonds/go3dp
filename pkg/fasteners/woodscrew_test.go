package fasteners

import (
	"flag"
	"os"
	"path/filepath"
	"testing"

	"codeberg.org/hum3/go3dp/pkg/svgslice"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
	"github.com/soypat/gsdf/gleval"
)

var update = flag.Bool("update", false, "regenerate testdata/ snapshots")

const testDate = "2026-05-04"

// evalAt builds a CPU evaluator from shape and returns distances at the
// given points.
func evalAt(t *testing.T, shape glbuild.Shader3D, points ...ms3.Vec) []float32 {
	t.Helper()
	cpusdf, err := gleval.NewCPUSDF3(shape)
	if err != nil {
		t.Fatalf("NewCPUSDF3: %v", err)
	}
	dist := make([]float32, len(points))
	if err := cpusdf.Evaluate(points, dist, nil); err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	return dist
}

func TestSchematic_NoPlug(t *testing.T) {
	var bld gsdf.Builder
	shape, err := Spax4x20.Schematic(&bld, 0)
	if err != nil {
		t.Fatalf("Schematic: %v", err)
	}
	if err := bld.Err(); err != nil {
		t.Fatalf("builder: %v", err)
	}
	// Bounds: head face at z=0, tip at z=-OverallLength=-20.
	bb := shape.Bounds()
	if bb.Max.Z > 0.001 || bb.Max.Z < -0.001 {
		t.Errorf("no-plug head face: bounds.Max.Z = %g, want ~0", bb.Max.Z)
	}
	if bb.Min.Z > -19.999 || bb.Min.Z < -20.001 {
		t.Errorf("tip Z bound: bounds.Min.Z = %g, want ~-20", bb.Min.Z)
	}

	// Sample: well above head should be outside; on axis just below
	// head face should be inside the shank.
	d := evalAt(t, shape,
		ms3.Vec{X: 0, Y: 0, Z: 5},   // 5mm above head face — no plug → outside
		ms3.Vec{X: 0, Y: 0, Z: -5},  // inside shank
		ms3.Vec{X: 0, Y: 0, Z: -25}, // below tip → outside
	)
	if d[0] <= 0 {
		t.Errorf("(0,0,5) above head, no plug: dist=%g, want >0 (outside)", d[0])
	}
	if d[1] >= 0 {
		t.Errorf("(0,0,-5) in shank: dist=%g, want <0 (inside)", d[1])
	}
	if d[2] <= 0 {
		t.Errorf("(0,0,-25) below tip: dist=%g, want >0 (outside)", d[2])
	}
}

func TestSchematic_PlugFillsAboveHead(t *testing.T) {
	var bld gsdf.Builder
	shape, err := Spax4x20.Schematic(&bld, 10)
	if err != nil {
		t.Fatalf("Schematic: %v", err)
	}
	if err := bld.Err(); err != nil {
		t.Fatalf("builder: %v", err)
	}
	bb := shape.Bounds()
	if bb.Max.Z < 9.999 || bb.Max.Z > 10.001 {
		t.Errorf("plug top: bounds.Max.Z = %g, want ~10", bb.Max.Z)
	}

	// On axis at z=5 (inside the plug) — should now be inside the screw shape.
	// Just outside the plug rim (radius DHead/2 = 4 → sample at r=4.5) — outside.
	d := evalAt(t, shape,
		ms3.Vec{X: 0, Y: 0, Z: 5},   // mid-plug, on axis → inside
		ms3.Vec{X: 0, Y: 0, Z: 9.5}, // near plug top → inside
		ms3.Vec{X: 4.5, Y: 0, Z: 5}, // outside plug radius → outside
		ms3.Vec{X: 0, Y: 0, Z: 11},  // above plug → outside
	)
	if d[0] >= 0 {
		t.Errorf("(0,0,5) mid-plug: dist=%g, want <0", d[0])
	}
	if d[1] >= 0 {
		t.Errorf("(0,0,9.5) near plug top: dist=%g, want <0", d[1])
	}
	if d[2] <= 0 {
		t.Errorf("(4.5,0,5) outside plug rim: dist=%g, want >0", d[2])
	}
	if d[3] <= 0 {
		t.Errorf("(0,0,11) above plug: dist=%g, want >0", d[3])
	}
}

// TestSchematic_BlockCutaway is the canonical "working model" test:
// drive a Spax4x20 into a 30×30×30 mm block centred at the origin so the
// head face sits 10 mm below the top, then subtract the schematic (with a
// 10 mm plug) from the block. Verify the resulting void includes the
// screwdriver-access column above the head.
func TestSchematic_BlockCutaway(t *testing.T) {
	var bld gsdf.Builder

	// Block 30×30×30, centred at origin → z range -15..15.
	block := bld.NewBox(30, 30, 30, 0)

	// Screw with a 10 mm plug. Place head face at z=5 so the plug top
	// reaches z=15 (the block's top face); tip lands at z=-15 (the
	// block's bottom face exactly).
	screw, err := Spax4x20.Schematic(&bld, 10)
	if err != nil {
		t.Fatalf("Schematic: %v", err)
	}
	screw = bld.Translate(screw, 0, 0, 5)

	cutaway := bld.Difference(block, screw)
	if err := bld.Err(); err != nil {
		t.Fatalf("builder: %v", err)
	}

	d := evalAt(t, cutaway,
		// Inside the block away from the screw: still solid.
		ms3.Vec{X: 12, Y: 12, Z: 0},
		// Inside the access column above the head: void.
		ms3.Vec{X: 0, Y: 0, Z: 10},
		// Inside the shank hole below the head: void.
		ms3.Vec{X: 0, Y: 0, Z: 0},
		// At the outer face of the block, screw axis: void (plug surface).
		ms3.Vec{X: 0, Y: 0, Z: 14.9},
		// Outside the block entirely.
		ms3.Vec{X: 20, Y: 0, Z: 0},
	)
	if d[0] >= 0 {
		t.Errorf("(12,12,0) in block away from screw: dist=%g, want <0 (solid)", d[0])
	}
	if d[1] <= 0 {
		t.Errorf("(0,0,10) in access column: dist=%g, want >0 (void)", d[1])
	}
	if d[2] <= 0 {
		t.Errorf("(0,0,0) in shank hole: dist=%g, want >0 (void)", d[2])
	}
	if d[3] <= 0 {
		t.Errorf("(0,0,14.9) at top of access column: dist=%g, want >0 (void)", d[3])
	}
	if d[4] <= 0 {
		t.Errorf("(20,0,0) outside block: dist=%g, want >0 (outside)", d[4])
	}
}

func TestSchematic_NegativePlugLenRejected(t *testing.T) {
	var bld gsdf.Builder
	if _, err := Spax4x20.Schematic(&bld, -1); err == nil {
		t.Error("expected error for negative plugLen, got nil")
	}
}

// TestDemoScrew_Snapshot renders two Spax 4×20 schematics side-by-side
// — bare schematic on the left, with a 10 mm access-column plug on the
// right — so the plug feature is visually obvious in the test data.
func TestDemoScrew_Snapshot(t *testing.T) {
	var bld gsdf.Builder
	bare, err := Spax4x20.Schematic(&bld, 0)
	if err != nil {
		t.Fatalf("Schematic bare: %v", err)
	}
	bare = bld.Translate(bare, -12, 0, 0)
	plugged, err := Spax4x20.Schematic(&bld, 10)
	if err != nil {
		t.Fatalf("Schematic plugged: %v", err)
	}
	plugged = bld.Translate(plugged, 12, 0, 0)
	pair := bld.Union(bare, plugged)
	if err := bld.Err(); err != nil {
		t.Fatalf("builder: %v", err)
	}
	cpusdf, err := gleval.NewCPUSDF3(pair)
	if err != nil {
		t.Fatalf("NewCPUSDF3: %v", err)
	}

	plane := svgslice.Plane{
		Origin: ms3.Vec{},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Z: 1},
		UMin:   -25, UMax: 25,
		VMin: -25, VMax: 15,
	}
	opts := svgslice.Options{
		GridX: 200, GridY: 160,
		Title:          "Spax 4×20 — bare (left) vs. with 10 mm plug (right)",
		Date:           testDate,
		AutoDimensions: true,
	}
	checkFastenerSnapshot(t, "demo_screw.svg", func(path string) error {
		return svgslice.WriteSliceOpt(path, cpusdf, plane, opts, nil)
	})
}

// TestDemoBlockWithScrew_Snapshot is the canonical "working model"
// visual: a 30 mm cube with the Spax 4×20 schematic (plus 10 mm plug)
// subtracted, sliced through the screw axis. The plug carves the
// screwdriver-access column above the head in one boolean.
func TestDemoBlockWithScrew_Snapshot(t *testing.T) {
	var bld gsdf.Builder
	block := bld.NewBox(30, 30, 30, 0)
	screw, err := Spax4x20.Schematic(&bld, 10)
	if err != nil {
		t.Fatalf("Schematic: %v", err)
	}
	screw = bld.Translate(screw, 0, 0, 5)
	cutaway := bld.Difference(block, screw)
	if err := bld.Err(); err != nil {
		t.Fatalf("builder: %v", err)
	}
	cpusdf, err := gleval.NewCPUSDF3(cutaway)
	if err != nil {
		t.Fatalf("NewCPUSDF3: %v", err)
	}

	plane := svgslice.Plane{
		Origin: ms3.Vec{},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Z: 1},
		UMin:   -50, UMax: 50,
		VMin: -50, VMax: 50,
	}
	opts := svgslice.Options{
		GridX: 256, GridY: 256,
		Title:          "30 mm block + Spax 4×20 cutaway",
		Date:           testDate,
		AutoDimensions: true,
	}
	checkFastenerSnapshot(t, "demo_block_with_screw.svg", func(path string) error {
		return svgslice.WriteSliceOpt(path, cpusdf, plane, opts, nil)
	})
}

func checkFastenerSnapshot(t *testing.T, name string, write func(path string) error) {
	t.Helper()
	golden := filepath.Join("testdata", name)
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatalf("mkdir testdata: %v", err)
		}
		if err := write(golden); err != nil {
			t.Fatalf("write golden: %v", err)
		}
		t.Logf("updated %s", golden)
		return
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatalf("read golden %s: %v (run `go test -update` to create it)", golden, err)
	}
	out := filepath.Join(t.TempDir(), name)
	if err := write(out); err != nil {
		t.Fatalf("write: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Errorf("output differs from %s — re-run with `-update` if intentional", golden)
	}
}

func TestWoodScrew_ValidationRejectsBadInputs(t *testing.T) {
	cases := []struct {
		name string
		s    WoodScrew
	}{
		{"zero shank diameter", WoodScrew{DHead: 5, HeadDepth: 1, OverallLength: 10}},
		{"head not bigger than shank", WoodScrew{DShank: 4, DHead: 4, HeadDepth: 1, OverallLength: 10}},
		{"zero head depth", WoodScrew{DShank: 4, DHead: 8, OverallLength: 10}},
		{"overall too short", WoodScrew{DShank: 4, DHead: 8, HeadDepth: 5, PointLength: 5, OverallLength: 9}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var bld gsdf.Builder
			if _, err := c.s.Schematic(&bld, 0); err == nil {
				t.Errorf("expected validation error, got nil")
			}
		})
	}
}
