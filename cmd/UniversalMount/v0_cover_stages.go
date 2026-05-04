package main

import (
	"fmt"

	"github.com/soypat/gsdf"
	"github.com/soypat/gsdf/glbuild"
)

// V0Stage1ExplodeGap is the Z-axis separation between the block and
// the V0BlockCut cavity in the stage-1 exploded view, mm.
const V0Stage1ExplodeGap = 5.0

// V0Stage1 returns an exploded view of the v0 block at the origin and
// its mating cavity (V0BlockCut, with default PLA params) lifted along
// +Z so the relationship between the two is visible in cross-section
// and 3D viewers. Documentation render only — not a printable part.
//
// The cavity is translated by sz.H + V0Stage1ExplodeGap so a clear gap
// sits between the block's top face (z=H) and the cavity's bottom face
// (z=H+gap).
func V0Stage1(bld *gsdf.Builder, sz V0Size) (glbuild.Shader3D, error) {
	block, err := V0Block(bld, sz)
	if err != nil {
		return nil, fmt.Errorf("V0Stage1 %s block: %w", sz.Name, err)
	}
	cutaway, err := V0BlockCutDefault(bld, sz)
	if err != nil {
		return nil, fmt.Errorf("V0Stage1 %s blockcut: %w", sz.Name, err)
	}
	cutaway = bld.Translate(cutaway, 0, 0, sz.H+V0Stage1ExplodeGap)
	return bld.Union(block, cutaway), nil
}
