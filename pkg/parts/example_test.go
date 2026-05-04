package parts

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

const exampleDate = "2026-05-04"

// Worked example: a 30 mm cube with a Ø8 mm peg through the centre.
// The peg is the "external" Part (e.g. a steel rod); the block is the
// "built" Part being designed around it. We use an exaggerated 2 mm
// radial tolerance so the gap is clearly visible in the SVG.
const (
	exBlockSide   = 30.0
	exPegRadius   = 4.0
	exPegHeight   = 36.0 // > block side, so the peg passes fully through
	exExplodeStep = 40.0 // peg lifted this far above its assembly position in stage 1
	// At step=40 the peg insert (H≈37) spans Z∈[21.5,58.5], leaving a
	// ~6.5 mm gap above the block (Z∈[-15,15]) so the two shapes
	// render as distinct silhouettes rather than a merged T.
)

var exampleTolerance = Tolerance{Radial: 2.0, Axial: 1.0}

func buildStage1Exploded(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	block, err := (Block{Lx: exBlockSide, Ly: exBlockSide, Lz: exBlockSide}).Shape(bld)
	if err != nil {
		return nil, err
	}
	peg := Cylinder{R: exPegRadius, H: exPegHeight}
	pegInsert, err := peg.Insert(bld, exampleTolerance)
	if err != nil {
		return nil, err
	}
	pegInsert = bld.Translate(pegInsert, 0, 0, exExplodeStep)
	return bld.Union(block, pegInsert), bld.Err()
}

func buildStage2Together(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	block, err := (Block{Lx: exBlockSide, Ly: exBlockSide, Lz: exBlockSide}).Shape(bld)
	if err != nil {
		return nil, err
	}
	peg := Cylinder{R: exPegRadius, H: exPegHeight}
	return Composite(bld, block,
		[]Placement{{Part: peg}},
		[]Placement{{Part: peg, Tolerance: exampleTolerance}},
	)
}

func buildStage3Final(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	block, err := (Block{Lx: exBlockSide, Ly: exBlockSide, Lz: exBlockSide}).Shape(bld)
	if err != nil {
		return nil, err
	}
	peg := Cylinder{R: exPegRadius, H: exPegHeight}
	return Composite(bld, block, nil,
		[]Placement{{Part: peg, Tolerance: exampleTolerance}},
	)
}

// TestPartsExample_ThreeStages_Snapshot generates the three SVG cutaways
// referenced from pkg/parts/README.md: exploded → together → final.
// All three use the same 100×100 mm plane for visual consistency.
func TestPartsExample_ThreeStages_Snapshot(t *testing.T) {
	plane := svgslice.Plane{
		U: ms3.Vec{X: 1}, V: ms3.Vec{Z: 1},
		UMin: -50, UMax: 50,
		VMin: -30, VMax: 70,
	}
	cases := []struct {
		name  string
		title string
		build func(*gsdf.Builder) (glbuild.Shader3D, error)
	}{
		{"example_stage1_exploded.svg", "stage 1 — exploded (with insert form)", buildStage1Exploded},
		{"example_stage2_together.svg", "stage 2 — assembled (insert + true peg)", buildStage2Together},
		{"example_stage3_final.svg", "stage 3 — final printed block", buildStage3Final},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var bld gsdf.Builder
			shape, err := c.build(&bld)
			if err != nil {
				t.Fatalf("build: %v", err)
			}
			if err := bld.Err(); err != nil {
				t.Fatalf("builder: %v", err)
			}
			cpu, err := gleval.NewCPUSDF3(shape)
			if err != nil {
				t.Fatalf("NewCPUSDF3: %v", err)
			}
			opts := svgslice.Options{
				GridX: 256, GridY: 256,
				Title:          c.title,
				Date:           exampleDate,
				AutoDimensions: true,
			}
			checkExampleSnapshot(t, c.name, func(path string) error {
				return svgslice.WriteSliceOpt(path, cpu, plane, opts, nil)
			})
		})
	}
}

func checkExampleSnapshot(t *testing.T, name string, write func(path string) error) {
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
