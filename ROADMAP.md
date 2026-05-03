# go3dp Roadmap

Status snapshot, organised by area. Things in *italics* are user-facing decisions still open.

## Universal Mount

- [x] **v0 size series** — XS / S / M / L solid octagonal frustums, 1 / 2 / 3 / 3 wood-screw fixings
- [x] **v0 covers** — matching octagonal-frustum cap with same screw-hole pattern; stacks on the block
- [x] **v0 slide-on adaptor (first family member)** — `V0SlideOn` lateral sleeve with half-octagonal closed end + 5 mm extension on the open end. PLA defaults at 0.15 mm tolerance / 1.5 mm wall thickness.
- [ ] **Per-material tolerance presets** — currently only `V0AdaptorPLA`. Add `V0AdaptorPETG` (~0.20 mm), `V0AdaptorABS` (~0.25 mm), `V0AdaptorSLA` (~0.05 mm). First validate by printing a PLA tolerance sweep and confirming the slide-on fit.
- [ ] **More adaptor types** — share the same `V0Adaptor*` mating pocket geometry but vary the outer "utility" portion: hooks, brackets, pipe saddles (15/22 mm copper, 32/40 mm waste), J-hooks, light shelves. (Was previously called the "v2 example catalogue".)
- [ ] **WASM-exposed parametric build** — let the user adjust `Tolerance` (and pick `V0Size`) live in-browser, regenerate the 3MF, and download. `gsdf` already runs on WASM; the only new piece is a small UI form + glue.
- [ ] **Strength testing** — bench rig + load measurement on each variant. Inputs: pull-out from PLA wall, shear at the screw heads, torsional resistance. *Open: which screw spacing / wall material gives the best strength-to-cost ratio across the four sizes.*
- [ ] **v1 — U-tube puck** — full design in `cmd/UniversalMount/README.md` §2 onward. Telescoping mating, 8-position rotation indexing, integrated cover. Picks up after v0 is validated in print.
- [ ] **v2 — example catalogue** — folded into the "More adaptor types" item above; tracking there.

## Fasteners

- [x] **`pkg/fasteners` taxonomy** — `WoodScrew`, `Head`, `Drive`, `ThreadKind`, `Render` enums; vendor metadata
- [x] **WoodScrew.Schematic** — bicone head + cylindrical shank + cone tip (cheap, for cutouts)
- [x] **WoodScrew.Threaded** — helical thread via `gsdf/forge/threads` (visual only, ISO threadform)
- [x] **WoodScrew.WallCutout** — through-block cone+cylinder negative for clearance + countersink holes
- [x] **Catalogue: Spax 3.5 × 16, Spax 4 × 20**
- [ ] **DIN 7997 catalogue entry** — slot drive, full thread, plain steel. Differences are documented in `cmd/fasteners/README.md`; the Go entry is straightforward once a real use case appears.
- [ ] **Drive recess negatives** — torx / posidriv / slot SDF profiles to subtract from the head bicone, so screw heads in renders aren't smooth domes.
- [ ] **Custom WoodThreader** — current threaded render uses ISO 60° at the screw's nominal Ø, which produces a thinner core than real wood screws. Add a `threads.Wood` Threader for accurate visual / strength simulation. *Skip until visual fidelity actually bites.*
- [ ] **Machine screws** — lift `cmd/UniversalMount/screws.go` (M3–M6 DIN 7991 + heat-set inserts) into `pkg/fasteners/machine.go` once the API has settled.

## Documentation site

- [x] **Repo-level docs site** — `index.md` landing page at root, `task docs:build` aggregates per-section content into `docs/`, deployed via task-plus + statichost as `h3-go3dp`
- [x] **3D in-browser previews** — Three.js + 3MFLoader, `<div class="model-viewer">` blocks in markdown, hydrated by `web/viewer.js`. Both Universal Mount and Fasteners pages.
- [x] **Self-hosted Three.js** — files mirrored under `web/three/`, no jsdelivr dependency at runtime. `task docs:fetch-three` re-downloads on version bumps.
- [x] **Favicon** — `web/favicon.svg` octagonal-frustum silhouette, copied to `docs/` by `docs:build`.
- [ ] **Web-resolution renders** — current `-resdiv 200` produces oversampled 3MFs (e.g. 800 KB for v0-XS, 4.5 MB for the matching STL) because resolution is bbox-relative and tiny parts get tiny cells. Add a separate web-output target with a fixed mm-resolution (≈ 0.3 mm) so browser previews download in ~1/4 the bytes without touching the print-quality STL. *Defer.*
- [ ] **glTF migration** — eventually swap 3MF→glTF + Three.js→`<model-viewer>` for a smaller JS payload. Requires writing a `gsdf-gltf` analogue. Only worth it if total docs JS becomes a problem.

## Strength + materials

- [ ] **Material datasheet entries** — PLA, PETG, ABS, PA12 with shrinkage / pull-out / shear numbers; pair with the Universal Mount strength tests.
- [ ] **Wall-material guidance** — separate notes per substrate: drywall + plug, masonry + plug, timber stud, plastic. Recommended size class per substrate.

## Tooling

- [ ] **Refactor pre-existing legacy `cmd/*` directories** — `BrabantiaPin`, `cap.go`, `ButtonCleaner`, `breadboard`, `plugs` are all on the retired `deadsy/sdfx` API or pre-refactor `gsdf` and don't compile against current dependencies. Either port to current `gsdf` or delete; either way they shouldn't sit in this state.
- [ ] **`pkg/fastener-rendering`** — small shared helper for the `gsdf → glrender → 3mf/stl` pipeline duplicated between `cmd/UniversalMount/main.go` and `cmd/fasteners/main.go`. Wait for a third caller before extracting.
