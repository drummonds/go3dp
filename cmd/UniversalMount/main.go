package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	threemf "codeberg.org/hum3/gsdf-3mf"

	"codeberg.org/hum3/go3dp/pkg/svgslice"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/gleval"
	"github.com/soypat/gsdf/glbuild"
	"github.com/soypat/gsdf/glrender"
	"github.com/soypat/gsdf/gsdfaux"
)

func main() {
	var (
		size   = flag.String("size", "m", "v0 size: xs, s, m, l, all")
		part   = flag.String("part", "all", "v0 part: block, cover, adaptor, all")
		out    = flag.String("out", "stl", "output kind: stl, svg, 3mf, all")
		resDiv = flag.Uint("resdiv", 200, "resolution = bounding-box diagonal / resdiv (mesh outputs)")
	)
	flag.Parse()

	if err := run(*size, *part, *out, *resDiv); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(sizeArg, partArg, out string, resDiv uint) error {
	sizes, err := selectSizes(sizeArg)
	if err != nil {
		return err
	}
	wantBlock, wantCover, wantAdaptor, err := selectParts(partArg)
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
			if err := buildPart(sz, "block", V0Block, wantSTL, wantSVG, want3MF, resDiv); err != nil {
				return fmt.Errorf("%s block: %w", sz.Name, err)
			}
		}
		if wantCover {
			if err := buildPart(sz, "cover", V0Cover, wantSTL, wantSVG, want3MF, resDiv); err != nil {
				return fmt.Errorf("%s cover: %w", sz.Name, err)
			}
		}
		if wantAdaptor {
			if err := buildPart(sz, "adaptor", V0SlideOnPLA, wantSTL, wantSVG, want3MF, resDiv); err != nil {
				return fmt.Errorf("%s adaptor: %w", sz.Name, err)
			}
		}
	}
	return nil
}

func selectParts(arg string) (block, cover, adaptor bool, err error) {
	switch arg {
	case "block":
		return true, false, false, nil
	case "cover":
		return false, true, false, nil
	case "adaptor":
		return false, false, true, nil
	case "all":
		return true, true, true, nil
	case "both": // legacy: block + cover
		return true, true, false, nil
	default:
		return false, false, false, fmt.Errorf("unknown -part %q (try block, cover, adaptor, all)", arg)
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

func buildPart(sz V0Size, partLabel string, build partBuilder, wantSTL, wantSVG, want3MF bool, resDiv uint) error {
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
		if err := write3MF(shape, baseName+".3mf", sz.Name+" "+partLabel, resDiv); err != nil {
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

func write3MF(shape glbuild.Shader3D, filename, partName string, resDiv uint) error {
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
	renderer, err := glrender.NewOctreeRenderer(cpusdf, resolution, bufSize)
	if err != nil {
		return fmt.Errorf("octree: %w", err)
	}
	tris, err := glrender.RenderAll(renderer, cpusdf.VecPool())
	if err != nil {
		return fmt.Errorf("render: %w", err)
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
	fmt.Printf("wrote %s (%d triangles)\n", filename, len(tris))
	return nil
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
	if err := svgslice.WriteSliceLabelled(baseName+"_cutaway_axial.svg", sdf3, axial, 256, 256, axialLabels(sz)); err != nil {
		return err
	}
	fmt.Printf("wrote %s_cutaway_axial.svg\n", baseName)

	if y, ok := firstOffAxisY(sz.Holes); ok {
		offAxis := axial
		offAxis.Origin = ms3.Vec{Y: y}
		name := fmt.Sprintf("%s_cutaway_screwY%g.svg", baseName, y)
		if err := svgslice.WriteSliceLabelled(name, sdf3, offAxis, 256, 256, axialLabels(sz)); err != nil {
			return err
		}
		fmt.Printf("wrote %s\n", name)
	}

	midZ := (bb.Min.Z + bb.Max.Z) / 2
	top := svgslice.Plane{
		Origin: ms3.Vec{Z: midZ},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Y: 1},
		UMin:   bb.Min.X - pad, UMax: bb.Max.X + pad,
		VMin: bb.Min.Y - pad, VMax: bb.Max.Y + pad,
	}
	if err := svgslice.WriteSliceLabelled(baseName+"_cutaway_top.svg", sdf3, top, 256, 256, topLabels(sz)); err != nil {
		return err
	}
	fmt.Printf("wrote %s_cutaway_top.svg\n", baseName)
	return nil
}

// axialLabels annotates the axial cutaway with Wi / Wo / H.
//
// Plane is U=X, V=Z. Block trapezoid: small face Wi at v=0 (wall,
// bottom of SVG), large face Wo at v=H (top of SVG). Cover is the
// inverse: Wo at v=0, Wi at v=H. The same labels work for either
// because the trapezoid silhouette is identical, just up/down-flipped.
func axialLabels(sz V0Size) []svgslice.Label {
	wMax := sz.Wo
	if sz.Wi > wMax {
		wMax = sz.Wi
	}
	side := wMax/2 + 1.5 // right of the part outline
	return []svgslice.Label{
		{U: 0, V: sz.H + 1.0, Text: fmt.Sprintf("Wo = %g mm", sz.Wo)},
		{U: 0, V: -1.0, Text: fmt.Sprintf("Wi = %g mm", sz.Wi)},
		{U: side, V: sz.H / 2, Text: fmt.Sprintf("H = %g", sz.H), Anchor: "start"},
	}
}

// topLabels annotates the top-down cutaway. The plane is U=X, V=Y at
// z=H/2 (mid-height for the block, mid-height for the cover). The
// octagonal cross-section there has width across flats = (Wi+Wo)/2.
func topLabels(sz V0Size) []svgslice.Label {
	mid := (sz.Wi + sz.Wo) / 2
	return []svgslice.Label{
		{U: 0, V: mid/2 + 1.0, Text: fmt.Sprintf("at z = H/2: %g mm across flats", mid)},
	}
}

func firstOffAxisY(holes []ms2.Vec) (float32, bool) {
	for _, h := range holes {
		if h.Y != 0 {
			return h.Y, true
		}
	}
	return 0, false
}
