# Build pipeline

How a Go source file in `cmd/` becomes a printable STL, a coloured 3MF, and a set of dimensioned SVG cutaways.

![go3dp build pipeline](process-flow.svg)

## Stages

1. **Source** — each part lives in its own `cmd/<part>/` directory as a small `main` package. Geometry is defined in plain Go using `gsdf` primitives plus the shared `pkg/fasteners` catalogue (`WoodScrew` schematic / threaded / wall-cutout variants). `main.go` parses flags (`-size`, `-part`, `-out`, `-mesh`) and dispatches to the build pipeline below.
2. **SDF tree** — [`gsdf`](https://github.com/soypat/gsdf) composes primitives (boxes, cylinders, octagonal prisms, screws, etc.) with boolean operations into a single signed-distance-function tree. No mesh yet — just a function that, given a 3D point, returns the signed distance to the surface.
3. **CPU evaluation** — `gleval.NewCPUSDF3` compiles the tree for fast batched evaluation on the CPU. The same evaluator drives both the mesh renderers (3D grid sampling) and the SVG slicer (2D plane sampling).
4. **Mesh rendering** — two algorithms, picked by the `-mesh` flag:
   - `OctreeRenderer` (marching cubes) — original pipeline, larger triangle counts, tolerant of slicer quirks.
   - `DualContourRenderer` (dual contouring) — sharper edges on flat faces, fewer triangles before merging.
5. **Optional planar merge** — `pkg/meshopt.PlanarMerge` collapses fans of coplanar triangles back to single polygons (re-triangulated as fans). Default for 3MF; cuts adaptor file size by ~95%, and blocks/covers by 15-35%. Skipped for `-mesh=octree` and `-mesh=dc`.
6. **Cross-section slicing** — `pkg/svgslice` evaluates the SDF on a user-supplied plane (origin + U/V basis vectors), runs marching squares to extract contour line segments, and writes them as an SVG `<path>` with optional dimension annotations and labels.
7. **Writers**:
   - STL — `gsdfaux.RenderShader3D` → binary STL via `glrender.NewOctreeRenderer`.
   - 3MF — [`gsdf-3mf`](https://codeberg.org/hum3/gsdf-3mf) zips a 3D Manufacturing Format file with one part per colour. Used for browser previews and for slicers that respect material assignments.
   - SVG — `svgslice.WriteSliceLabelled` writes a millimetre-scale SVG, suitable both for documentation and for round-tripping into 2D CAD.

## Mesher choice

The `-mesh` flag for 3MF output trades file size against pipeline complexity. The `task v0:3mf` shortcut uses the default (`merge`).

| `-mesh`  | Triangles | When to use                                                           |
|----------|-----------|------------------------------------------------------------------------|
| `merge`  | Smallest  | **Default.** Coplanar fans collapsed. Adaptors shrink ~95%.            |
| `dc`     | Medium    | Same triangle count as `merge` before the merge step — useful to bisect a slicer issue blamed on the merge pass. |
| `octree` | Largest   | Original pipeline. Use if a slicer rejects the merged output (potential T-junctions across faces). |

## Where each artefact comes from

| Artefact            | Pipeline                                                          |
|---------------------|-------------------------------------------------------------------|
| `*.stl`             | gsdf → gleval → octree → `WriteBinarySTL`                         |
| `*.3mf` (default)   | gsdf → gleval → dual contouring → planar merge → gsdf-3mf         |
| `*.3mf` (`-mesh=dc`)| gsdf → gleval → dual contouring → gsdf-3mf                        |
| `*.3mf` (`-mesh=octree`) | gsdf → gleval → octree → gsdf-3mf                            |
| `*_cutaway_*.svg`   | gsdf → gleval → svgslice (marching squares + labels)              |

## Source

- Diagram source: [process-flow.d2](process-flow.d2) — rendered with [d2](https://d2lang.com).
