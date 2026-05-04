package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	threemf "codeberg.org/hum3/gsdf-3mf"

	"codeberg.org/hum3/go3dp/pkg/meshopt"
	"codeberg.org/hum3/go3dp/pkg/svgslice"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
	"github.com/soypat/gsdf/gleval"
	"github.com/soypat/gsdf/glrender"
	"github.com/soypat/gsdf/gsdfaux"
)

func main() {
	var (
		size   = flag.String("size", "m", "v0 size: xs, s, m, l, all")
		part   = flag.String("part", "all", "v0 part: block, cover, adaptor, all")
		out    = flag.String("out", "stl", "output kind: stl, svg, 3mf, all")
		resDiv = flag.Uint("resdiv", 200, "resolution = bounding-box diagonal / resdiv (mesh outputs)")
		mesh   = flag.String("mesh", "merge", "3MF mesher: octree (marching cubes), dc (dual contouring), merge (DC + planar merge, default)")
	)
	flag.Parse()

	if err := run(*size, *part, *out, *resDiv, *mesh); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(sizeArg, partArg, out string, resDiv uint, mesh string) error {
	switch mesh {
	case "octree", "dc", "merge":
	default:
		return fmt.Errorf("unknown -mesh %q (try octree, dc, merge)", mesh)
	}

	sizes, err := selectSizes(sizeArg)
	if err != nil {
		return err
	}
	wantBlock, wantCover, wantAdaptor, wantAssembly, err := selectParts(partArg)
	if err != nil {
		return err
	}
	wantSTL := out == "stl" || out == "all" || out == "both"
	wantSVG := out == "svg" || out == "all" || out == "both"
	want3MF := out == "3mf" || out == "all"
	if !wantSTL && !wantSVG && !want3MF {
		return fmt.Errorf("unknown -out %q (try stl, svg, 3mf, all)", out)
	}

	for _, sz := range sizes {
		if wantBlock {
			if err := buildPart(sz, "block", V0Block, wantSTL, wantSVG, want3MF, resDiv, mesh); err != nil {
				return fmt.Errorf("%s block: %w", sz.Name, err)
			}
		}
		if wantCover {
			if err := buildPart(sz, "cover", V0Cover, wantSTL, wantSVG, want3MF, resDiv, mesh); err != nil {
				return fmt.Errorf("%s cover: %w", sz.Name, err)
			}
		}
		if wantAdaptor {
			if err := buildPart(sz, "adaptor", V0SlideOnPLA, wantSTL, wantSVG, want3MF, resDiv, mesh); err != nil {
				return fmt.Errorf("%s adaptor: %w", sz.Name, err)
			}
		}
		if wantAssembly {
			if err := buildPart(sz, "assembly", V0Assembly, wantSTL, wantSVG, want3MF, resDiv, mesh); err != nil {
				return fmt.Errorf("%s assembly: %w", sz.Name, err)
			}
		}
	}
	return nil
}

func selectParts(arg string) (block, cover, adaptor, assembly bool, err error) {
	switch arg {
	case "block":
		return true, false, false, false, nil
	case "cover":
		return false, true, false, false, nil
	case "adaptor":
		return false, false, true, false, nil
	case "assembly":
		return false, false, false, true, nil
	case "all":
		return true, true, true, false, nil // assembly is opt-in (debug only)
	case "both": // legacy: block + cover
		return true, true, false, false, nil
	default:
		return false, false, false, false, fmt.Errorf("unknown -part %q (try block, cover, adaptor, assembly, all)", arg)
	}
}

func selectSizes(arg string) ([]V0Size, error) {
	switch arg {
	case "xs":
		return []V0Size{V0_XS}, nil
	case "s":
		return []V0Size{V0_S}, nil
	case "m":
		return []V0Size{V0_M}, nil
	case "l":
		return []V0Size{V0_L}, nil
	case "all":
		return V0Sizes, nil
	default:
		return nil, fmt.Errorf("unknown -size %q (try xs, s, m, l, all)", arg)
	}
}

type partBuilder func(*gsdf.Builder, V0Size) (glbuild.Shader3D, error)

func buildPart(sz V0Size, partLabel string, build partBuilder, wantSTL, wantSVG, want3MF bool, resDiv uint, mesh string) error {
	var bld gsdf.Builder
	shape, err := build(&bld, sz)
	if err != nil {
		return fmt.Errorf("build: %w", err)
	}
	if berr := bld.Err(); berr != nil {
		return fmt.Errorf("builder: %w", berr)
	}

	baseName := sz.Name + "_" + partLabel
	if wantSTL {
		if err := writeSTL(shape, baseName+".stl", resDiv); err != nil {
			return fmt.Errorf("STL: %w", err)
		}
	}
	if want3MF {
		if err := write3MF(shape, baseName+".3mf", sz.Name+" "+partLabel, resDiv, mesh); err != nil {
			return fmt.Errorf("3MF: %w", err)
		}
	}
	if wantSVG {
		if err := writeSVGCutaways(shape, baseName, sz); err != nil {
			return fmt.Errorf("SVG: %w", err)
		}
	}
	return nil
}

func write3MF(shape glbuild.Shader3D, filename, partName string, resDiv uint, mesh string) error {
	if err := glbuild.ShortenNames3D(&shape, 12); err != nil {
		return fmt.Errorf("shorten names: %w", err)
	}
	cpusdf, err := gleval.NewCPUSDF3(shape)
	if err != nil {
		return fmt.Errorf("cpu sdf: %w", err)
	}
	const bufSize = 4096
	cpusdf.VecPool().SetMinAllocationLen(bufSize)

	resolution := shape.Bounds().Diagonal() / float32(resDiv)
	tris, err := renderTriangles(cpusdf, resolution, mesh, bufSize)
	if err != nil {
		return err
	}
	rawN := len(tris)
	if mesh == "merge" {
		// Tolerances scaled to mesh resolution: a face is considered planar
		// if its normals match within ~1° and offsets within res/20, and
		// loop vertices closer than res/100 to their neighbours' line are
		// dropped as collinear inserts.
		tris = meshopt.PlanarMerge(tris, meshopt.Options{
			VertexTol: resolution / 100,
			OffsetTol: resolution / 20,
		})
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	parts := []threemf.Part{{
		Name:      partName,
		Color:     threemf.Color{R: 0xC8, G: 0xC8, B: 0xC8, A: 0xFF}, // light grey PLA
		Triangles: tris,
	}}
	if err := threemf.Write(f, parts, threemf.UnitMillimeter); err != nil {
		return err
	}
	if mesh == "merge" {
		fmt.Printf("wrote %s (%d triangles, %d after planar merge)\n", filename, rawN, len(tris))
	} else {
		fmt.Printf("wrote %s (%d triangles, mesher=%s)\n", filename, len(tris), mesh)
	}
	return nil
}

func renderTriangles(cpusdf *gleval.SDF3CPU, resolution float32, mesh string, bufSize int) ([]ms3.Triangle, error) {
	switch mesh {
	case "octree":
		r, err := glrender.NewOctreeRenderer(cpusdf, resolution, bufSize)
		if err != nil {
			return nil, fmt.Errorf("octree: %w", err)
		}
		tris, err := glrender.RenderAll(r, cpusdf.VecPool())
		if err != nil {
			return nil, fmt.Errorf("render: %w", err)
		}
		return tris, nil
	case "dc", "merge":
		var dcr glrender.DualContourRenderer
		if err := dcr.Reset(cpusdf, resolution, &glrender.DualContourLeastSquares{}, cpusdf.VecPool()); err != nil {
			return nil, fmt.Errorf("dual contour reset: %w", err)
		}
		tris, err := dcr.RenderAll(nil, cpusdf.VecPool())
		if err != nil {
			return nil, fmt.Errorf("render: %w", err)
		}
		return tris, nil
	default:
		return nil, fmt.Errorf("unknown mesher %q", mesh)
	}
}

func writeSTL(shape glbuild.Shader3D, filename string, resDiv uint) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	bb := shape.Bounds()
	resolution := float32(bb.Diagonal()) / float32(resDiv)
	cfg := gsdfaux.RenderConfig{
		STLOutput:  f,
		Resolution: resolution,
		Silent:     true,
	}
	if err := gsdfaux.RenderShader3D(shape, cfg); err != nil {
		return err
	}
	fmt.Printf("wrote %s\n", filename)
	return nil
}

func writeSVGCutaways(shape glbuild.Shader3D, baseName string, sz V0Size) error {
	if err := glbuild.ShortenNames3D(&shape, 12); err != nil {
		return fmt.Errorf("shorten names: %w", err)
	}
	sdf3, err := gleval.NewCPUSDF3(shape)
	if err != nil {
		return fmt.Errorf("new CPU SDF: %w", err)
	}

	bb := shape.Bounds()
	pad := float32(2.0)

	axial := svgslice.Plane{
		Origin: ms3.Vec{},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Z: 1},
		UMin:   bb.Min.X - pad, UMax: bb.Max.X + pad,
		VMin: bb.Min.Z - pad, VMax: bb.Max.Z + pad,
	}
	if err := svgslice.WriteSliceLabelled(baseName+"_cutaway_axial.svg", sdf3, axial, 256, 256, nil, axialDimensions(sz)...); err != nil {
		return err
	}
	fmt.Printf("wrote %s_cutaway_axial.svg\n", baseName)

	if y, ok := firstOffAxisY(sz.Holes); ok {
		offAxis := axial
		offAxis.Origin = ms3.Vec{Y: y}
		name := fmt.Sprintf("%s_cutaway_screwY%g.svg", baseName, y)
		if err := svgslice.WriteSliceLabelled(name, sdf3, offAxis, 256, 256, nil, axialDimensions(sz)...); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", name)
	}

	// YZ slice (third orthogonal axis) — looking from +X toward -X.
	// For symmetric parts (block, cover) the silhouette matches the
	// axial XZ slice, so reuse axialDimensions; for the adaptor the
	// numbers are wrong but at least the shape is informative.
	side := svgslice.Plane{
		Origin: ms3.Vec{},
		U:      ms3.Vec{Y: 1},
		V:      ms3.Vec{Z: 1},
		UMin:   bb.Min.Y - pad, UMax: bb.Max.Y + pad,
		VMin: bb.Min.Z - pad, VMax: bb.Max.Z + pad,
	}
	if err := svgslice.WriteSliceLabelled(baseName+"_cutaway_side.svg", sdf3, side, 256, 256, nil, axialDimensions(sz)...); err != nil {
		return err
	}
	fmt.Printf("wrote %s_cutaway_side.svg\n", baseName)

	midZ := (bb.Min.Z + bb.Max.Z) / 2
	top := svgslice.Plane{
		Origin: ms3.Vec{Z: midZ},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Y: 1},
		UMin:   bb.Min.X - pad, UMax: bb.Max.X + pad,
		VMin: bb.Min.Y - pad, VMax: bb.Max.Y + pad,
	}
	if err := svgslice.WriteSliceLabelled(baseName+"_cutaway_top.svg", sdf3, top, 256, 256, nil, topDimensions(sz)...); err != nil {
		return err
	}
	fmt.Printf("wrote %s_cutaway_top.svg\n", baseName)
	return nil
}

