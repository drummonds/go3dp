package svgslice

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/chewxy/math32"
	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
)

var update = flag.Bool("update", false, "regenerate golden files under testdata/")

// testDate keeps snapshot output deterministic — the footer would
// otherwise interpolate today's date on every run.
const testDate = "2026-05-04"

// sphereSDF is a closed-form gleval.SDF3 implementation used by the tests to
// avoid pulling in the full gsdf builder pipeline.
type sphereSDF struct {
	centre ms3.Vec
	radius float32
}

func (s sphereSDF) Evaluate(pos []ms3.Vec, dist []float32, _ any) error {
	for i, p := range pos {
		dx := p.X - s.centre.X
		dy := p.Y - s.centre.Y
		dz := p.Z - s.centre.Z
		dist[i] = math32.Sqrt(dx*dx+dy*dy+dz*dz) - s.radius
	}
	return nil
}

func (s sphereSDF) Bounds() ms3.Box {
	r := s.radius
	return ms3.Box{
		Min: ms3.Vec{X: s.centre.X - r, Y: s.centre.Y - r, Z: s.centre.Z - r},
		Max: ms3.Vec{X: s.centre.X + r, Y: s.centre.Y + r, Z: s.centre.Z + r},
	}
}

// bracketSDF is an asymmetric (60×30×40 mm) box with two through-holes:
// a Z-axis cylinder (r=8) and an X-axis cylinder (r=5). Each principal-
// plane slice shows different combinations of these holes — exercises
// asymmetric output, internal contour loops (fill-rule="evenodd"), and
// auto-dimension extents.
type bracketSDF struct{}

func (bracketSDF) Bounds() ms3.Box {
	return ms3.Box{
		Min: ms3.Vec{X: -30, Y: -15, Z: -20},
		Max: ms3.Vec{X: 30, Y: 15, Z: 20},
	}
}

func (bracketSDF) Evaluate(pos []ms3.Vec, dist []float32, _ any) error {
	const (
		hx, hy, hz = float32(30), float32(15), float32(20)
		rZHole     = float32(8) // through-hole along Z
		rXHole     = float32(5) // through-hole along X
	)
	for i, p := range pos {
		qx := math32.Abs(p.X) - hx
		qy := math32.Abs(p.Y) - hy
		qz := math32.Abs(p.Z) - hz
		ox := math32.Max(qx, 0)
		oy := math32.Max(qy, 0)
		oz := math32.Max(qz, 0)
		out := math32.Sqrt(ox*ox + oy*oy + oz*oz)
		in := math32.Min(math32.Max(qx, math32.Max(qy, qz)), 0)
		d := out + in
		// Subtract Z-axis cylinder.
		dZ := math32.Sqrt(p.X*p.X+p.Y*p.Y) - rZHole
		if -dZ > d {
			d = -dZ
		}
		// Subtract X-axis cylinder.
		dX := math32.Sqrt(p.Y*p.Y+p.Z*p.Z) - rXHole
		if -dX > d {
			d = -dX
		}
		dist[i] = d
	}
	return nil
}

