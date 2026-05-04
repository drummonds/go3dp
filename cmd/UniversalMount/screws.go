package main

import (
	"errors"

	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)

// ScrewSize captures the geometry of a screw + countersink that we want
// to remove from a printed part. All values are in millimetres.
type ScrewSize struct {
	DShaft     float32 // clearance hole Ø (e.g. 4.3 for M4 with fit clearance)
	DHead      float32 // countersink head Ø
	HeadDepth  float32 // depth of the countersink (90° cone → (DHead-DShaft)/2)
	InsertD    float32 // brass heat-set insert pocket Ø
	InsertLen  float32 // brass heat-set insert pocket depth
	ThreadName string  // e.g. "M4"
}

// Standard M-series sizes — clearance fit (close), DIN 7991 countersunk heads.
// Head Ø values are nominal +0.4 mm clearance for easy seating.
var (
	M3 = ScrewSize{DShaft: 3.4, DHead: 6.5, HeadDepth: 1.6, InsertD: 4.0, InsertLen: 5.0, ThreadName: "M3"}
	M4 = ScrewSize{DShaft: 4.3, DHead: 9.0, HeadDepth: 2.4, InsertD: 5.6, InsertLen: 5.0, ThreadName: "M4"}
	M5 = ScrewSize{DShaft: 5.3, DHead: 11.0, HeadDepth: 2.9, InsertD: 6.4, InsertLen: 9.5, ThreadName: "M5"}
	M6 = ScrewSize{DShaft: 6.4, DHead: 13.0, HeadDepth: 3.3, InsertD: 8.0, InsertLen: 12.7, ThreadName: "M6"}
)

// CountersunkHoleZ builds a negative shape that, when subtracted from a part,
// leaves a countersunk-screw hole oriented along the world -Z axis.
//
// Coordinates: the head opening sits at z=0 (large face of the cone, full
// DHead Ø), the cone narrows to DShaft Ø at z=-HeadDepth, and the shaft
// continues straight down to z=-length.
//
// `length` is the total hole depth from the part surface (z=0) to the tip
// of the screw cavity. `length` must be ≥ s.HeadDepth.
//
// The countersink cone is approximated as a regular octagonal frustum.
// At the diameters used here that is well within 3D-print resolution.
func CountersunkHoleZ(bld *gsdf.Builder, s ScrewSize, length float32) (glbuild.Shader3D, error) {
	if length < s.HeadDepth {
		return nil, errors.New("CountersunkHoleZ: length must be >= HeadDepth")
	}

	// Shaft cylinder: full length, centred on origin then translated so it
	// spans z ∈ [-length, 0].
	shaft := bld.NewCylinder(s.DShaft/2, length, 0)
	shaft = bld.Translate(shaft, 0, 0, -length/2)

	// Countersink head: octagonal frustum tapering inward as z decreases.
	// We oversize the small face by 0.05 mm to ensure clean union with the
	// shaft cylinder (no slim ring gap from numerical noise).
	head, err := OctagonalFrustum(bld, s.DShaft+0.1, s.DHead, s.HeadDepth)
	if err != nil {
		return nil, err
	}
	// OctagonalFrustum places small face at z=0, large face at z=h.
	// Translate so the large face (head opening) sits at z=0.
	head = bld.Translate(head, 0, 0, -s.HeadDepth)

	return bld.Union(shaft, head), nil
}

// ThreadedInsertPocketZ builds a negative cylindrical pocket for a brass
// heat-set threaded insert, opening at z=0 and extending down -Z to the
// insert's full length. Subtract from a part with Difference.
func ThreadedInsertPocketZ(bld *gsdf.Builder, s ScrewSize) glbuild.Shader3D {
	pocket := bld.NewCylinder(s.InsertD/2, s.InsertLen, 0)
	return bld.Translate(pocket, 0, 0, -s.InsertLen/2)
}
