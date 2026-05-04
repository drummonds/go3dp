// Package meshopt provides post-meshing optimisations for triangle soup.
//
// PlanarMerge groups connected coplanar triangles, replaces each group's
// boundary loop with a fresh ear-clip triangulation, and emits the union.
// On a meshed cube each face's many MC/DC triangles collapse to two —
// roughly N-2 output triangles per N-vertex boundary loop.
package meshopt

import (
	"math"

	"github.com/chewxy/math32"
	"github.com/soypat/geometry/ms3"
)

type Options struct {
	// VertexTol is the weld distance for snapping near-coincident vertices.
	// Default 1e-5 (units = mesh units, typically mm).
	VertexTol float32
	// NormalTol is the cosine threshold above which two unit normals are
	// considered the same direction. Default 0.9998 (~1.15°).
	NormalTol float32
	// OffsetTol is the absolute distance two parallel planes may differ by
	// to be considered the same plane. Default 1e-4.
	OffsetTol float32
}

func (o Options) withDefaults() Options {
	if o.VertexTol == 0 {
		o.VertexTol = 1e-5
	}
	if o.NormalTol == 0 {
		o.NormalTol = 0.9998
	}
	if o.OffsetTol == 0 {
		o.OffsetTol = 1e-4
	}
	return o
}

// PlanarMerge merges connected coplanar triangle regions.
// Groups whose boundary is not a single simple loop are left untouched,
// so the result is always at least as large as the truly mergeable subset.
func PlanarMerge(tris []ms3.Triangle, opts Options) []ms3.Triangle {
	opts = opts.withDefaults()
	if len(tris) == 0 {
		return tris
	}

	// 1. Weld vertices into a unique-vertex array.
	verts, vidx := weldVertices(tris, opts.VertexTol)

	// 2. Per-triangle plane (unit normal + offset). Drop degenerate tris.
	td := make([]triData, len(tris))
	for i, t := range tris {
		v0, v1, v2 := vidx[3*i], vidx[3*i+1], vidx[3*i+2]
		if v0 == v1 || v1 == v2 || v2 == v0 {
			td[i].dead = true
			continue
		}
		nv := t.Normal()
		ln := ms3.Norm(nv)
		if ln == 0 {
			td[i].dead = true
			continue
		}
		n := ms3.Scale(1/ln, nv)
		td[i] = triData{
			v:   [3]int{v0, v1, v2},
			n:   n,
			d:   ms3.Dot(n, verts[v0]),
			key: keyFor(n, ms3.Dot(n, verts[v0]), opts),
		}
	}

	// 3. Half-edge map: directed edge (a,b) -> tri index that owns it.
	//    A coplanar neighbour is a triangle that owns the reverse edge (b,a)
	//    AND shares the same plane key.
	type heKey struct{ a, b int }
	half := make(map[heKey]int, len(tris)*3)
	for i, t := range td {
		if t.dead {
			continue
		}
		for k := 0; k < 3; k++ {
			a, b := t.v[k], t.v[(k+1)%3]
			half[heKey{a, b}] = i
		}
	}

	// 4. Union-find: union neighbouring triangles that share a half-edge
	//    and a plane key (with finer-grained verification).
	uf := newUF(len(td))
	for i, t := range td {
		if t.dead {
			continue
		}
		for k := 0; k < 3; k++ {
			a, b := t.v[k], t.v[(k+1)%3]
			j, ok := half[heKey{b, a}]
			if !ok {
				continue
			}
			tj := td[j]
			if tj.dead || tj.key != t.key {
				continue
			}
			if !planesMatch(t.n, t.d, tj.n, tj.d, opts) {
				continue
			}
			uf.union(i, j)
		}
	}

	// 5. Bucket triangles by component root.
	groups := make(map[int][]int)
	for i, t := range td {
		if t.dead {
			continue
		}
		r := uf.find(i)
		groups[r] = append(groups[r], i)
	}

	out := make([]ms3.Triangle, 0, len(tris))
	for _, members := range groups {
		if len(members) == 1 {
			i := members[0]
			t := td[i]
			out = append(out, ms3.Triangle{verts[t.v[0]], verts[t.v[1]], verts[t.v[2]]})
			continue
		}

		// Collect all directed half-edges of the group.
		edges := make(map[heKey]struct{}, len(members)*3)
		for _, i := range members {
			t := td[i]
			for k := 0; k < 3; k++ {
				edges[heKey{t.v[k], t.v[(k+1)%3]}] = struct{}{}
			}
		}
		// Boundary half-edges = those whose reverse is not present.
		var bnd []heKey
		next := make(map[int]int) // tail -> head; if a vertex has multiple outgoing boundary edges we bail.
		ambig := false
		for e := range edges {
			if _, ok := edges[heKey{e.b, e.a}]; ok {
				continue
			}
			bnd = append(bnd, e)
			if _, dup := next[e.a]; dup {
				ambig = true
				break
			}
			next[e.a] = e.b
		}
		if ambig || len(bnd) == 0 {
			emitOriginals(&out, td, verts, members)
			continue
		}

		// Walk the loop. Verify single closed loop covering all boundary edges.
		start := bnd[0].a
		loop := []int{start}
		cur := next[start]
		for cur != start {
			loop = append(loop, cur)
			nv, ok := next[cur]
			if !ok {
				ambig = true
				break
			}
			cur = nv
			if len(loop) > len(bnd)+1 {
				ambig = true
				break
			}
		}
		if ambig || len(loop) != len(bnd) {
			emitOriginals(&out, td, verts, members)
			continue
		}

		// Project loop into 2D using a basis on the plane.
		n := td[members[0]].n
		u, v := planeBasis(n)
		loop2 := make([][2]float32, len(loop))
		for i, vi := range loop {
			p := verts[vi]
			loop2[i] = [2]float32{ms3.Dot(p, u), ms3.Dot(p, v)}
		}
		// Drop collinear loop vertices — adjacent edges of merged faces
		// often subdivide the perimeter along seams shared with the
		// neighbouring plane's triangulation.
		loop, loop2 = dropCollinear(loop, loop2, opts.VertexTol)
		if len(loop) < 3 {
			emitOriginals(&out, td, verts, members)
			continue
		}

		// Ensure CCW winding in 2D so ear-clip's convexity test is consistent.
		if signedArea2D(loop2) < 0 {
			for i, j := 0, len(loop)-1; i < j; i, j = i+1, j-1 {
				loop[i], loop[j] = loop[j], loop[i]
				loop2[i], loop2[j] = loop2[j], loop2[i]
			}
		}

		newTris, ok := earClip(loop, loop2)
		if !ok {
			emitOriginals(&out, td, verts, members)
			continue
		}
		for _, ti := range newTris {
			out = append(out, ms3.Triangle{verts[ti[0]], verts[ti[1]], verts[ti[2]]})
		}
	}
	return out
}

