// Package svgslice extracts a 2D outline of a 3D SDF on an arbitrary plane
// using marching squares, and writes it as an SVG file. Useful for cutaway
// drawings of printed parts.
package svgslice

import (
	"bufio"
	"fmt"
	"math"
	"os"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf/gleval"
)

// Plane describes a 2D slice of a 3D SDF. Local plane coordinates (u, v)
// map to world coordinates as p = Origin + u·U + v·V. U and V should be
// unit vectors orthogonal to each other; UMin/UMax/VMin/VMax bound the
// rendered region. Output SVG coordinates use u as X and v as Y.
type Plane struct {
	Origin     ms3.Vec
	U, V       ms3.Vec
	UMin, UMax float32
	VMin, VMax float32
}

// XZ returns a Plane positioned at the origin, U=+X, V=+Z, bounded by the
// given half-extents in u and v. Convenient for axisymmetric parts oriented
// along Z (e.g. screws with the head at z=0 and the tip at -z).
func XZ(uHalf, vHalfPos, vHalfNeg float32) Plane {
	return Plane{
		Origin: ms3.Vec{},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Z: 1},
		UMin:   -uHalf, UMax: uHalf,
		VMin: -vHalfNeg, VMax: vHalfPos,
	}
}

// Label is a text annotation drawn on top of the contour. (U, V) is the
// position in plane-local coordinates (same axes as Plane.UMin..UMax /
// Plane.VMin..VMax). Anchor is "start", "middle", or "end" — defaults to
// "middle" when empty. FontSize is in mm; defaults to 1.5. Color is a
// hex string ("#444…"); defaults to "#333".
type Label struct {
	U, V     float32
	Text     string
	Anchor   string
	FontSize float32
	Color    string
}

// Dimension is a mechanical-drawing-style dimension annotation: two
// extension lines from the measured feature points (From, To) to the
// dimension line endpoints (DimFrom, DimTo), the dimension line itself
// with inward-pointing arrows at both ends, and a centred label. If
// Label is empty, the dimension line length |DimTo-DimFrom| is auto-
// formatted to one decimal millimetre (e.g. "10.0 mm"). FontSize defaults
// to 1.5 mm; Color defaults to "#333".
// seg is one marching-squares contour segment in plane-local (U, V).
type seg struct{ a, b ms2.Vec }

type Dimension struct {
	From, To       ms2.Vec
	DimFrom, DimTo ms2.Vec
	Label          string
	FontSize       float32
	Color          string
}

// WriteSlice samples sdf3 on a gridX×gridY grid covering the Plane
// rectangle, runs marching squares to extract zero-crossing line segments,
// and writes the result to filename as an SVG.
func WriteSlice(filename string, sdf3 gleval.SDF3, plane Plane, gridX, gridY int) error {
	return WriteSliceLabelled(filename, sdf3, plane, gridX, gridY, nil)
}