// axialDimensions annotates the axial cutaway with Wo (top), Wi
// (bottom), and H (right side) dimension lines.
//
// Plane is U=X, V=Z. Block trapezoid: small face Wi at v=0 (wall,
// bottom of SVG), large face Wo at v=H (top of SVG). Cover is the
// inverse: Wo at v=0, Wi at v=H. The same dimension positions work
// for either because the trapezoid silhouette is identical, just
// up/down-flipped.
func axialDimensions(sz V0Size) []svgslice.Dimension {
	wMax := sz.Wo
	if sz.Wi > wMax {
		wMax = sz.Wi
	}
	const dimGap = 1.5 // perpendicular gap from feature to dim line, mm
	side := wMax/2 + dimGap
	return []svgslice.Dimension{
		{ // Wo across the top face
			From:    ms2.Vec{X: -sz.Wo / 2, Y: sz.H},
			To:      ms2.Vec{X: +sz.Wo / 2, Y: sz.H},
			DimFrom: ms2.Vec{X: -sz.Wo / 2, Y: sz.H + dimGap},
			DimTo:   ms2.Vec{X: +sz.Wo / 2, Y: sz.H + dimGap},
		},
		{ // Wi across the bottom face
			From:    ms2.Vec{X: -sz.Wi / 2, Y: 0},
			To:      ms2.Vec{X: +sz.Wi / 2, Y: 0},
			DimFrom: ms2.Vec{X: -sz.Wi / 2, Y: -dimGap},
			DimTo:   ms2.Vec{X: +sz.Wi / 2, Y: -dimGap},
		},
		{ // H on the right side
			From:    ms2.Vec{X: sz.Wi / 2, Y: 0},
			To:      ms2.Vec{X: sz.Wo / 2, Y: sz.H},
			DimFrom: ms2.Vec{X: side, Y: 0},
			DimTo:   ms2.Vec{X: side, Y: sz.H},
		},
	}
}

