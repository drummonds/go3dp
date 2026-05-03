// Package svgslice extracts a 2D outline of a 3D SDF on an arbitrary plane
// using marching squares, and writes it as an SVG file. Useful for cutaway
// drawings of printed parts.
package svgslice

import (
	"bufio"
	"fmt"
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

// WriteSlice samples sdf3 on a gridX×gridY grid covering the Plane
// rectangle, runs marching squares to extract zero-crossing line segments,
// and writes the result to filename as an SVG.
func WriteSlice(filename string, sdf3 gleval.SDF3, plane Plane, gridX, gridY int) error {
	return WriteSliceLabelled(filename, sdf3, plane, gridX, gridY, nil)
}

// WriteSliceLabelled is WriteSlice with an optional list of labels drawn
// on top of the contour. Pass nil labels for plain output.
func WriteSliceLabelled(filename string, sdf3 gleval.SDF3, plane Plane, gridX, gridY int, labels []Label) error {
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

	type seg struct{ a, b ms2.Vec }
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

	margin := float32(2.0)
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
	fmt.Fprint(w, "  <g stroke=\"black\" stroke-width=\"0.2\" fill=\"none\">\n")
	for _, s := range segs {
		fmt.Fprintf(w,
			"    <line x1=\"%g\" y1=\"%g\" x2=\"%g\" y2=\"%g\"/>\n",
			s.a.X, -s.a.Y, s.b.X, -s.b.Y)
	}
	fmt.Fprint(w, "  </g>\n")

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

	fmt.Fprint(w, "</svg>\n")
	return nil
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