// WriteSliceLabelled is WriteSlice with optional text labels and
// dimension annotations drawn on top of the contour.
func WriteSliceLabelled(filename string, sdf3 gleval.SDF3, plane Plane, gridX, gridY int, labels []Label, dims ...Dimension) error {
	if gridX < 2 || gridY < 2 {
		return fmt.Errorf("svgslice: grid must be at least 2x2")
	}

	n := gridX * gridY
	pos := make([]ms3.Vec, n)
	dist := make([]float32, n)

	du := (plane.UMax - plane.UMin) / float32(gridX-1)
	dv := (plane.VMax - plane.VMin) / float32(gridY-1)
	corner00 := ms3.Add(plane.Origin,
		ms3.Add(ms3.Scale(plane.UMin, plane.U), ms3.Scale(plane.VMin, plane.V)))

	for j := 0; j < gridY; j++ {
		for i := 0; i < gridX; i++ {
			p := ms3.Add(corner00,
				ms3.Add(ms3.Scale(float32(i)*du, plane.U),
					ms3.Scale(float32(j)*dv, plane.V)))
			pos[j*gridX+i] = p
		}
	}
	if err := sdf3.Evaluate(pos, dist, nil); err != nil {
		return fmt.Errorf("svgslice: SDF evaluation failed: %w", err)
	}

	var segs []seg

	interp := func(d0, d1 float32, p0, p1 ms2.Vec) ms2.Vec {
		if d1 == d0 {
			return p0
		}
		t := d0 / (d0 - d1)
		return ms2.Vec{X: p0.X + t*(p1.X-p0.X), Y: p0.Y + t*(p1.Y-p0.Y)}
	}
	cornerUV := func(i, j int) ms2.Vec {
		return ms2.Vec{X: plane.UMin + float32(i)*du, Y: plane.VMin + float32(j)*dv}
	}
	insideBit := func(d float32) int {
		if d < 0 {
			return 1
		}
		return 0
	}

	for j := 0; j < gridY-1; j++ {
		for i := 0; i < gridX-1; i++ {
			dBL := dist[j*gridX+i]
			dBR := dist[j*gridX+i+1]
			dTR := dist[(j+1)*gridX+i+1]
			dTL := dist[(j+1)*gridX+i]
			pBL := cornerUV(i, j)
			pBR := cornerUV(i+1, j)
			pTR := cornerUV(i+1, j+1)
			pTL := cornerUV(i, j+1)

			code := insideBit(dBL) | insideBit(dBR)<<1 | insideBit(dTR)<<2 | insideBit(dTL)<<3
			if code == 0 || code == 15 {
				continue
			}
			eB := interp(dBL, dBR, pBL, pBR)
			eR := interp(dBR, dTR, pBR, pTR)
			eT := interp(dTL, dTR, pTL, pTR)
			eL := interp(dBL, dTL, pBL, pTL)

			switch code {
			case 1, 14:
				segs = append(segs, seg{a: eL, b: eB})
			case 2, 13:
				segs = append(segs, seg{a: eB, b: eR})
			case 3, 12:
				segs = append(segs, seg{a: eL, b: eR})
			case 4, 11:
				segs = append(segs, seg{a: eR, b: eT})
			case 5:
				segs = append(segs, seg{a: eL, b: eB})
				segs = append(segs, seg{a: eR, b: eT})
			case 6, 9:
				segs = append(segs, seg{a: eB, b: eT})
			case 7, 8:
				segs = append(segs, seg{a: eL, b: eT})
			case 10:
				segs = append(segs, seg{a: eL, b: eT})
				segs = append(segs, seg{a: eB, b: eR})
			}
		}
	}

	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()

	// Margin around the slice region in plane units. Sized so a single
	// dimension annotation (extension line ~1.5 mm + label ~1.5 mm font
	// + label width allowance) doesn't overflow.
	margin := float32(5.0)
	uMin, uMax := plane.UMin-margin, plane.UMax+margin
	vMin, vMax := plane.VMin-margin, plane.VMax+margin
	yMin, yMax := -vMax, -vMin

	width, height := uMax-uMin, yMax-yMin
	fmt.Fprint(w, "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n")
	fmt.Fprintf(w,
		"<svg xmlns=\"http://www.w3.org/2000/svg\" viewBox=\"%g %g %g %g\" width=\"%gmm\" height=\"%gmm\">\n",
		uMin, yMin, width, height, width, height)
	fmt.Fprintf(w,
		"  <rect x=\"%g\" y=\"%g\" width=\"%g\" height=\"%g\" fill=\"#fafafa\"/>\n",
		uMin, yMin, width, height)
	// Chain marching-squares line segments into closed loops, then emit
	// as a single filled+stroked SVG path. fill-rule="evenodd" makes
	// inner loops (e.g., screw holes) subtract from the outer fill.
	loops := chainLoops(segs)
	if len(loops) > 0 {
		fmt.Fprint(w, "  <path fill=\"#e8e8e8\" fill-rule=\"evenodd\" stroke=\"black\" stroke-width=\"0.2\" stroke-linejoin=\"round\" d=\"")
		for _, loop := range loops {
			fmt.Fprintf(w, "M %g %g", loop[0].X, -loop[0].Y)
			for _, p := range loop[1:] {
				fmt.Fprintf(w, " L %g %g", p.X, -p.Y)
			}
			fmt.Fprint(w, " Z")
		}
		fmt.Fprint(w, "\"/>\n")
	}

	if len(labels) > 0 {
		fmt.Fprint(w, "  <g font-family=\"sans-serif\">\n")
		for _, l := range labels {
			anchor := l.Anchor
			if anchor == "" {
				anchor = "middle"
			}
			fontSize := l.FontSize
			if fontSize == 0 {
				fontSize = 1.5
			}
			color := l.Color
			if color == "" {
				color = "#333"
			}
			// SVG y is flipped relative to plane V (see line above).
			fmt.Fprintf(w,
				"    <text x=\"%g\" y=\"%g\" font-size=\"%g\" fill=\"%s\" text-anchor=\"%s\" dominant-baseline=\"middle\">%s</text>\n",
				l.U, -l.V, fontSize, color, anchor, escapeXML(l.Text))
		}
		fmt.Fprint(w, "  </g>\n")
	}

	for _, d := range dims {
		writeDimension(w, d)
	}

	// XYZ axis gizmo in the bottom-left corner of the viewBox.
	writeAxisGizmo(w, plane, uMin+3, -vMin-3)

	fmt.Fprint(w, "</svg>\n")
	return nil
}

