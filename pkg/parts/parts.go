// Package parts defines the Part abstraction used to describe physical
// items — both external (catalogue fasteners, off-the-shelf inserts)
// and built (3D-printed components). Each Part exposes two SDF forms:
//
//   - Shape: the true catalogue/CAD geometry, used for visualisation
//     and as the positive form when the part itself is being made.
//
//   - Insert: an oversized version of the same shape, sized for
//     clearance when subtracted from a host part. Tolerance controls
//     the slack — different printer materials need different fits.
//
// Composite and Exploded helpers assemble positioned parts into a
// single SDF for as-built and exploded-view documentation.
package parts

import (
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)

// Part is anything that can produce both a true-scale solid and an
// oversized "insert" form for subtraction from a host.
type Part interface {
	Shape(bld *gsdf.Builder) (glbuild.Shader3D, error)
	Insert(bld *gsdf.Builder, tol Tolerance) (glbuild.Shader3D, error)
}

// Tolerance describes how much an Insert is enlarged beyond the true
// part to leave room for printer error and easy assembly. Values are
// totals: Radial grows the diameter by Radial mm; Axial extends the
// length by Axial mm (typically split between the two ends — exact
// distribution is implementer-defined).
type Tolerance struct {
	Radial float32 // mm
	Axial  float32 // mm
}

// Material-tuned starting points. Re-tune per Part if fit matters; FDM
// printers vary by ~0.1 mm so these are conservative.
var (
	TolerancePLA  = Tolerance{Radial: 0.20, Axial: 0.20}
	TolerancePETG = Tolerance{Radial: 0.30, Axial: 0.25}
	ToleranceABS  = Tolerance{Radial: 0.25, Axial: 0.20}
	ToleranceSLA  = Tolerance{Radial: 0.05, Axial: 0.05}
)

// Block is a Lx×Ly×Lz cuboid centred at the origin. Insert grows the
// cuboid by tol.Radial in X and Y and tol.Axial in Z.
type Block struct{ Lx, Ly, Lz float32 }

func (b Block) Shape(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	return bld.NewBox(b.Lx, b.Ly, b.Lz, 0), bld.Err()
}
func (b Block) Insert(bld *gsdf.Builder, tol Tolerance) (glbuild.Shader3D, error) {
	return bld.NewBox(b.Lx+tol.Radial, b.Ly+tol.Radial, b.Lz+tol.Axial, 0), bld.Err()
}

// Cylinder is an axis-aligned cylinder along Z, radius R and full
// height H, centred at the origin. Insert grows the radius by
// tol.Radial/2 (so diameter grows by tol.Radial) and the height by
// tol.Axial.
type Cylinder struct{ R, H float32 }

func (c Cylinder) Shape(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	return bld.NewCylinder(c.R, c.H, 0), bld.Err()
}
func (c Cylinder) Insert(bld *gsdf.Builder, tol Tolerance) (glbuild.Shader3D, error) {
	return bld.NewCylinder(c.R+tol.Radial/2, c.H+tol.Axial, 0), bld.Err()
}

// Placement positions a Part at an offset in assembly space. Tolerance
// is consulted only when the placement is used as an insert (clearance
// hole) — for true-shape assembly it's ignored.
type Placement struct {
	Part      Part
	Offset    ms3.Vec
	Tolerance Tolerance
}

// Composite returns host with each insert subtracted (as its oversized
// Insert form) and each part unioned on top (as its true Shape form).
// Use it for as-built renderings of an assembly: a printed host with
// clearance holes for fasteners, plus the fasteners themselves drawn
// at true scale where they sit.
//
// inserts and parts may overlap conceptually — the same screw might
// appear in both lists if you want to render the printed host's hole
// AND the screw inside it. Pass it once in inserts for the hole and
// once in parts for the visible screw.
func Composite(bld *gsdf.Builder, host glbuild.Shader3D, parts, inserts []Placement) (glbuild.Shader3D, error) {
	out := host
	for _, ins := range inserts {
		s, err := ins.Part.Insert(bld, ins.Tolerance)
		if err != nil {
			return nil, err
		}
		s = bld.Translate(s, ins.Offset.X, ins.Offset.Y, ins.Offset.Z)
		out = bld.Difference(out, s)
	}
	for _, p := range parts {
		s, err := p.Part.Shape(bld)
		if err != nil {
			return nil, err
		}
		s = bld.Translate(s, p.Offset.X, p.Offset.Y, p.Offset.Z)
		out = bld.Union(out, s)
	}
	return out, nil
}

// Exploded returns host plus each placement's true Shape, with each
// shape additionally translated by axis · step · (i+1) so the parts
// fan out along that axis. Use for exploded-view assembly diagrams.
// axis should be a unit vector; step is in mm.
func Exploded(bld *gsdf.Builder, host glbuild.Shader3D, placements []Placement, axis ms3.Vec, step float32) (glbuild.Shader3D, error) {
	out := host
	for i, p := range placements {
		s, err := p.Part.Shape(bld)
		if err != nil {
			return nil, err
		}
		d := step * float32(i+1)
		s = bld.Translate(s,
			p.Offset.X+axis.X*d,
			p.Offset.Y+axis.Y*d,
			p.Offset.Z+axis.Z*d,
		)
		out = bld.Union(out, s)
	}
	return out, nil
}