type triData struct {
	v    [3]int   // vertex indices into verts
	n    ms3.Vec  // unit normal
	d    float32  // plane offset (n·v0)
	key  planeKey // bucket for plane equivalence
	dead bool
}

func emitOriginals(out *[]ms3.Triangle, td []triData, verts []ms3.Vec, members []int) {
	for _, i := range members {
		t := td[i]
		*out = append(*out, ms3.Triangle{verts[t.v[0]], verts[t.v[1]], verts[t.v[2]]})
	}
}

// planeKey buckets triangles into coarse plane equivalence classes for
// efficient neighbour queries. The fine-grained match is done by planesMatch.
type planeKey struct{ nx, ny, nz, d int32 }

func keyFor(n ms3.Vec, d float32, opts Options) planeKey {
	// Use a quantisation finer than the tolerances so the bucket reliably
	// catches matches; planesMatch then validates the actual angle/offset.
	const qN = 64.0 // ~1.4° resolution per axis component
	qD := 1.0 / opts.OffsetTol / 4
	return planeKey{
		nx: int32(math32.Round(n.X * qN)),
		ny: int32(math32.Round(n.Y * qN)),
		nz: int32(math32.Round(n.Z * qN)),
		d:  int32(math32.Round(d * float32(qD))),
	}
}

func planesMatch(n1 ms3.Vec, d1 float32, n2 ms3.Vec, d2 float32, opts Options) bool {
	if ms3.Dot(n1, n2) < opts.NormalTol {
		return false
	}
	return math32.Abs(d1-d2) <= opts.OffsetTol
}

// weldVertices returns the unique vertex array and a flat slice of
// vertex indices into it (3 per input triangle).
func weldVertices(tris []ms3.Triangle, tol float32) ([]ms3.Vec, []int) {
	q := 1.0 / tol
	type k struct{ x, y, z int64 }
	idx := make(map[k]int)
	verts := make([]ms3.Vec, 0, len(tris)*3)
	out := make([]int, 0, len(tris)*3)
	for _, t := range tris {
		for j := 0; j < 3; j++ {
			p := t[j]
			key := k{
				x: int64(math.Round(float64(p.X) * float64(q))),
				y: int64(math.Round(float64(p.Y) * float64(q))),
				z: int64(math.Round(float64(p.Z) * float64(q))),
			}
			i, ok := idx[key]
			if !ok {
				i = len(verts)
				idx[key] = i
				verts = append(verts, p)
			}
			out = append(out, i)
		}
	}
	return verts, out
}

// planeBasis returns two unit vectors orthogonal to n and to each other.
func planeBasis(n ms3.Vec) (u, v ms3.Vec) {
	// Pick the world axis least aligned with n.
	ax, ay, az := math32.Abs(n.X), math32.Abs(n.Y), math32.Abs(n.Z)
	var seed ms3.Vec
	switch {
	case ax <= ay && ax <= az:
		seed = ms3.Vec{X: 1}
	case ay <= az:
		seed = ms3.Vec{Y: 1}
	default:
		seed = ms3.Vec{Z: 1}
	}
	u = ms3.Unit(ms3.Cross(n, seed))
	v = ms3.Cross(n, u) // already unit since n and u are orthonormal
	return
}