// writeAxisGizmo emits a small XYZ orientation indicator at SVG position
// (originX, originY). The two in-plane axes (Plane.U, Plane.V) are drawn
// as labelled arrows; the perpendicular axis is shown as a coloured dot
// at the origin. Convention: X red, Y green, Z blue.
func writeAxisGizmo(w *bufio.Writer, plane Plane, originX, originY float32) {
	uAx := axisOf(plane.U)
	vAx := axisOf(plane.V)
	perpAx := perpAxisOf(plane.U, plane.V)
	if uAx == "?" || vAx == "?" {
		return // non-cardinal slice; skip the gizmo
	}
	const (
		armLen   = 4.0
		arrowW   = 0.4
		strokeW  = 0.3
		fontSize = 1.5
	)

	// U arrow: along +X SVG (right).
	tipUX, tipUY := originX+armLen, originY
	fmt.Fprintf(w,
		"  <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\" stroke=\"%s\" stroke-width=\"%g\"/>\n",
		originX, originY, tipUX, tipUY, axisColor(uAx), strokeW)
	fmt.Fprintf(w,
		"  <polygon points=\"%g,%g %g,%g %g,%g\" fill=\"%s\"/>\n",
		tipUX, tipUY, tipUX-arrowW, tipUY-arrowW/2, tipUX-arrowW, tipUY+arrowW/2,
		axisColor(uAx))
	fmt.Fprintf(w,
		"  <text x=\"%g\" y=\"%g\" font-size=\"%g\" fill=\"%s\" text-anchor=\"start\" dominant-baseline=\"middle\" font-family=\"sans-serif\">%s</text>\n",
		tipUX+0.6, tipUY, fontSize, axisColor(uAx), uAx)

	// V arrow: along -Y SVG (up on screen, since SVG y is flipped from V).
	tipVX, tipVY := originX, originY-armLen
	fmt.Fprintf(w,
		"  <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\" stroke=\"%s\" stroke-width=\"%g\"/>\n",
		originX, originY, tipVX, tipVY, axisColor(vAx), strokeW)
	fmt.Fprintf(w,
		"  <polygon points=\"%g,%g %g,%g %g,%g\" fill=\"%s\"/>\n",
		tipVX, tipVY, tipVX-arrowW/2, tipVY+arrowW, tipVX+arrowW/2, tipVY+arrowW,
		axisColor(vAx))
	fmt.Fprintf(w,
		"  <text x=\"%g\" y=\"%g\" font-size=\"%g\" fill=\"%s\" text-anchor=\"middle\" dominant-baseline=\"alphabetic\" font-family=\"sans-serif\">%s</text>\n",
		tipVX, tipVY-0.6, fontSize, axisColor(vAx), vAx)

	// Perpendicular axis: dot at origin, label to the lower-left.
	fmt.Fprintf(w,
		"  <circle cx=\"%g\" cy=\"%g\" r=\"0.55\" fill=\"%s\"/>\n",
		originX, originY, axisColor(perpAx))
	fmt.Fprintf(w,
		"  <text x=\"%g\" y=\"%g\" font-size=\"%g\" fill=\"%s\" text-anchor=\"end\" dominant-baseline=\"hanging\" font-family=\"sans-serif\">%s</text>\n",
		originX-0.8, originY+0.6, fontSize, axisColor(perpAx), perpAx)
}

