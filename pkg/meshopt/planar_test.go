package meshopt

import (
	"testing"

	"github.com/soypat/geometry/ms3"
)

// A unit square on z=0, subdivided into a 4×4 grid of cells (32 triangles)
// must collapse to 2 triangles.
func TestPlanarMerge_SubdividedSquare(t *testing.T) {
	const N = 4
	var tris []ms3.Triangle
	step := float32(1.0) / N
	for i := 0; i < N; i++ {
		for j := 0; j < N; j++ {
			x0 := float32(i) * step
			y0 := float32(j) * step
			x1 := x0 + step
			y1 := y0 + step
			a := ms3.Vec{X: x0, Y: y0}
			b := ms3.Vec{X: x1, Y: y0}
			c := ms3.Vec{X: x1, Y: y1}
			d := ms3.Vec{X: x0, Y: y1}
			tris = append(tris,
				ms3.Triangle{a, b, c},
				ms3.Triangle{a, c, d},
			)
		}
	}
	if len(tris) != 32 {
		t.Fatalf("setup: want 32 tris, got %d", len(tris))
	}
	out := PlanarMerge(tris, Options{})
	if len(out) != 2 {
		t.Errorf("want 2 triangles after merge, got %d", len(out))
	}
}

// 12-tri axis-aligned cube must remain a 12-tri cube (each face is already
// only two triangles around a 4-vertex loop → 2 ear-clip output).
func TestPlanarMerge_Cube(t *testing.T) {
	cube := unitCubeTriangles()
	if len(cube) != 12 {
		t.Fatalf("setup: want 12 tris, got %d", len(cube))
	}
	out := PlanarMerge(cube, Options{})
	if len(out) != 12 {
		t.Errorf("want 12 triangles after merge, got %d", len(out))
	}
}

func unitCubeTriangles() []ms3.Triangle {
	v := [8]ms3.Vec{
		{X: 0, Y: 0, Z: 0},
		{X: 1, Y: 0, Z: 0},
		{X: 1, Y: 1, Z: 0},
		{X: 0, Y: 1, Z: 0},
		{X: 0, Y: 0, Z: 1},
		{X: 1, Y: 0, Z: 1},
		{X: 1, Y: 1, Z: 1},
		{X: 0, Y: 1, Z: 1},
	}
	face := func(a, b, c, d int) []ms3.Triangle {
		return []ms3.Triangle{{v[a], v[b], v[c]}, {v[a], v[c], v[d]}}
	}
	var t []ms3.Triangle
	t = append(t, face(0, 3, 2, 1)...) // -Z
	t = append(t, face(4, 5, 6, 7)...) // +Z
	t = append(t, face(0, 1, 5, 4)...) // -Y
	t = append(t, face(2, 3, 7, 6)...) // +Y
	t = append(t, face(1, 2, 6, 5)...) // +X
	t = append(t, face(0, 4, 7, 3)...) // -X
	return t
}