func TestWriteSlice_Sphere(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	plane := XZ(8, 8, 8)
	out := filepath.Join(t.TempDir(), "sphere.svg")

	if err := WriteSliceOpt(out, sdf, plane, Options{GridX: 48, GridY: 48, Date: testDate}, nil); err != nil {
		t.Fatalf("WriteSliceOpt: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	s := string(data)
	for _, want := range []string{
		`<?xml version="1.0"`,
		`<svg `,
		`viewBox=`,
		`<path `,
		` Z`, // closed loop terminator
		`</svg>`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("output missing %q", want)
		}
	}
}

func TestWriteSliceOpt_AutoGrid(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	out := filepath.Join(t.TempDir(), "auto.svg")
	if err := WriteSliceOpt(out, sdf, XZ(8, 8, 8), Options{}, nil); err != nil {
		t.Fatalf("WriteSliceOpt: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `width="378"`) {
		t.Errorf("default DisplayPx should produce width=378, got first line:\n%s",
			strings.SplitN(string(data), "\n", 3)[1])
	}
}

func TestWriteSliceOpt_AbsoluteUnits(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	out := filepath.Join(t.TempDir(), "abs.svg")
	opts := Options{AbsoluteUnits: true, GridX: 48, GridY: 48}
	if err := WriteSliceOpt(out, sdf, XZ(8, 8, 8), opts, nil); err != nil {
		t.Fatalf("WriteSliceOpt: %v", err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `width="26mm"`) {
		t.Errorf("AbsoluteUnits should emit mm-based width")
	}
}

func TestWriteSliceOpt_HideAxisGizmo(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	plane := XZ(8, 8, 8)

	withGizmo := filepath.Join(t.TempDir(), "with.svg")
	if err := WriteSliceOpt(withGizmo, sdf, plane, Options{GridX: 48, GridY: 48}, nil); err != nil {
		t.Fatal(err)
	}
	hidden := filepath.Join(t.TempDir(), "hidden.svg")
	if err := WriteSliceOpt(hidden, sdf, plane, Options{GridX: 48, GridY: 48, HideAxisGizmo: true}, nil); err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(withGizmo)
	b, _ := os.ReadFile(hidden)
	// The gizmo emits the X-axis colour (#dc2626) in an XZ slice.
	if !strings.Contains(string(a), "#dc2626") {
		t.Errorf("default output should include the axis gizmo (#dc2626 missing)")
	}
	if strings.Contains(string(b), "#dc2626") {
		t.Errorf("HideAxisGizmo should suppress the gizmo (#dc2626 still present)")
	}
}

func TestWriteSliceOpt_Title(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	out := filepath.Join(t.TempDir(), "title.svg")
	opts := Options{GridX: 48, GridY: 48, Date: testDate, Title: "test title"}
	if err := WriteSliceOpt(out, sdf, XZ(8, 8, 8), opts, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "test title") {
		t.Error("Title not present in footer block")
	}
}

func TestWriteSliceOpt_AutoDimensions(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	out := filepath.Join(t.TempDir(), "autodim.svg")
	opts := Options{GridX: 48, GridY: 48, Date: testDate, AutoDimensions: true}
	if err := WriteSliceOpt(out, sdf, XZ(8, 8, 8), opts, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(out)
	s := string(data)
	// Two auto-dim labels (horiz + vert) should appear, formatted as bare
	// numbers (~10.0 for a r=5 sphere). The unit lives in the title block.
	if strings.Count(s, ">9.9</text>") < 2 && strings.Count(s, ">10.0</text>") < 2 {
		t.Errorf("expected two ~10 dim labels (horiz + vert), got:\n%s",
			extractDimLabels(s))
	}
}

func extractDimLabels(svg string) string {
	var out []string
	for line := range strings.SplitSeq(svg, "\n") {
		if strings.Contains(line, "fill=\"#1e40af\"") && strings.Contains(line, "<text ") {
			out = append(out, strings.TrimSpace(line))
		}
	}
	return strings.Join(out, "\n")
}

func TestBracket_ThreeCutaways(t *testing.T) {
	sdf := bracketSDF{}
	cases := []struct {
		name  string
		plane Plane
		title string
	}{
		{
			name: "bracket_xz.svg",
			plane: Plane{
				U: ms3.Vec{X: 1}, V: ms3.Vec{Z: 1},
				UMin: -42, UMax: 42, VMin: -32, VMax: 32,
			},
			title: "bracket front (XZ)",
		},
		{
			name: "bracket_xy.svg",
			plane: Plane{
				U: ms3.Vec{X: 1}, V: ms3.Vec{Y: 1},
				UMin: -42, UMax: 42, VMin: -27, VMax: 27,
			},
			title: "bracket top (XY)",
		},
		{
			name: "bracket_yz.svg",
			plane: Plane{
				U: ms3.Vec{Y: 1}, V: ms3.Vec{Z: 1},
				UMin: -27, UMax: 27, VMin: -32, VMax: 32,
			},
			title: "bracket side (YZ)",
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			opts := Options{
				GridX: 96, GridY: 96,
				Date:           testDate,
				Title:          c.title,
				AutoDimensions: true,
			}
			checkSnapshot(t, c.name, func(path string) error {
				return WriteSliceOpt(path, sdf, c.plane, opts, nil)
			})
		})
	}
}

func TestWriteSliceOpt_HideFooter(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	plane := XZ(8, 8, 8)

	withFooter := filepath.Join(t.TempDir(), "with.svg")
	if err := WriteSliceOpt(withFooter, sdf, plane,
		Options{GridX: 48, GridY: 48, Date: testDate}, nil); err != nil {
		t.Fatal(err)
	}
	hidden := filepath.Join(t.TempDir(), "hidden.svg")
	if err := WriteSliceOpt(hidden, sdf, plane,
		Options{GridX: 48, GridY: 48, Date: testDate, HideFooter: true}, nil); err != nil {
		t.Fatal(err)
	}
	a, _ := os.ReadFile(withFooter)
	b, _ := os.ReadFile(hidden)
	if !strings.Contains(string(a), testDate) || !strings.Contains(string(a), "mm") {
		t.Errorf("default output should include the footer (units + date)")
	}
	if strings.Contains(string(b), testDate) {
		t.Errorf("HideFooter should suppress the footer date")
	}
}

func TestWriteSlice_GridTooSmall(t *testing.T) {
	sdf := sphereSDF{radius: 1}
	out := filepath.Join(t.TempDir(), "tiny.svg")
	if err := WriteSliceOpt(out, sdf, XZ(2, 2, 2), Options{GridX: 1, GridY: 4}, nil); err == nil {
		t.Fatal("expected error for gridX<2, got nil")
	}
}

// TestWriteSlice_Snapshot is a regression check against a checked-in golden
// SVG. Run `go test -update` to regenerate after intentional changes.
func TestWriteSlice_Snapshot(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	plane := XZ(8, 8, 8)
	opts := Options{GridX: 48, GridY: 48, Date: testDate}
	checkSnapshot(t, "sphere_xz.svg", func(path string) error {
		return WriteSliceOpt(path, sdf, plane, opts, nil)
	})
}

func TestWriteSlice_LargeSphereSnapshot(t *testing.T) {
	sdf := sphereSDF{radius: 250}
	plane := XZ(400, 400, 400)
	opts := Options{
		GridX: 48, GridY: 48,
		Date:           testDate,
		Title:          "sphere r=250",
		AutoDimensions: true,
	}
	checkSnapshot(t, "sphere_xz_500mm.svg", func(path string) error {
		return WriteSliceOpt(path, sdf, plane, opts, nil)
	})
}

func TestWriteSlice_LabelledSnapshot(t *testing.T) {
	sdf := sphereSDF{radius: 5}
	plane := XZ(8, 8, 8)
	labels := []Label{
		{U: 0, V: 6.5, Text: "sphere r=5", Anchor: "middle"},
	}
	dim := Dimension{
		From:    ms2.Vec{X: -5, Y: 0},
		To:      ms2.Vec{X: 5, Y: 0},
		DimFrom: ms2.Vec{X: -5, Y: -7},
		DimTo:   ms2.Vec{X: 5, Y: -7},
	}
	opts := Options{GridX: 48, GridY: 48, Date: testDate}
	checkSnapshot(t, "sphere_xz_labelled.svg", func(path string) error {
		return WriteSliceOpt(path, sdf, plane, opts, labels, dim)
	})
}

func checkSnapshot(t *testing.T, name string, write func(path string) error) {
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
		t.Fatalf("read output: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("output differs from %s — re-run with `-update` if the change is intentional", golden)
	}
}
