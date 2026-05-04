# go3dp

3D-printable parts and fasteners modelled in Go using [`gsdf`](https://github.com/soypat/gsdf), exported as STL and 3MF via [`gsdf-3mf`](https://codeberg.org/hum3/gsdf-3mf).

This site is the navigation hub for the project's documentation. Each section below has its own design notes, parameter tables, and generated artefacts (STL meshes, 3MF previews, SVG cutaways).

## Project documents

- [Process flow](process-flow.html) — how Go source becomes STL, 3MF, and SVG (build pipeline diagram).
- [README](README.html) — project overview and links.
- [ROADMAP](ROADMAP.html) — what is built, what is planned, by area.

## Sections

### [Universal Mount](universal-mount/)

A standardised mounting system: an octagonal frustum block fixed to the wall by countersunk wood screws, with future support for sliding adaptors. Comes in four sizes (XS / S / M / L) for 1, 2 or 3 wall screws.

Status: **v0 size series** built and rendered. v1 (U-tube puck with rotation index and cover) and v2 (example catalogue — pipe saddles, hooks, shelves) deferred.

### [Fasteners catalogue](fasteners/)

Wood screws and machine screws as Go types, with two render fidelities — a cheap bicone schematic for boolean cutouts, and a slow but accurate helical-thread render for visualisation. Each catalogue entry carries vendor / SKU metadata so models trace back to real parts.

Currently in the catalogue: Spax 3.5 × 16, Spax 4 × 20.

## Older prints

Earlier 3D-printed parts that pre-date the current `gsdf`-based pipeline. Most are on Soypat's retired `deadsy/sdfx` API and don't compile against current dependencies; they're kept for reference and slated for port-or-delete (see [ROADMAP](ROADMAP.html) → *Tooling*).

Source lives under [`cmd/`](https://codeberg.org/hum3/go3dp/src/branch/main/cmd) on Codeberg:

| Part | Notes |
|---|---|
| [BrabantiaPin](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/BrabantiaPin) | Replacement pin for old Brabantia 40/50 L touch-lid bins. Legacy `sdfx`. |
| [epoxy_insert](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/epoxy_insert) | Wall insert for filled / lath-and-plaster walls where plasterboard plugs have failed. |
| [BoundaryWireInserter](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/BoundaryWireInserter) | Peg for installing a robot-mower boundary wire. |
| [bowlstacker](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/bowlstacker) | Stacking spacer for nesting bowls. |
| [breadboard](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/breadboard) | Bread cutting board. Legacy `sdfx`. |
| [ButtonCleaner](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/ButtonCleaner) | Cleaning aid for keypad buttons. Legacy `sdfx`. |
| [fridgedrainer](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/fridgedrainer) | Drip-tray drain for a fridge. |
| [gravelwasher](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/gravelwasher) | Aquarium-gravel rinsing aid. |
| [mieleadaptor](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/mieleadaptor) | Hose adaptor for a Miele appliance. |
| [plugs](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/plugs) | Generic plugs / caps. Legacy `sdfx`. |
| [stackSperator](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/stackSperator) | Spacer for separating stacked items. |
| [TorxFloorPlug](https://codeberg.org/hum3/go3dp/src/branch/main/cmd/TorxFloorPlug) | Floor plug with a Torx-driven head. |

## Source

| | |
|---|---|
| Source (Codeberg) | https://codeberg.org/hum3/go3dp |
| Mirror (GitHub) | https://github.com/drummonds/go3dp |
| Modelling library | [gsdf](https://github.com/soypat/gsdf) |
| 3MF writer | [gsdf-3mf](https://codeberg.org/hum3/gsdf-3mf) |