// topDimensions annotates the top-down cutaway with the across-flats
// width at z=H/2 and (when present) the counterbore diameter at the
// first screw centre.
func topDimensions(sz V0Size) []svgslice.Dimension {
	mid := (sz.Wi + sz.Wo) / 2
	const dimGap = 1.5
	dims := []svgslice.Dimension{
		{
			From:    ms2.Vec{X: -mid / 2, Y: mid / 2},
			To:      ms2.Vec{X: +mid / 2, Y: mid / 2},
			DimFrom: ms2.Vec{X: -mid / 2, Y: mid/2 + dimGap},
			DimTo:   ms2.Vec{X: +mid / 2, Y: mid/2 + dimGap},
		},
	}
	// The top-view plane is at z = H/2 (mid-height), well below the
	// counterbore, so the counterbore itself isn't visible in the
	// outline. Annotate it as a diameter call-out at the first screw
	// position with an explicit "Ø…" label.
	if sz.Counterbore > 0 && len(sz.Holes) > 0 {
		r := sz.Screw.DHead/2 + sz.Recess
		c := sz.Holes[0]
		dims = append(dims, svgslice.Dimension{
			From:    ms2.Vec{X: c.X - r, Y: c.Y},
			To:      ms2.Vec{X: c.X + r, Y: c.Y},
			DimFrom: ms2.Vec{X: c.X - r, Y: c.Y - mid/2 - dimGap},
			DimTo:   ms2.Vec{X: c.X + r, Y: c.Y - mid/2 - dimGap},
			Label:   fmt.Sprintf("Ø%.1f mm counterbore", 2*r),
		})
	}
	return dims
}

func firstOffAxisY(holes []ms2.Vec) (float32, bool) {
	for _, h := range holes {
		if h.Y != 0 {
			return h.Y, true
		}
	}
	return 0, false
}
