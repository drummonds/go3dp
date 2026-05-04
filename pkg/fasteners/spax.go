package fasteners

// Spax is the registered trademark of ABC Verbindungstechnik (now Altenloh,
// Brinck & Co.) for their countersunk-torx-drive partial-thread wood screw
// family. There is no public ISO/DIN standard that matches Spax exactly;
// the dimensions below come from the manufacturer's published datasheets.
//
// Distinguishing features (vs traditional DIN 7997 wood screws):
//   - 4CUT serrated point — drills its own pilot hole into softwood.
//   - Partial thread on lengths ≥ 25 mm; full thread on shorter sizes.
//   - Torx (TX) drive at all sizes; never slot or Pozidriv from the factory.
//   - Wax coating reduces driving torque ≈ 30 %.
//   - Hardened carbon steel, typically with a yellow zinc or "Wirox" finish.
//
// See cmd/fasteners/README.md for a side-by-side comparison with DIN 7997.

// Spax3_5x16 — Spax countersunk torx wood screw, 3.5 mm × 16 mm.
// Smallest size in regular use; pairs with the v0-XS Universal Mount block.
var Spax3_5x16 = WoodScrew{
	Name:          "Spax 3.5×16",
	DShank:        3.5,
	DHead:         7.0,
	HeadDepth:     1.75, // (DHead-DShank)/2 — true 90° countersink
	OverallLength: 16.0,
	ThreadLength:  12.25, // OverallLength - HeadDepth - PointLength (full thread)
	PointLength:   2.0,
	ThreadPitch:   1.5,
	Drive:         DriveTorx,
	Standard:      "no formal standard (Spax)",
	Vendors: []Vendor{
		{Name: "Spax", SKU: "0191010350163", URL: ""},
	},
}

// Spax4x20 — Spax countersunk torx wood screw, 4 mm × 20 mm. As-built
// dimensions per Spax product datasheet: full-thread on this length,
// 90° countersunk head Ø 8 mm, 4CUT serrated point ≈ 2 mm.
var Spax4x20 = WoodScrew{
	Name:          "Spax 4×20",
	DShank:        4.0,
	DHead:         8.0,
	HeadDepth:     2.0,
	OverallLength: 20.0,
	ThreadLength:  16.0, // OverallLength - HeadDepth - PointLength (full thread)
	PointLength:   2.0,
	ThreadPitch:   1.75,
	Drive:         DriveTorx,
	Standard:      "no formal standard (Spax)",
	Vendors: []Vendor{
		{Name: "Spax", SKU: "0191010400203", URL: ""},
	},
}
