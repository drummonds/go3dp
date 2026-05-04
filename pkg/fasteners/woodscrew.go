package fasteners

import (
	"errors"
	"fmt"
	"math"

	"github.com/soypat/geometry/ms2"
	"github.com/soypat/geometry/ms3"
	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/forge/threads"
	"github.com/soypat/gsdf/glbuild"
)

// WoodScrew describes a wood / chipboard / construction screw at the
// abstract level: countersunk head + parallel shank + conical point. All
// dimensions are millimetres.
//
// The schematic SDF places the head's outer face at world z = 0 and the
// tip at z = -OverallLength, so the screw drives in the -Z direction. This
// matches the convention used by CountersunkHoleZ in cmd/UniversalMount
// so the same coordinate system flows through the codebase.
type WoodScrew struct {
	Name string // catalogue name, e.g. "Spax 4×20"

	DShank        float32 // major thread diameter ≈ shank diameter
	DHead         float32 // head outside diameter
	HeadDepth     float32 // axial depth of countersink (head top to shank top)
	OverallLength float32 // total axial length, head face to tip
	ThreadLength  float32 // axial length of the threaded portion (excluding any unthreaded shank)
	PointLength   float32 // length of the conical point at the tip
	ThreadPitch   float32 // distance between successive thread crests (mm)

	Drive    Drive
	Standard string   // e.g. "DIN 7997", "ISO 14586", "no formal standard (Spax)"
	Vendors  []Vendor // concrete sources, may be empty
}

// Vendor links a screw to a real-world part listing.
type Vendor struct {
	Name string // e.g. "Screwfix", "McMaster-Carr"
	SKU  string // catalogue / part number
	URL  string // direct product page; may be omitted
}

// Schematic builds the cheap bicone-style SDF of the screw: a true 90°
// (or near-90°) conical head, parallel cylindrical shank, and a conical
// tip. Suitable for boolean subtraction (clearance + countersink cutouts)
// where actual thread geometry is irrelevant. See Threaded for the slow
// path with helical threads.
//
// Returned shape is centred on the world Z axis with the head's outer
// face at z = 0 and the tip at z = -s.OverallLength.
func (s WoodScrew) Schematic(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}

	// Half-profile in builder (X, Y) where X = radius and Y = axial
	// coordinate. After Revolve about builder Y, then a +π/2 rotation
	// about X, builder Y becomes world Z.
	//
	// Vertex order (counter-clockwise viewed from +Z in builder space):
	//   axis top -> head rim -> countersink toe -> point shoulder -> tip
	rHead := s.DHead / 2
	rShank := s.DShank / 2
	yHead := float32(0)
	yToe := -s.HeadDepth
	yPoint := -(s.OverallLength - s.PointLength)
	yTip := -s.OverallLength

	verts := []ms2.Vec{
		{X: 0, Y: yHead},
		{X: rHead, Y: yHead},
		{X: rShank, Y: yToe},
		{X: rShank, Y: yPoint},
		{X: 0, Y: yTip},
	}
	profile := bld.NewPolygon(verts)
	solid := bld.Revolve(profile, 0)
	// Builder Y maps to world Z under +π/2 rotation about world X (same
	// convention as cmd/UniversalMount/octagon.go trapezoidalPrism).
	solid = bld.Rotate(solid, float32(math.Pi/2), ms3.Vec{X: 1})
	return solid, nil
}

// WallCutout builds a cone-plus-cylinder negative shape suitable for
// subtracting from a printed part to make a clearance + countersink hole
// for this screw passing through it. Different from Schematic in two
// ways: the shaft is straight (no conical tip) and is generously sized
// for clearance, and the head opening is widened so the head sits
// `recess` below the surface.
//
// Coordinates: head opening at z = 0, shaft extends down to
// z = -throughLen. Origin is the centre of the head opening; translate
// the result to position it on a part.
//
// Shank clearance is 0.5 mm on diameter (matches the convention in
// cmd/UniversalMount/screws.go for machine-screw clearance fits).
func (s WoodScrew) WallCutout(bld *gsdf.Builder, throughLen, recess float32) (glbuild.Shader3D, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	if throughLen <= 0 {
		return nil, errors.New("WoodScrew.WallCutout: throughLen must be positive")
	}
	if recess < 0 {
		return nil, errors.New("WoodScrew.WallCutout: recess must be non-negative")
	}

	const shankClearance = 0.5 // total clearance on diameter, mm
	rShankCutout := (s.DShank + shankClearance) / 2
	// Cone widens above the head's actual rim by `recess` so the head
	// can be driven that far below the surface and still sit fully in
	// the cone (a 90° cone widens by exactly `recess` per `recess` of
	// extra depth, since DHead = DShank + 2·HeadDepth ⇒ 45° wall).
	rHeadCutout := s.DHead/2 + recess
	headDepthCutout := s.HeadDepth + recess

	// Half-profile in builder (X, Y), revolved about builder Y.
	// Cylinder runs the full throughLen so it picks up wherever the
	// cone is narrower; their union is the cutout.
	//
	// Vertex order CW (matches OctagonalFrustum convention).
	verts := []ms2.Vec{
		{X: 0, Y: 0},
		{X: rHeadCutout, Y: 0},
		{X: rShankCutout, Y: -headDepthCutout},
		{X: rShankCutout, Y: -throughLen},
		{X: 0, Y: -throughLen},
	}
	profile := bld.NewPolygon(verts)
	solid := bld.Rotate(bld.Revolve(profile, 0), float32(math.Pi/2), ms3.Vec{X: 1})
	return solid, nil
}

