# Universal Mount

A standardised, 3D-printable mounting system. Anything that hangs off a wall, ceiling, or another printed object — pipes, cables, shelves, hooks, light fittings — connects through a single shared interface.

Two design properties drive everything else:

1. **45° outer slope on every external face.** A separate cover slips over a mount and fully encloses it when not loaded — no exposed hardware. The same 45° rule means a mount feature *integrated into another printed object* prints support-free in any of six orthogonal orientations: top, bottom, side. No overhang exceeds 45°.
2. **8-fold rotational indexing.** Adaptors clock to any of 8 positions in 45° steps after install, like the *Rotatable v3* reference but built into the standard interface rather than added on top.

The system is intended to live in [go3dp](https://codeberg.org/hum3/go3dp), generated from Go using [`github.com/soypat/gsdf`](https://github.com/soypat/gsdf) (the same library used in `cmd/BrabantiaPin`), one STL per part. This file is the design specification; the Go implementation follows in a later round.

## 1. Review of existing designs

Two adjacent systems — AngelR's *Modular Wall Mounting System* and *Multiboard / MultiBuild* — sit outside the main comparison table because they each need separate treatment. AngelR is the closest analogue to the design this document specifies and gets a dedicated subsection (§1.2). Multiboard is a different category — a whole-wall ecosystem with a commercial monetisation model and no first-class single-mount-point story — and is documented for completeness in §1.3.

### 1.1 Comparison: single-mount Printables designs

| # | Design | Two-part | Mating | Rotation | Lock | Print orient. constraint | Size range | Cover |
|---|---|---|---|---|---|---|---|---|
| 1 | [Strong Universal Wall Mount](https://www.printables.com/model/1156481-strong-universal-wall-mount) (VC Design) | Yes | Slim flat-back tab + slot | None | Friction / screw | Single orientation (flat back) | One small (4 g) | No |
| 2 | [Universal Wall Mount Small](https://www.printables.com/model/1281413-universal-wall-mount-small) (Blake Unsell) | Yes | Tab + slot | None | 5 mm wood-screw head + slot | Single orientation | S/M/L series | No |
| 3 | [Universal Wall Mount Rotatable v3](https://www.printables.com/model/1274249-universal-wall-mount-rotatable-v3) (RC_3DWorks) | Yes | Disk-in-housing | 8 × 45° | Toothed disc + screw | Single orientation | One (~Ø50, 19 g) | No — outer wall is straight, no clean shroud |
| 4 | [Universal Wall Mount Easy to Print](https://www.printables.com/model/926124-universal-wall-mount-easy-to-print) (Firstlayer Oy) | Yes | Rectangular tab + slot (90° walls — square overhangs) | None | Slide friction; M5 countersink to wall | Both parts upright | One (60 × 20 × 10 mm, screws 32 mm pitch) | No |

**Commentary.**

- **#1 / #2 / #4** are simple two-part flat-and-tab designs. Easy to print, easy to use, but no rotation, no cover, and the mount sits proud of the wall in any orientation other than its single intended one. None can be embedded into the top, bottom, *and* side of another printed object without redesigning the mating geometry.
- **#4 (Easy to Print) specifically.** Useful as a starting point but has two concrete problems this design fixes: (a) the tab-and-slot has 90° walls — they print as square overhangs in any orientation other than the intended upright, and the inside of the slot traps support material; and (b) the slide axis runs vertically, so the attachment is held in place by gravity alone — a knock upward will lift an unloaded attachment off the wall. The puck specified in §2 swaps both: 45° trapezoidal tapers throughout, and a horizontal push-on with a clamp screw rather than a gravity-only slide.
- **Strength-by-multiple-edges variants.** Several Printables remixes of #1, #2 and #4 increase pull-out and torsional resistance by adding *multiple parallel mating edges* — a top + bottom tab, or two nested rectangular shoulders, or a pair of side rails framing the slot. The principle is sound: more engaged perimeter for the same footprint gives more strength. The U-tube puck specified here applies the same principle in a different geometry: instead of two or three parallel rectangular tabs, the engagement is a single closed octagonal sleeve, which gives an *eight*-edged perimeter of contact along the full sleeve depth `H`. Strength scales with `H` and with `t_sleeve`, both already in the parameter table.
- **#3 (Rotatable v3)** is the closest reference for the rotation requirement. It has the 8-position toothed disc this design borrows. Its weakness is a vertical outer wall: nothing covers it, and it cannot be reused as a feature on the side of a printed object without supports.
- **#1 (Strong Universal Wall Mount) — close, but inverted.** The two-part flat-back-with-mating-feature architecture is essentially what this design adopts; #1 is the most architecturally similar reference of the four. The gender is, however, the wrong way round: #1 puts a male tab on the wall and a female slot on the attachment, leaving the wall protrusion exposed when nothing is mounted. This design flips the gender — the **female socket is recessed into the wall plate, and the male spigot lives on the adaptor** — so an unloaded mount presents only a flush plug on the wall (see §2 and §4.3), and any printed object that needs an integrated mounting point grows a male spigot (a simple 45° tapered protrusion) rather than an enclosed cavity (which would be harder to integrate into arbitrary geometry).

### 1.2 Closest match: Modular Wall Mounting System (AngelR)

[*Modular Wall Mounting System*](https://www.printables.com/model/152930-modular-wall-mounting-system) by AngelR is the closest existing design to what this document specifies, and the most useful single point of comparison.

**Strengths:**
- **Two-part with snap-in attachments.** A wall-mounted base accepts a range of attachments (default hook, filament-spool holder with bearings, broom holder, drill holder) via a sturdy snap-in connection that audibly clicks home and is released by pressing tabs.
- **Strength-tested.** The default hook is rated to **at least 30 kg** by the author. Real numbers, not aspirations — uncommon in this category.
- **Hides screws.** Once the attachment is snapped on, the mounting screws are concealed beneath it. Visually clean.
- **Open design files.** SolidWorks source for the snap-in base is included, allowing custom attachments.

**Gaps versus this design:**
- **Mating geometry is a flat tab with snap detents — no trapezoidal taper.** The tab's rectangular profile means it has 90° walls; embedded into another printed object face-down or on its side, those walls become unprintable overhangs. The 45° trapezoidal taper of the puck specified here is specifically chosen to remove this constraint, so the mating feature can be moulded directly into the bottom or side of any printed object without supports.
- **No rotation indexing.** The snap is single-orientation (or at most a fixed flip) — the user cannot reorient the attachment after install. The 8-position rotation index of the puck design here addresses this directly.
- **No cover.** When the base is unused (no attachment snapped on) the bare base is visible on the wall. The puck design here adds a cover that fully encloses an unloaded mount.
- **Mounting-screw size unspecified.** The published documentation does not give the recommended screw size, head type, hole spacing or wall-anchor compatibility — users have to size them empirically. This design specifies M3 / M4 / M5 / M6 with countersink Ø3.5 holes on a stated pitch (§4.3).

In short: AngelR demonstrates that a strength-rated snap-in modular system works, validates the basic two-part architecture, and is missing exactly three things — printable-in-any-orientation mating geometry, rotation indexing, and a cover. This document picks those up.

### 1.3 Adjacent ecosystem: Multiboard / MultiBuild

[*Multiboard / MultiBuild*](https://multibuild.io/) is the dominant modular wall system on the market, but it is **not really a wall-mount-comparison** — it is an ecosystem positioned as a whole-wall organisation grid:

- **Whole-wall scale, not single-point.** Multiboard is a tessellated honeycomb panel that covers an area of wall, with hundreds of accessories (bolts, snaps, pegboard adapters, Gridfinity-style boxes, shelves) attaching anywhere on the grid. This is a different problem domain from "mount one specific item to one specific spot on a wall".
- **No first-class wall-mount story.** The board itself attaches to the wall via 8 mm wall-standoff fasteners and the user mounts the board first; there is no analogue of a small standalone wall plate for a single hook or pipe clip.
- **Commercial / rental positioning.** The Multiboard project promotes itself as a creator-supported ecosystem and is moving toward subscription / rental income from the design files and accessory library, rather than fully open-licence Printables-style sharing. That distinguishes it culturally and licensing-wise from the rest of the systems in §1.1.
- **No cover concept.** The honeycomb grid is the visible, intended aesthetic — the design philosophy is "the wall *is* the mounting system" rather than "hide the mounting hardware".

Multiboard is worth knowing about as the high-end of the modular-mount market, but it is not the gap this design tries to fill.

### 1.4 Gap this design fills

No reviewed single-mount-point system simultaneously gives: (a) cover-able exterior, (b) 8-position rotation indexing, (c) printability in any of six orthogonal orientations as an integrated feature on another printed object, (d) explicit mounting-screw specification. The 45° tapered U-tube puck (§2) is the unifying solution.

## Build plan

The full U-tube puck described in §2 onward is the long-term target (**v1**). The first iteration (**v0**) ships a much simpler shape — a single solid octagonal frustum block, fastened directly to the wall by axial screws — together with the underlying tooling that the v1 design will reuse. v0 is small and self-contained; it delivers screw-modelling primitives, an SVG-cutaway slicer, and one printable wall block + trivial adaptor pair to validate the toolchain end-to-end before committing to the more complex puck.

### v0 — solid octagonal frustum + screws

```
SIDE VIEW (cross section)                       LARGE FACE (outer, looking at the wall)

   ▲ wall screws                              _________________
   │ pass through                            /  o           o  \    ← 2 wall-screw
   │ from outer face                        /                   \     countersink heads,
   │ into wall                             |          ⊕          |    recessed below the
   │                                       |                     |    outer face
   ────────────────────                     \                   /   ← 1 central M4
   \                  /  ← outer (large)    \_________________/      threaded insert
    \                /     face              (large face = Wo
     \              /                         across flats)
      \  (solid    /
       \  block)  /        ← 45° on all
        \        /            8 sides
         \      /
          \    /          ← inner (small) face
           ────              against the wall
            ▼

  Geometry:
  - octagonal frustum, 45° on all 8 sides (printable in any orientation)
  - inner face width across flats: Wi (sits flush against the wall)
  - outer face width across flats: Wo (larger; the adaptor-mounting surface)
  - height H = (Wo − Wi)/2 (forced by the 45° slope)
  - the small face is on the wall: less wall contact, less material, and the
    visible face on the room side is the larger one — better for adaptor
    bolting and (in v1) for a flush cover

  Default size (M):
  - Wi = 24 mm  (small face, on wall)
  - Wo = 40 mm  (large face, out)
  - H  = 8 mm
  - 1–2 wall screws: M4 countersunk, heads sunk into the outer (large) face,
    shafts pass axially through the block into the wall
  - 1 central M4 threaded insert pressed into the outer face for adaptor
    mounting (the adaptor's mating surface covers the wall-screw heads)

  No mating cavity. No rotation indexing. No cover. (All deferred to v1.)
```

What v0 delivers in code:

- **`screws.go`** — parametric primitives for M3 / M4 / M5 / M6 with countersunk and pan-head variants, plus threaded-insert pockets. Reused by every subsequent design.
- **`svgcut.go`** — marching-squares-to-SVG slicer for `gsdf`. Takes a `glbuild.Shader3D`, a slice plane (point + normal), evaluates the resulting 2D field on a grid, extracts contour line segments via marching squares, writes them as an SVG `<path>`. Uses gsdf's existing CPU evaluator (`gleval.NewCPUSDF3` or its 2D equivalent) — does not need GPU.
- **`v0.go`** — the wall block and a trivial flat-plate adaptor. Two STLs and a small SVG cutaway diagram package per print.

### v1 — U-tube puck with 8-position rotation indexing and cover

The full spec is in §2 onward. Adopted once v0 is printed and validated.

### v2 — example catalogue

Pipe saddles (15/22 mm copper, 32/40 mm waste), hooks, J-hook, light shelves — see §5.

### Library + tooling decisions

- **Modelling library:** `github.com/soypat/gsdf`. The intro paragraph already calls this out.
- **SVG output:** `gsdf` does not emit SVG natively — its renderer outputs STL and raster images only. Rather than fall back to `deadsy/sdfx` (which has built-in SVG via marching squares + `render/svg.go`), v0 adds a small ~100-line `svgcut.go` helper on top of gsdf's 2D evaluator. This keeps the project on a single modelling library and gives us a slicer we control.
- **STL output:** unchanged from the existing `cmd/BrabantiaPin` pattern — `glrender.NewOctreeRenderer` → `WriteBinarySTL`.

## 2. Design parameters (v1 target)

Working design: **Octagonal Puck (U-tube section)** — a hollow octagonal sleeve, closed at the back (the side fixed to the wall plate or to the integrated object) and open at the front (the mating face). Outer walls and inner cavity walls both taper at 45°. Mating is socket-into-socket: an adaptor's female pocket slides over a wall plate's spigot, with deep telescoping contact along the full sleeve length, not just disc-on-disc. The U-tube cross-section gives substantially more bending and torsional stiffness for the same material than a solid frustum, while the 45° outer slope and 45° inner cavity walls preserve the orientation-agnostic printability rule.

This is locked in as the primary; runners-up are recorded at the bottom of this section.

### Standardised contract

| Symbol | Meaning | S | M | L | XL |
|---|---|---|---|---|---|
| `W` | Width across flats of the open (front) octagon | 25 mm | 40 mm | 60 mm | 90 mm |
| `H` | Sleeve depth (back face → mating-face edge) | 10 mm | 16 mm | 24 mm | 36 mm |
| `θ` | Outer taper angle from vertical | 45° | 45° | 45° | 45° |
| `Wb` | Width of the (small) closed back face | W − 2H tan θ = W − 2H | W − 2H | W − 2H | W − 2H |
| `t_sleeve` | Sleeve wall thickness (constant from back to front) | 2.0 mm | 2.4 mm | 3.0 mm | 4.0 mm |
| `D_cavity` | Cavity opening width across flats (= W − 2·t_sleeve) | 21 mm | 35.2 mm | 54 mm | 82 mm |
| `Wb_cav` | Cavity floor width (closed end of cavity) | Wb − 2·t_sleeve | same | same | same |
| `t_pitch` | Indexing teeth on cavity-floor face | 8 teeth on Ø(D_cavity·0.55) | same | same | same |
| `t_h` | Tooth height (trapezoidal, 90° included) | 1.0 mm | 1.2 mm | 1.5 mm | 2.0 mm |
| `d_screw` | Centre clamp screw clearance | M3 | M4 | M5 | M6 |
| `t_wall_cov` | Cover wall thickness | 1.6 mm | 1.6 mm | 2.0 mm | 2.4 mm |
| `c_mate` | Spigot-to-socket clearance (telescoping fit) | 0.2 mm | 0.2 mm | 0.3 mm | 0.3 mm |
| `c_cover` | Cover-to-puck clearance (slip fit) | 0.3 mm | 0.3 mm | 0.4 mm | 0.4 mm |
| `s_shrink` | Print shrinkage compensation | 1/0.999 | 1/0.999 | 1/0.999 | 1/0.999 |

Material: PLA primary, PETG/ABS for outdoor or load-bearing use. Print: 0.4 mm nozzle, 0.2 mm layer height (0.15 mm for the disc), 4 perimeters on load-bearing parts. Shrinkage value matches `cmd/plugs` convention.

The interface contract is: **any two parts of the same size class share the same `W`, `H`, `t_sleeve`, and the same 8-tooth pattern at the same `t_pitch` radius on the cavity floor**. A size-M adaptor never connects to a size-L plate; sizes do not interoperate. Adaptors of the same size always do.

### Why a U-tube section

Replacing the original solid frustum with a hollow U-tube serves three goals:

1. **Stiffness.** A spigot telescoping into a matching socket along its full sleeve depth `H` resists bending and twisting along a long lever arm. A flat-disc-on-flat-disc joint can only resist these via the central clamp screw and the small footprint of the indexing teeth.
2. **Material economy.** A `t_sleeve = 2.4 mm` shell with a 35 mm cavity uses roughly 40 % of the plastic of a solid M frustum.
3. **Orientation-agnostic printability is preserved.** Both outer walls and inner cavity walls slope at 45°, so no overhang exceeds 45° in any of six orthogonal print orientations.

The puck is not mirror-symmetric — it has a defined "back" (closed, carrying the central screw thread and the 8 indexing teeth on the cavity floor) and a "front" (open, the mating face).

**Gender convention (revised).** Wall plates carry the **female socket** (cavity recessed into the front face of the plate). Adaptors carry the **male spigot** (sleeve protruding from the back of the adaptor body). The cover is a flush plug that fills the wall-plate's empty socket when no adaptor is mounted. This convention has three consequences:

- An unloaded wall plate presents only a flush plug on the wall — no protrusion, no exposed mating geometry.
- Any printed object that needs an integrated mounting point grows a male spigot, which is a simple 45°-tapered protrusion. This is much easier to embed into arbitrary printed geometry than an enclosed cavity would be — and crucially, it prints support-free in any of six orientations because every external face is at 45°.
- The wall plate becomes thicker (it must contain the socket depth `H` plus structural margin — total thickness ≈ `H + 4 mm`), at the cost of more material per plate. This is a deliberate trade in favour of visual cleanliness on the wall and ease of embedding the male feature in other prints.

Both parts independently print in any orientation.

### Runners-up (rejected, recorded for future)

- **Design B — single-sided frustum.** Flat back, 45° front taper. Simpler and stronger as a wall plate, but the back face is a 90° overhang when integrated face-down on another object.
- **Design C — dovetail bar with H × W groove.** A long bar with a 45°-walled trapezoidal slot. Closest to the user's "groove with H × W" wording. Dropped because it gives no rotation indexing and locks the adaptor's orientation along the bar axis. Still appealing as a *secondary* interface for long runs (think a kitchen utility rail) — listed in the catalogue under "future".

## 3. Size series

```
   S          M             L                XL
  ┌─┐       ┌──┐          ┌──────┐         ┌────────┐
  └─┘       └──┘          └──────┘         └────────┘
 W=25      W=40            W=60              W=90
 H=10      H=16            H=24              H=36
 M3        M4              M5                M6
```

Use cases:
- **S** — small hooks, cable clips, light items < 1 kg.
- **M** — pipe saddles up to 22 mm, picture mounts, small shelves; default size.
- **L** — heavier brackets, tool holders, items 5–15 kg.
- **XL** — bicycles, wall-mounted equipment, items 15–40 kg.

## 4. Design sketches

All sketches are orthographic, fixed-pitch ASCII. Dimensions cited assume size **M** (`W=40, H=16, t_sleeve=2.4, D_cavity=35.2, Wb=8`) but the geometry generalises by size symbol.

### 4.1 Octagonal puck — male spigot (used on a wall plate, or extruded out of any printed object)

```
TOP VIEW (looking into the open mating face)        SIDE VIEW (cross section)

         ______________                              ←────  W=40  ────→
        / ╔══════════╗ \                            ┌──────────────────┐  ← open mating face
       /  ║          ║  \                           │\                /│  (cavity opening)
      /   ║  cavity  ║   \                          │ \              / │
      \   ║  + screw ║   /                          │  \   cavity   /  │  ← inner walls
       \  ║          ║  /                           │   \ (45° tap) /   │    taper 45°
        \ ╚══════════╝ /                            │    \         /    │    inward
         ──────────────                             │     \  ⊕    /     │
                                                    │      ───────      │  ← cavity floor:
       outer edge: 8-sided, W=40 across flats       │     ╲         ╱   │    8 teeth +
       inner edge: 8-sided, D_cavity=35.2           │      ╲       ╱    │    M4 thread
       sleeve thickness t_sleeve=2.4 (constant)     │       ╲     ╱     │
       8 trapezoidal teeth, t_h=1.2 mm,             │        ╲   ╱      │  ← 45° outer taper
         on Ø19 circle (= D_cavity·0.55)            │         ╲ ╱       │
       M4 thread emerges through cavity floor       └──────────┴────────┘  ← Wb=8 closed back
                                                            ↑
                                                    closed back face on host
                                                    (plate / printed object)
                                                    H=16 deep
```

The puck is a hollow octagonal sleeve. Outer skin tapers 45° outward from the closed back to the open front; inner cavity wall tapers 45° inward (cavity narrower at the floor than at the opening). Sleeve wall thickness `t_sleeve` is constant. Eight raised trapezoidal teeth (`t_h=1.2 mm`) sit on the cavity floor on a Ø19 mm circle, giving 45° rotation indexing. The central M4 clamp screw passes through the cavity floor.

### 4.2 Octagonal puck — female socket (the mating cavity in any adaptor)

```
SIDE VIEW (cross section)

  open mating face (toward wall plate)
       ↓
   ┌───────────────────────────────────┐  ← cavity opening: W+c_mate = 40.2 across flats
   │  ╲                            ╱   │
   │   ╲                          ╱    │
   │    ╲                        ╱     │  ← inner walls match the male spigot's outer
   │     ╲                      ╱      │    taper plus c_mate=0.2 mm clearance
   │      ╲────────────────────╱       │  ← cavity-floor face
   │       │ teeth on this face │      │    8 teeth aligned with male's cavity-floor
   │       │ mate the male's    │      │    teeth — co-planar at full insertion
   │       │ cavity-floor teeth │      │
   │       │       ⊕            │      │  ← M4 clearance through to adaptor body
   │      /                      \     │
   │     /                        \    │  ← below: adaptor body continues
   └────/                          \───┘
```

A female socket is the *negative* of a male spigot, scaled out by `c_mate` clearance. The adaptor's mating face presents this cavity, which swallows the wall-plate spigot to depth `H`. The teeth on the adaptor's cavity floor meet the teeth on the wall-plate spigot's cavity floor: their tooth circles are co-planar when the spigot is fully seated, giving 8-position rotation lock. The same M4 screw threads through the wall plate, up through the adaptor's cavity floor, and pulls the two together.

### 4.3 Wall plate (size M) — female socket recessed into the front face

```
FRONT VIEW                                  SIDE VIEW (cross section)

  ┌──────────────────────────┐              ┌──────────────────────────┐
  │  o                    o  │              │                          │  ← front face (out)
  │       ┌──────────┐       │              │       ╲          ╱       │
  │       │ ╲      ╱ │       │              │        ╲        ╱        │
  │       │  ╲ ⊕  ╱  │       │      →       │         ╲      ╱         │  ← female socket
  │       │   ╲╱     │       │              │          ╲────╱          │    (4.2)
  │       └──────────┘       │              │          │teeth│         │    recessed into
  │  o                    o  │              │          │  ⊕  │         │    the plate
  └──────────────────────────┘              │__________│_____│_________│  ← back face (wall)
                                            │  ⊕    ⊕    ⊕    ⊕        │   countersink screws

  Outer plate: 60 × 60 mm, 20 mm thick (= H + 4 mm structural margin)
  Screw holes: 4 × Ø3.5 countersink (heads recessed into the front face)
                  on a 40 mm square pattern
  Socket: female puck cavity (4.2), 16 mm deep, opens flush with the front face
  Centre: M4 threaded insert sunk into the back of the plate; the spigot of any
          adaptor seats against the cavity floor and is pulled tight by the M4 screw
          which threads from the spigot's open end into the wall plate's insert
```

Print orientation: front face down on the bed (open socket flush on bed). The cavity prints upward as a chimney with 45° tapered walls — no supports. The cavity floor (with its 8 indexing teeth) is at the top of the print and prints as small bridges spanning <2 mm. The back of the plate (which sits against the wall) prints last as the top layer.

### 4.4 Cover (size M) — flush plug for the wall plate's empty socket

```
SIDE VIEW (cross section)                 TOP VIEW

  ┌────────────────────────────┐              ┌──────────────────────┐
  │ ───────────────────────── │              │                      │
  │                           │              │     ┌──────────┐     │
  │  ╲                    ╱   │              │     │          │     │
  │   ╲                  ╱    │   ← 45°      │     │ flat top │     │
  │    ╲                ╱     │     taper    │     │          │     │
  │     ╲              ╱      │              │     └──────────┘     │
  │      ╲────────────╱       │              │                      │
  │       │ small  │          │   ← optional └──────────────────────┘
  │       │ finger │          │     finger
  │       │ recess │          │     recess        flat top sits flush
  │       │   ⊕    │          │     for         with wall plate's front face
  │       └────────┘          │     removal
  └────────────────────────────┘

  - the cover IS a male spigot per 4.1, plus a flat top instead of an
    attachment-side body
  - top face: flat octagon, sized to sit flush with the wall plate front face
              (= W − 2·c_cover wide across flats)
  - small concave fingernail recess in the top face, off-centre, for prying
    the cover out
  - centre: optional M4 clearance hole (so the cover can be retained by
    the same screw that would clamp an adaptor — prevents loss)
  - print: flat-top down on the bed, spigot points up, no supports
```

When no adaptor is fitted, the cover plugs into the wall plate's empty socket, taper-on-taper, friction-held by `c_cover=0.3 mm` clearance. Optional retention: tighten an M4 screw through the cover's centre hole into the wall plate's insert, so the cover cannot be knocked out.

### 4.5 Rotating adaptor — generic shape (male spigot + attachment geometry)

```
SIDE VIEW (cross section)                   The male-spigot side seats into the wall
                                            plate's female socket. The attachment side
        ┌──────────────────┐                carries custom geometry — pipe saddle,
        │   attachment     │                hook, J-hook, flat plate, etc.
        │    geometry      │
        │                  │                The 8 teeth on the spigot's cavity floor
        ├──────────────────┤                engage the wall plate's 8 teeth: the
        │     ╔══════╗     │                adaptor can be clocked to any of 8
        │     ║      ║     │                positions.
        │     ║cavity║     │
        │     ║  +   ║     │
        │     ║teeth ║     │
        │     ║  ⊕   ║     │  ← M4 clearance through to attachment side
        │      ╲════╱      │
        │       ╲  ╱       │  ← 45° outer taper of spigot
        └────────╲╱────────┘
        spigot points downward (toward wall plate, into its socket)

  Central M4 clearance hole runs all the way through the adaptor — the same
  screw threads from the cavity floor into the wall plate's M4 insert and
  clamps the assembly tight.
```

The adaptor's mating side is *always* a male spigot; the wall plate's mating side is *always* a female socket. This convention removes any ambiguity about print orientation and eliminates the need for two flavours of every adaptor.

### 4.6 Pipe saddle adaptor — 15 mm copper (size M)

```
FRONT VIEW                                  SIDE VIEW (cross section)

         _______                                ┌─────────────────┐
       /         \   ← 15.4 mm ID saddle        │      _____      │  ← saddle wraps 240°
      /           \    cradles 15 mm pipe       │    /       \    │
     │   (pipe)    │   with 0.4 mm clearance    │   │  pipe   │   │
      \           /                             │    \_______/    │
       \_________/                              ├─────────────────┤  ← arm to adaptor body
            │                                   │                 │
            │                                   ├─────────────────┤
        ╔════╧════╗                             │   ╔═════════╗   │
        ║ spigot  ║   ← M-size male spigot      │   ║ cavity  ║   │
        ║         ║     (4.1) sticks down       │   ║   +     ║   │  ← male spigot (4.1)
        ╚═════════╝     toward the wall plate   │   ║ teeth   ║   │    sticks toward wall
                                                │   ║   ⊕     ║   │    plate's socket
                                                │    ╲═══════╱    │
                                                │     ╲     ╱     │  ← 45° outer taper
                                                └──────╲___╱──────┘

  - saddle inner Ø: 15.4 mm     (15 mm pipe + 0.4 mm fit clearance)
  - saddle outer Ø: 22.0 mm     (3.3 mm wall, 4 perimeters at 0.4 mm nozzle)
  - saddle wrap: 240°           (pipe snaps in past the 120° opening)
  - arm: 12 mm wide × 5 mm tall, transitions to the adaptor body
  - adaptor body: 44 × 44 × 5 mm + male spigot (4.1) protruding 16 mm = 26 mm total depth
  - mating side: male spigot per 4.1 (M size)
  - central M4 clearance: pipe centreline → through adaptor body
                          → through spigot cavity floor → into wall-plate thread
  - total height (saddle top to spigot tip): ~33 mm
```

Print orientation: spigot tip down on the bed (the spigot's open end flush on the bed). The spigot prints upward as a 45°-tapered tower, with the cavity prints inside it as a chimney — no supports. The adaptor body, arm and saddle print on top. The saddle's 240° wrap leaves a 120° bridge at its peak, which prints reliably in PLA at 0.2 mm layers.

## 5. Examples catalogue

Marked **[code]** ships in the first Go-implementation round. **[spec]** is documented here only and implemented later.

| Part | Sizes | Description |
|---|---|---|
| Wall plate | S/M/L/XL | **[code]** Flat back + male puck. 4 corner countersink holes + central clamp screw thread. |
| Cover | S/M/L/XL | **[code]** Slip-on hood, encloses an unused puck flush with the wall. |
| Disc / female adaptor base | S/M/L/XL | **[code]** Female-pocket primitive that other adaptors extend. Reusable Go function. |
| Pipe saddle — 15 mm copper | M | **[code]** First worked example. As specified in §4.6. |
| Pipe saddle — 22 mm copper | M | **[spec]** Same arm + pocket as 15 mm; saddle ID = 22.4 mm, OD = 27 mm. |
| Pipe saddle — 32 mm waste | L | **[spec]** Size-L pocket. Saddle ID = 32.5 mm, OD = 38 mm. |
| Pipe saddle — 40 mm waste | L | **[spec]** Saddle ID = 40.5 mm, OD = 47 mm. |
| Generic flat-plate adaptor | S/M/L/XL | **[spec]** Flat 30×30 (M) plate on the front of the female pocket. Users graft custom geometry onto it. |
| Hook | S/M/L | **[spec]** J-shaped wire-form hook, ID 12/20/30 mm. |
| Cable clip | S | **[spec]** U-clip over the puck pocket; takes 6–10 mm cable. |
| Light shelf adaptor | M | **[spec]** 80 × 50 mm shelf, 5 mm thick, on top of the female pocket; rated to 2 kg. |
| Dovetail rail (Design C) | parametric H×W | **[spec, future]** Long rail with trapezoidal groove for kitchen rails, pipe runs etc. Listed for completeness — not the same interface as the puck, but would share the project. |

## 6. Open questions

- **Tooth profile.** Trapezoidal teeth chosen for printability, but their exact width-vs-height ratio will affect both lock strength and ease of engagement. The first Go implementation should make `t_h` and the tooth width independent constants so they can be tuned after the first print test.
- **Clamp screw vs. captive nut.** Wall plate could either embed a nut on the back face (captive) or use a heat-set insert. Captive nut is cheaper; heat-set is cleaner but requires the user to own inserts. Default: design for both — a hex pocket on the back accepts a standard nut, *and* the same hole accepts an M4 heat-set insert.
- **Cover retention.** Pure friction on the 45° taper may be too loose with `c_cover = 0.3`. Consider a 0.1 mm interference at the closed-top end only, so the cover snaps on at the last 1 mm of travel.
- **Sleeve depth `H` vs. width `W`.** The current `H ≈ 0.4·W` is a guess. Deeper sleeves give more bending stiffness but consume more material and slow printing. Re-evaluate after printing the M set under load.
- **Anti-pull-out.** With pure 45° tapers and a clamp screw, axial pull-out resistance is set entirely by the screw. Consider an optional small lip / undercut at the cavity opening, only on adaptors that don't use a clamp screw, for a snap-only variant.
- **3MF bundling.** Per-part STLs are the standard go3dp output. A `task um:3mf` step that bundles parts with default material assignments would be useful for the multi-material case (e.g. cover in a contrast colour). Defer until first round of code is shipped.
- **Project-level CLAUDE.md.** As the catalogue grows past ~5 parts, a `cmd/UniversalMount/CLAUDE.md` will be worth adding so future sessions can pick up the parameter table without re-deriving it. Add when first triggered.
- **Shrinkage.** `1/0.999` is the project default from `cmd/plugs`. Confirm with the first printed M wall plate + cover pairing — if cover is too tight, reduce; too loose, increase.
