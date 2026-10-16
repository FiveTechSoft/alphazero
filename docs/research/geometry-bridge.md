# Geometry bridge, source-only v1

Status: prepared, not compiled or executed. No solver success, timing, training
improvement, regression PASS, or equivalence validation is claimed. Validation
must be coordinated with the other variants before any computation is launched.
The branch starts from the preserved historical updater H, not C or the search
variants. Existing source, model, datasets, library and numerical fixtures stay
unchanged. This is a separate opt-in bank, not a change to the default curriculum.

## What the bridge does

`mathzero/go/geometry.go` translates Point/Pt, DistSq, Dot, IsPerpendicular and
IsCollinear into the existing Expr constructors. The two Is* functions return
polynomial residuals, not booleans or proved geometric facts. Vanishing encodes
the stated relation, including degeneracies. No division, square root, slope,
floating point, new Expr opcode, or change to Canon/canonEq/MathRing is involved.
Caller expressions must be non-nil. Variable names should be simple identifiers
(such as a..f), not monomial delimiters. Canon uses machine int coefficients,
not arbitrary precision: overflow remains an existing limitation. This bridge
must not be described as an unrestricted exact geometry prover.

`TwiceMidpoint(A,B)` returns A+B, the coordinates of 2M, never M.
`MidpointResidual(M,A,B)` returns 2Mx-Ax-Bx and 2My-Ay-By. These encode the
midpoint constraints over the integer-coefficient ring without dividing by 2.
Other fractional constructions must likewise clear denominators consistently
in all coordinates and metric factors. There is no generic rational-coordinate
or division API. Squared distance from doubled coordinates scales by 4.

## Separate identity bank

`geometry_levels.json` uses the existing {level:[{a:Expr,b:Expr}]} format, with
id/meaning metadata ignored by the existing loader. Four one-problem levels:

1. Axis-aligned Pythagoras: O=(0,0), B=(a,0), C=(a,b). The right angle is built
   into the coordinates. OC^2=a^2+b^2 is an unconditional identity, not a proof
   of arbitrary Pythagoras from an independent perpendicularity hypothesis.
2. Squared-distance symmetry for arbitrary planar A and B.
3. Parallelogram metric identity for arbitrary planar vectors u,v.
4. Doubled-midpoint metric identity:
   |A+B-2P|^2 = 2|A-P|^2 + 2|B-P|^2 - |A-B|^2.
   The left side is 4|M-P|^2 when 2M=A+B.

`LoadGeometryProblems("1",64)` (or levels "2".."4") returns ordinary MState
values with a caller-supplied path budget (default64). It is callable by the
validation harness; no CLI mode or training loop is installed or run. Disk files
in the current/parent directory follow the existing override convention; an
embedded copy is available otherwise. Canonical equality is necessary for an
identity; solver success additionally requires a path to the target structure.
The larger identities may exceed a small search budget. No difficulty ranking
or guaranteed 64-step proof is asserted by these level labels.

## Library integration and limits

The returned expressions use the ordinary GuidedSolve/SolvePath/MathActions
interfaces. When the existing library is active, its normal exact-key rewrite
mechanism is available without a geometry-specific bypass. Nothing is inserted
or saved automatically by loading this bank. A caller that has found a verified
path may submit its endpoints to Library.Insert and use the existing Save path.
Insert still enforces canonical equality, saving>=2, distinct structural keys,
duplicate checks and the existing MaxSize=12 limit. Many full metric expressions
are too large to insert unchanged; they must be rejected rather than silently
raising that limit. Keys match concrete expressions; no new schematic-variable
matching or universal geometry lemma generalization is implemented. A proved
identity is not guaranteed to become a reusable one-step rule in another shape.

Canon equality tests A=B for all assignments. It does NOT test whether hypotheses
h1=0,...,hn=0 imply g=0. Adding hypotheses together loses information and is not
an accepted reduction. Ideal membership, certificates, denominator saturation,
nondegeneracy conditions and constructions such as concurrency of altitudes
remain out of scope. No Wu or Groebner algorithm is implemented. Nonzero segment
checks and geometric existence assumptions are not proved by the residuals.

The full prepared validation suite is retained privately, not published. It
covers bridge formulas, the four identity fixtures, negative residuals,
degenerate cases, midpoint scaling, embedded loading, unchanged core hashes and
ordinary library insertion/rewrite. None has been run. Numerical H/C tests,
fixtures and tolerances remain unchanged. No remote computation was launched.