// dropCollinear removes vertices that lie on the segment between their
// neighbours. Tolerance is the perpendicular distance threshold.
func dropCollinear(loop []int, p2 [][2]float32, tol float32) ([]int, [][2]float32) {
	n := len(loop)
	if n < 4 {
		return loop, p2
	}
	keep := make([]bool, n)
	for i := range keep {
		keep[i] = true
	}
	// Iterate until stable: removing a vertex may make its neighbour collinear.
	for changed := true; changed; {
		changed = false
		// Find prev/next that are still kept.
		idx := make([]int, 0, n)
		for i, k := range keep {
			if k {
				idx = append(idx, i)
			}
		}
		if len(idx) < 4 {
			break
		}
		for k, i := range idx {
			ip := idx[(k-1+len(idx))%len(idx)]
			in := idx[(k+1)%len(idx)]
			a, b, c := p2[ip], p2[i], p2[in]
			// Perpendicular distance from b to line a-c.
			abx, aby := c[0]-a[0], c[1]-a[1]
			length := math32.Sqrt(abx*abx + aby*aby)
			if length == 0 {
				continue
			}
			cross := math32.Abs((c[0]-a[0])*(a[1]-b[1]) - (a[0]-b[0])*(c[1]-a[1]))
			d := cross / length
			if d <= tol {
				keep[i] = false
				changed = true
			}
		}
	}
	var nl []int
	var np [][2]float32
	for i, k := range keep {
		if k {
			nl = append(nl, loop[i])
			np = append(np, p2[i])
		}
	}
	return nl, np
}

func signedArea2D(p [][2]float32) float32 {
	var a float32
	n := len(p)
	for i := 0; i < n; i++ {
		j := (i + 1) % n
		a += p[i][0]*p[j][1] - p[j][0]*p[i][1]
	}
	return a / 2
}

// earClip triangulates a simple polygon (CCW). Returns triples of original
// vertex indices. Reports false if it can't make progress — caller falls back.
func earClip(loop []int, p2 [][2]float32) ([][3]int, bool) {
	n := len(loop)
	if n < 3 {
		return nil, false
	}
	if n == 3 {
		return [][3]int{{loop[0], loop[1], loop[2]}}, true
	}
	prev := make([]int, n)
	next := make([]int, n)
	for i := 0; i < n; i++ {
		prev[i] = (i - 1 + n) % n
		next[i] = (i + 1) % n
	}
	out := make([][3]int, 0, n-2)
	remaining := n
	guard := 0
	i := 0
	for remaining > 3 {
		guard++
		if guard > 4*n*n {
			return nil, false
		}
		ip, in := prev[i], next[i]
		if isEar(p2, ip, i, in, next) {
			out = append(out, [3]int{loop[ip], loop[i], loop[in]})
			next[ip] = in
			prev[in] = ip
			remaining--
			i = ip
			continue
		}
		i = next[i]
	}
	out = append(out, [3]int{loop[prev[i]], loop[i], loop[next[i]]})
	return out, true
}

func isEar(p [][2]float32, a, b, c int, next []int) bool {
	if cross2D(p[a], p[b], p[c]) <= 0 {
		return false // reflex
	}
	// No other polygon vertex inside triangle abc.
	for k := next[c]; k != a; k = next[k] {
		if pointInTri(p[k], p[a], p[b], p[c]) {
			return false
		}
	}
	return true
}

func cross2D(a, b, c [2]float32) float32 {
	return (b[0]-a[0])*(c[1]-a[1]) - (b[1]-a[1])*(c[0]-a[0])
}

func pointInTri(p, a, b, c [2]float32) bool {
	d1 := cross2D(a, b, p)
	d2 := cross2D(b, c, p)
	d3 := cross2D(c, a, p)
	hasNeg := d1 < 0 || d2 < 0 || d3 < 0
	hasPos := d1 > 0 || d2 > 0 || d3 > 0
	return !(hasNeg && hasPos)
}

// --- union-find ---

type uf struct{ parent, rank []int }

func newUF(n int) *uf {
	p := make([]int, n)
	for i := range p {
		p[i] = i
	}
	return &uf{parent: p, rank: make([]int, n)}
}

func (u *uf) find(x int) int {
	for u.parent[x] != x {
		u.parent[x] = u.parent[u.parent[x]]
		x = u.parent[x]
	}
	return x
}

func (u *uf) union(x, y int) {
	rx, ry := u.find(x), u.find(y)
	if rx == ry {
		return
	}
	if u.rank[rx] < u.rank[ry] {
		rx, ry = ry, rx
	}
	u.parent[ry] = rx
	if u.rank[rx] == u.rank[ry] {
		u.rank[rx]++
	}
}