func axisOf(v ms3.Vec) string {
	switch {
	case v.X != 0 && v.Y == 0 && v.Z == 0:
		return "X"
	case v.X == 0 && v.Y != 0 && v.Z == 0:
		return "Y"
	case v.X == 0 && v.Y == 0 && v.Z != 0:
		return "Z"
	}
	return "?"
}

func perpAxisOf(u, v ms3.Vec) string {
	uAx, vAx := axisOf(u), axisOf(v)
	for _, c := range []string{"X", "Y", "Z"} {
		if c != uAx && c != vAx {
			return c
		}
	}
	return "?"
}

func axisColor(ax string) string {
	switch ax {
	case "X":
		return "#dc2626"
	case "Y":
		return "#16a34a"
	case "Z":
		return "#2563eb"
	}
	return "#666"
}

// chainLoops walks the marching-squares segment soup and returns closed
// (or near-closed) loops by following adjacency through the shared
// endpoints. The MS interpolator produces bitwise-identical floats for
// the same grid edge from adjacent cells, so endpoints match exactly
// without quantisation. Saddle vertices (degree 4) are resolved
// arbitrarily — fine for our parts since saddles are rare at the
// resolutions we use.
func chainLoops(segs []seg) [][]ms2.Vec {
	type segKey struct{ a, b ms2.Vec }
	keyOf := func(a, b ms2.Vec) segKey {
		if a.X < b.X || (a.X == b.X && a.Y < b.Y) {
			return segKey{a, b}
		}
		return segKey{b, a}
	}
	adj := map[ms2.Vec][]ms2.Vec{}
	for _, s := range segs {
		adj[s.a] = append(adj[s.a], s.b)
		adj[s.b] = append(adj[s.b], s.a)
	}
	visited := make(map[segKey]bool, len(segs))
	var loops [][]ms2.Vec
	for _, s := range segs {
		if visited[keyOf(s.a, s.b)] {
			continue
		}
		visited[keyOf(s.a, s.b)] = true
		loop := []ms2.Vec{s.a, s.b}
		prev, cur := s.a, s.b
		for {
			var nx ms2.Vec
			found := false
			for _, n := range adj[cur] {
				if n == prev {
					continue
				}
				if visited[keyOf(cur, n)] {
					continue
				}
				nx = n
				found = true
				break
			}
			if !found {
				break // dead end (open contour at boundary or saddle artifact)
			}
			visited[keyOf(cur, nx)] = true
			if nx == loop[0] {
				break // closed
			}
			loop = append(loop, nx)
			prev, cur = cur, nx
		}
		loops = append(loops, loop)
	}
	return loops
}

