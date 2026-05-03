# Fasteners catalogue

Wood screws and (later) machine screws as Go types. Each catalogue entry has:

- **As-built dimensions** — DShank, DHead, HeadDepth, OverallLength, ThreadLength, PointLength, ThreadPitch — taken from the manufacturer's datasheet, not stylised.
- **Two render fidelities**:
  - **Schematic** — bicone head + cylindrical shank + cone tip, single revolve. Cheap (~10 SDF ops). Use for boolean subtraction (clearance / countersink cutouts in printed parts).
  - **Threaded** — same body plus a helical thread modelled with `gsdf/forge/threads` at the catalogue entry's `ThreadPitch`. Expensive but visually realistic.
- **Vendor metadata** — Name + SKU + (optional) URL so the abstract model traces back to real parts.

See the [package source](https://codeberg.org/hum3/go3dp/src/branch/main/pkg/fasteners) for the Go API and the [Spax vs DIN 7997 notes](README.md) for why Spax-style is the default.

## Catalogue entries

### Spax 3.5 × 16

The smallest size in regular use; pairs with the `v0-XS` Universal Mount block.

#### Schematic (cheap, for cutouts)

<div class="columns is-vcentered">
<div class="column">

| | |
|---|---|
| DShank | 3.5 mm |
| DHead | 7.0 mm |
| HeadDepth | 1.75 mm (true 90° countersink) |
| OverallLength | 16 mm |
| ThreadLength | 12.25 mm (full thread) |
| PointLength | 2.0 mm |
| ThreadPitch | 1.5 mm |
| Drive | Torx |

**Cross section**

![Spax 3.5×16 schematic cross-section](spax-3_5x16_xz.svg)

**Downloads**

- [3MF](spax-3_5x16.3mf)
- [Cross section SVG](spax-3_5x16_xz.svg)

</div>
<div class="column">

<div class="model-viewer" data-model="spax-3_5x16.3mf" style="height: 380px; width: 100%; border: 1px solid #ddd;"></div>

</div>
</div>

#### Threaded (visualisation)

<div class="columns is-vcentered">
<div class="column">

**Cross section**

![Spax 3.5×16 threaded cross-section](spax-3_5x16_threaded_xz.svg)

**Downloads**

- [3MF](spax-3_5x16_threaded.3mf)
- [Cross section SVG](spax-3_5x16_threaded_xz.svg)

</div>
<div class="column">

<div class="model-viewer" data-model="spax-3_5x16_threaded.3mf" style="height: 380px; width: 100%; border: 1px solid #ddd;"></div>

</div>
</div>

### Spax 4 × 20

#### Schematic

<div class="columns is-vcentered">
<div class="column">

| | |
|---|---|
| DShank | 4.0 mm |
| DHead | 8.0 mm |
| HeadDepth | 2.0 mm |
| OverallLength | 20 mm |
| ThreadLength | 16 mm (full thread) |
| PointLength | 2.0 mm |
| ThreadPitch | 1.75 mm |
| Drive | Torx |

**Cross section**

![Spax 4×20 schematic cross-section](spax-4x20_xz.svg)

**Downloads**

- [3MF](spax-4x20.3mf)
- [Cross section SVG](spax-4x20_xz.svg)

</div>
<div class="column">

<div class="model-viewer" data-model="spax-4x20.3mf" style="height: 380px; width: 100%; border: 1px solid #ddd;"></div>

</div>
</div>

#### Threaded

<div class="columns is-vcentered">
<div class="column">

**Cross section**

![Spax 4×20 threaded cross-section](spax-4x20_threaded_xz.svg)

**Downloads**

- [3MF](spax-4x20_threaded.3mf)
- [Cross section SVG](spax-4x20_threaded_xz.svg)

</div>
<div class="column">

<div class="model-viewer" data-model="spax-4x20_threaded.3mf" style="height: 380px; width: 100%; border: 1px solid #ddd;"></div>

</div>
</div>

## Building

```
task fasteners:all              # render every catalogue entry
task fasteners:spax             # Spax 4×20 schematic only
task fasteners:spax-threaded    # Spax 4×20 threaded only
task fasteners:docs:build       # this page, rendered to docs/index.html
```

## Why these dimensions?

See [README.md](README.md) for the philosophy: **Spax style is the default** for new wood-screw entries because it's what's actually stocked at builders' merchants (UK + DE). DIN 7997 (slot drive, full-thread, plain steel) is the textbook standard but rarely sold; a DIN 7997 catalogue entry will land alongside Spax once a use case appears.

<script type="importmap">
{
  "imports": {
    "three": "../js/three/three.module.js",
    "three/addons/": "../js/three/addons/"
  }
}
</script>
<script type="module" src="../js/viewer.js"></script>
