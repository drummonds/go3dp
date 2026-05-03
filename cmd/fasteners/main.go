// Command fasteners renders the catalogue entries from pkg/fasteners as
// 3MF and SVG cutaway under docs/.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	threemf "codeberg.org/hum3/gsdf-3mf"

	"codeberg.org/hum3/go3dp/pkg/fasteners"
	"codeberg.org/hum3/go3dp/pkg/svgslice"

	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
	"github.com/soypat/gsdf/gleval"
	"github.com/soypat/gsdf/glrender"
)

func main() {
	var (
		part   = flag.String("part", "spax-4x20", "catalogue entry: spax-3_5x16 | spax-4x20")
		render = flag.String("render", "schematic", "fidelity: schematic | threaded")
		outDir = flag.String("out", "docs", "output directory")
		resDiv = flag.Uint("resdiv", 200, "resolution = bbox diagonal / resdiv")
	)
	flag.Parse()

	if err := run(*part, *render, *outDir, *resDiv); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run(part, render, outDir string, resDiv uint) error {
	var bld gsdf.Builder
	var screw fasteners.WoodScrew
	switch part {
	case "spax-3_5x16":
		screw = fasteners.Spax3_5x16
	case "spax-4x20":
		screw = fasteners.Spax4x20
	default:
		return fmt.Errorf("unknown -part %q (try spax-3_5x16 | spax-4x20)", part)
	}

	var (
		shape glbuild.Shader3D
		err   error
		suf   string
	)
	switch render {
	case "schematic":
		shape, err = screw.Schematic(&bld)
		suf = ""
	case "threaded":
		shape, err = screw.Threaded(&bld)
		suf = "_threaded"
	default:
		return fmt.Errorf("unknown -render %q (try schematic | threaded)", render)
	}
	if err != nil {
		return fmt.Errorf("build %s: %w", render, err)
	}
	if berr := bld.Err(); berr != nil {
		return fmt.Errorf("builder: %w", berr)
	}
	if err := glbuild.ShortenNames3D(&shape, 12); err != nil {
		return fmt.Errorf("shorten names: %w", err)
	}

	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}
	base := filepath.Join(outDir, part+suf)

	cpusdf, err := gleval.NewCPUSDF3(shape)
	if err != nil {
		return fmt.Errorf("cpu sdf: %w", err)
	}
	const bufSize = 4096
	cpusdf.VecPool().SetMinAllocationLen(bufSize)

	tris, err := renderTriangles(shape, cpusdf, resDiv, bufSize)
	if err != nil {
		return fmt.Errorf("triangles: %w", err)
	}
	if err := writeMF(base+".3mf", screw.Name, tris); err != nil {
		return fmt.Errorf("3mf: %w", err)
	}
	fmt.Printf("wrote %s.3mf (%d triangles)\n", base, len(tris))

	if err := writeSVG(base+"_xz.svg", cpusdf, screw); err != nil {
		return fmt.Errorf("svg: %w", err)
	}
	fmt.Printf("wrote %s_xz.svg\n", base)
	return nil
}

func renderTriangles(s glbuild.Shader3D, cpusdf *gleval.SDF3CPU, resDiv uint, bufSize int) ([]ms3.Triangle, error) {
	resolution := s.Bounds().Diagonal() / float32(resDiv)
	renderer, err := glrender.NewOctreeRenderer(cpusdf, resolution, bufSize)
	if err != nil {
		return nil, err
	}
	return glrender.RenderAll(renderer, cpusdf.VecPool())
}

func writeMF(filename, name string, tris []ms3.Triangle) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	parts := []threemf.Part{{
		Name:      name,
		Color:     threemf.Color{R: 0xB0, G: 0x70, B: 0x20, A: 0xFF},
		Triangles: tris,
	}}
	return threemf.Write(f, parts, threemf.UnitMillimeter)
}

func writeSVG(filename string, sdf3 gleval.SDF3, s fasteners.WoodScrew) error {
	plane := svgslice.Plane{
		Origin: ms3.Vec{},
		U:      ms3.Vec{X: 1},
		V:      ms3.Vec{Z: 1},
		UMin:   -s.DHead, UMax: s.DHead,
		VMin: -s.OverallLength - 1, VMax: 1,
	}
	grid := 400
	return svgslice.WriteSlice(filename, sdf3, plane, grid, grid)
}
