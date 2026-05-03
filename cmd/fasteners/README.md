# fasteners

Renders the catalogue entries from [`pkg/fasteners`](../../pkg/fasteners/) as
3MF (with `gsdf-3mf`) and SVG cross-sections (XZ midplane) into `docs/`.

```sh
go run .                                 # default: spax-4x20, schematic
go run . -render threaded                # helical thread render
go run . -part spax-4x20 -resdiv 300     # finer mesh
go run . -out out                        # custom output dir
```

Outputs:

- `docs/<part>.3mf` — coloured solid mesh (schematic), loads in MeshLab,
  PrusaSlicer, Bambu Studio, etc.
- `docs/<part>_xz.svg` — marching-squares cutaway through the screw axis.
- `docs/<part>_threaded.3mf` and `<part>_threaded_xz.svg` — same but with
  the helical thread modelled (file size ~ 1.5×, triangle count ~ 1.3×).

## Two fidelity levels

The library models each screw at two fidelity levels:

| Level | Method | Cost | Use for |
|---|---|---|---|
| **Schematic** | `WoodScrew.Schematic(bld)` — bicone head + cyl shank + cone tip, single `Revolve` | cheap (≈10 SDF ops) | clearance / countersink cutouts in printed parts |
| **Threaded** | `WoodScrew.Threaded(bld)` — head bicone + (optional) plain shank + `threads.Screw` ISO 60° at `WoodScrew.ThreadPitch` + cone tip | expensive (helical mesh) | visualisation, exploded assembly drawings |

Boolean subtractors only need the schematic. The threaded form exists for
documentation and visualisation, not for printing — wood screws aren't 3D
printed.

## Spax vs DIN 7997 — what differs

The first concrete catalogue entry is a **Spax-style** screw. Spax is not a
public standard; it's a registered trademark and product family from
Altenloh, Brinck & Co. The closest formal standard for traditional wood
screws is **DIN 7997** (countersunk-head wood screw, single slot drive).
The two look superficially similar but differ on five practical points:

| Aspect | DIN 7997 (traditional) | Spax (modern) |
|---|---|---|
| Drive | Slot only | Torx (TX10–TX40 by size) |
| Thread | Single-lead, full thread to head | Single-lead, **partial** thread on lengths ≥ 25 mm |
| Point | Plain conical, ≈30°–45° angle | "4CUT" serrated point — drills its own pilot in softwood |
| Coating | Plain steel or zinc | Wax-coated; **≈30 % less driving torque** |
| Material | Mild steel | Hardened carbon steel |
| Head profile | 90° flat-top countersunk | 90° flat-top countersunk *with ribs under the head* on some sizes |

The abstract `WoodScrew` type expresses both styles by varying:

- `Drive` field (`DriveSlot` for DIN 7997, `DriveTorx` for Spax),
- `ThreadLength` ≠ `OverallLength - HeadDepth` for partial thread (Spax),
- `PointLength` for the conical taper (4CUT vs plain comes in at the
  threaded-render layer, not the schematic).

The schematic SDF is identical for both — the differences only show up in
the threaded render and in the metadata. That's intentional: anyone using
a screw as a *cutout* in a printed part doesn't care whether it's a Spax
or a DIN 7997. Anyone making a *visualisation* does.

A DIN 7997 catalogue entry will land in `pkg/fasteners/din7997.go` once
the schematic API has settled.