// writeDimension renders one dimension annotation (extension lines, dim
// line, two inward-pointing arrowheads, centred text label). All math
// is done in plane (U, V); SVG y is the negation of V.
func writeDimension(w *bufio.Writer, d Dimension) {
	fontSize := d.FontSize
	if fontSize == 0 {
		fontSize = 1.5
	}
	color := d.Color
	if color == "" {
		color = "#1e40af" // dark blue
	}
	dx := d.DimTo.X - d.DimFrom.X
	dy := d.DimTo.Y - d.DimFrom.Y
	dlen := float32(math.Hypot(float64(dx), float64(dy)))
	if dlen == 0 {
		return
	}
	ux, uy := dx/dlen, dy/dlen // unit along dim line, From→To
	// Perpendicular (90° CCW in plane V): rotate (x,y) → (-y, x).
	px, py := -uy, ux
	label := d.Label
	if label == "" {
		label = fmt.Sprintf("%.1f mm", dlen)
	}

	// Stroke width tuned to look right against the 0.2 contour stroke.
	const strokeW = 0.12

	// Extension lines (From → DimFrom, To → DimTo).
	fmt.Fprintf(w,
		"  <g stroke=\"%s\" stroke-width=\"%g\" fill=\"none\">\n",
		color, strokeW)
	fmt.Fprintf(w,
		"    <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\"/>\n",
		d.From.X, -d.From.Y, d.DimFrom.X, -d.DimFrom.Y)
	fmt.Fprintf(w,
		"    <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\"/>\n",
		d.To.X, -d.To.Y, d.DimTo.X, -d.DimTo.Y)
	// Dimension line.
	fmt.Fprintf(w,
		"    <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\"/>\n",
		d.DimFrom.X, -d.DimFrom.Y, d.DimTo.X, -d.DimTo.Y)
	fmt.Fprint(w, "  </g>\n")

	// Arrowheads — solid triangles. Inside arrows (tip at endpoint, base
	// pointing into the dim line) when the line is long enough; outside
	// arrows (tip at endpoint, base pointing AWAY from the dim line)
	// when it isn't, so they don't overlap each other or the label.
	const (
		arrowLen        = 1.0
		arrowHalf       = 0.35
		insideThreshold = 2.0 // dlen ≤ this → outside arrows
	)
	sign := float32(1)
	if dlen <= insideThreshold {
		sign = -1 // base offset goes outward
	}
	baseFromX, baseFromY := d.DimFrom.X+ux*arrowLen*sign, d.DimFrom.Y+uy*arrowLen*sign
	baseToX, baseToY := d.DimTo.X-ux*arrowLen*sign, d.DimTo.Y-uy*arrowLen*sign
	fmt.Fprintf(w,
		"  <g fill=\"%s\" stroke=\"none\">\n", color)
	fmt.Fprintf(w,
		"    <polygon points=\"%g,%g %g,%g %g,%g\"/>\n",
		d.DimFrom.X, -d.DimFrom.Y,
		baseFromX+px*arrowHalf, -(baseFromY + py*arrowHalf),
		baseFromX-px*arrowHalf, -(baseFromY - py*arrowHalf))
	fmt.Fprintf(w,
		"    <polygon points=\"%g,%g %g,%g %g,%g\"/>\n",
		d.DimTo.X, -d.DimTo.Y,
		baseToX+px*arrowHalf, -(baseToY + py*arrowHalf),
		baseToX-px*arrowHalf, -(baseToY - py*arrowHalf))
	fmt.Fprint(w, "  </g>\n")
	// When arrows go outside, extend the dimension line to reach the
	// arrow bases so the line connects visually to both arrowheads.
	if sign < 0 {
		fmt.Fprintf(w,
			"  <g stroke=\"%s\" stroke-width=\"%g\" fill=\"none\">\n",
			color, strokeW)
		fmt.Fprintf(w,
			"    <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\"/>\n",
			d.DimFrom.X, -d.DimFrom.Y, baseFromX, -baseFromY)
		fmt.Fprintf(w,
			"    <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\"/>\n",
			d.DimTo.X, -d.DimTo.Y, baseToX, -baseToY)
		fmt.Fprint(w, "  </g>\n")
	}

	// Label at midpoint, offset by fontSize*0.7 perpendicular to the dim
	// line, on the side AWAY from From (so the text doesn't sit on top
	// of the part). Side selection: pick the perp direction whose dot
	// with (DimFrom - From) is positive.
	midX, midY := (d.DimFrom.X+d.DimTo.X)/2, (d.DimFrom.Y+d.DimTo.Y)/2
	awayX, awayY := d.DimFrom.X-d.From.X, d.DimFrom.Y-d.From.Y
	if px*awayX+py*awayY < 0 {
		px, py = -px, -py
	}
	off := fontSize * 0.7
	fmt.Fprintf(w,
		"  <text x=\"%g\" y=\"%g\" font-size=\"%g\" fill=\"%s\" text-anchor=\"middle\" dominant-baseline=\"middle\" font-family=\"sans-serif\">%s</text>\n",
		midX+px*off, -(midY + py*off), fontSize, color, escapeXML(label))
}

func escapeXML(s string) string {
	const replace = "&<>\"'"
	for i := 0; i < len(s); i++ {
		if c := s[i]; c < 128 && (c == '&' || c == '<' || c == '>' || c == '"' || c == '\'') {
			// fall through to slow path
			b := make([]byte, 0, len(s)+8)
			for j := 0; j < len(s); j++ {
				switch s[j] {
				case '&':
					b = append(b, "&amp;"...)
				case '<':
					b = append(b, "&lt;"...)
				case '>':
					b = append(b, "&gt;"...)
				case '"':
					b = append(b, "&quot;"...)
				case '\'':
					b = append(b, "&#39;"...)
				default:
					b = append(b, s[j])
				}
			}
			return string(b)
		}
		_ = replace
	}
	return s
}
