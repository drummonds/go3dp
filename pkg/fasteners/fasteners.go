// Package fasteners models screws, bolts, and similar fasteners as
// abstract Go types that can be rendered by the gsdf SDF builder at two
// fidelity levels:
//
//   - Schematic: cheap bicone/cylinder approximation, suitable for boolean
//     subtraction (clearance holes, countersinks) where the consumer only
//     cares about the envelope of material to be removed.
//   - Threaded: an accurate helical-thread mesh, suitable for visualisation
//     and assembly diagrams.
//
// Concrete catalogue entries (e.g. a Spax 4×20 wood screw) live in
// per-vendor files (spax.go, din7997.go, mcmaster.go) and reference an
// abstract type from this package.
package fasteners

// Head is the visible top of a fastener.
type Head uint8

const (
	HeadCountersunk       Head = iota // 90° conical, sits flush
	HeadRaisedCountersunk             // dome on a 90° cone (oval)
	HeadPan                           // domed cylinder
	HeadRound                         // hemispherical
	HeadHex                           // external hex (bolt)
	HeadHexFlange                     // hex with integral washer
	HeadButton                        // low-profile dome
	HeadCheese                        // tall straight-sided cylinder
)

// Drive is the recess shape used to apply torque.
type Drive uint8

const (
	DriveSlot      Drive = iota // single straight slot
	DrivePhillips               // cruciform (Phillips H)
	DrivePozidriv               // cruciform with extra ribs (Pz)
	DriveTorx                   // 6-point star (TX)
	DriveHexSocket              // internal hex (Allen)
	DriveRobertson              // square
	DriveNone                   // no drive (e.g. external hex bolt)
)

// ThreadKind classifies the threadform.
type ThreadKind uint8

const (
	ThreadWood        ThreadKind = iota // single-lead, sharp-tipped, often partial
	ThreadMachineISO                    // ISO metric (M3, M4 ...)
	ThreadSelfTapping                   // for sheet metal / plastic, full-thread
	ThreadLag                           // coarse coach screw
)

// Render selects the SDF fidelity level when a Fastener is built.
type Render uint8

const (
	RenderSchematic Render = iota // bicone head + cylinder shank, fast
	RenderThreaded                // helical threads, slow but accurate
)
