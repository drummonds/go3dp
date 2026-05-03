# go3dp

3D-printable parts and fasteners modelled in Go using [`gsdf`](https://github.com/soypat/gsdf), exported as STL and 3MF via [`gsdf-3mf`](https://codeberg.org/hum3/gsdf-3mf).

This site is the navigation hub for the project's documentation. Each section below has its own design notes, parameter tables, and generated artifacts (STL meshes, 3MF previews, SVG cutaways).

## Sections

### [Universal Mount](universal-mount/)

A standardised mounting system: an octagonal frustum block fixed to the wall by countersunk wood screws, with future support for sliding adaptors. Comes in four sizes (XS / S / M / L) for 1, 2 or 3 wall screws.

Status: **v0 size series** built and rendered. v1 (U-tube puck with rotation index and cover) and v2 (example catalogue — pipe saddles, hooks, shelves) deferred.

### [Fasteners catalogue](fasteners/)

Wood screws and machine screws as Go types, with two render fidelities — a cheap bicone schematic for boolean cutouts, and a slow but accurate helical-thread render for visualisation. Each catalogue entry carries vendor / SKU metadata so models trace back to real parts.

Currently in the catalogue: Spax 3.5 × 16, Spax 4 × 20.

## Source

| | |
|---|---|
| Source (Codeberg) | https://codeberg.org/hum3/go3dp |
| Mirror (GitHub) | https://github.com/drummonds/go3dp |
| Modelling library | [gsdf](https://github.com/soypat/gsdf) |
| 3MF writer | [gsdf-3mf](https://codeberg.org/hum3/gsdf-3mf) |