// Threaded builds the accurate helical-thread SDF: head bicone, optional
// unthreaded shank, helical thread of length s.ThreadLength using the ISO
// 60° threadform at s.ThreadPitch, and a conical point. Slow to mesh —
// expect ≥10× the triangle count of Schematic at the same resolution.
//
// The ISO threader is used as a reasonable visual approximation; real
// wood screws have a single-lead thread with a thinner core than ISO. The
// crest geometry (60° angle, sharp peak) is correct, the core diameter
// will be smaller than the actual screw.
//
// Coordinate convention matches Schematic: head face at z = 0, tip at
// z = -s.OverallLength.
func (s WoodScrew) Threaded(bld *gsdf.Builder) (glbuild.Shader3D, error) {
	if err := s.validate(); err != nil {
		return nil, err
	}
	if s.ThreadPitch <= 0 {
		return nil, errors.New("WoodScrew.Threaded: ThreadPitch must be positive")
	}

	rHead := s.DHead / 2
	rShank := s.DShank / 2

	// Head: revolved trapezoid (axis -> head rim -> shank rim -> back to axis).
	headProfile := bld.NewPolygon([]ms2.Vec{
		{X: 0, Y: 0},
		{X: rHead, Y: 0},
		{X: rShank, Y: -s.HeadDepth},
		{X: 0, Y: -s.HeadDepth},
	})
	head := bld.Rotate(bld.Revolve(headProfile, 0), float32(math.Pi/2), ms3.Vec{X: 1})

	// Threaded region: from threadTopZ down to threadBotZ. By default
	// thread runs from under the head to the start of the point; if
	// s.ThreadLength is shorter than that span, we add an unthreaded
	// shank between head and thread.
	threadBotZ := -(s.OverallLength - s.PointLength)
	maxThreadLen := -s.HeadDepth - threadBotZ // = OverallLength - HeadDepth - PointLength
	threadLen := s.ThreadLength
	if threadLen > maxThreadLen {
		threadLen = maxThreadLen
	}
	threadTopZ := threadBotZ + threadLen
	unthreadedLen := -s.HeadDepth - threadTopZ // ≥ 0

	// Optional unthreaded shank cylinder.
	var shank glbuild.Shader3D
	if unthreadedLen > 0 {
		shank = bld.NewCylinder(rShank, unthreadedLen, 0)
		shank = bld.Translate(shank, 0, 0, -s.HeadDepth-unthreadedLen/2)
	}

	// Threaded portion via gsdf/forge/threads.
	threader := threads.ISO{D: s.DShank, P: s.ThreadPitch, Ext: true}
	screw, err := threads.Screw(bld, threadLen, threader)
	if err != nil {
		return nil, fmt.Errorf("threads.Screw: %w", err)
	}
	screw = bld.Translate(screw, 0, 0, (threadTopZ+threadBotZ)/2)

	// Conical point: revolved triangle.
	tipProfile := bld.NewPolygon([]ms2.Vec{
		{X: 0, Y: threadBotZ},
		{X: rShank, Y: threadBotZ},
		{X: 0, Y: -s.OverallLength},
	})
	tip := bld.Rotate(bld.Revolve(tipProfile, 0), float32(math.Pi/2), ms3.Vec{X: 1})

	out := bld.Union(head, screw)
	if shank != nil {
		out = bld.Union(out, shank)
	}
	out = bld.Union(out, tip)
	return out, nil
}

func (s WoodScrew) validate() error {
	switch {
	case s.DShank <= 0 || s.DHead <= 0:
		return errors.New("WoodScrew: DShank and DHead must be positive")
	case s.DHead <= s.DShank:
		return errors.New("WoodScrew: DHead must exceed DShank")
	case s.HeadDepth <= 0:
		return errors.New("WoodScrew: HeadDepth must be positive")
	case s.OverallLength <= s.HeadDepth+s.PointLength:
		return errors.New("WoodScrew: OverallLength must exceed HeadDepth + PointLength")
	case s.PointLength < 0:
		return errors.New("WoodScrew: PointLength must be non-negative")
	}
	return nil
}
