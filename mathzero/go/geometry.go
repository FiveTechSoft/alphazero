// Geometry bridge: polynomial translations only. No geometric inference,
// floating point, division, solver change, or automatic training is performed.
package main

import _ "embed"

// Point holds Cartesian coordinates in the existing integer-coefficient ring.
// Callers must supply non-nil expressions and keep coefficients within int range.
type Point struct {
	X, Y *Expr
}

func Pt(xName, yName string) Point { return Point{X: V(xName), Y: V(yName)} }

// DistSq avoids square roots: |AB|^2 = dx^2 + dy^2.
func DistSq(a, b Point) *Expr {
	dx := EAdd(b.X, ENeg(a.X))
	dy := EAdd(b.Y, ENeg(a.Y))
	return EAdd(EMul(dx, dx), EMul(dy, dy))
}

func Dot(v1x, v1y, v2x, v2y *Expr) *Expr {
	return EAdd(EMul(v1x, v2x), EMul(v1y, v2y))
}

// IsPerpendicular returns a residual polynomial, NOT a boolean. Its vanishing
// expresses orthogonality; zero-length segments also satisfy it. It does not
// discharge assumptions, prove nondegeneracy, or invoke the exact judge.
func IsPerpendicular(a, b, c, d Point) *Expr {
	abx, aby := EAdd(b.X, ENeg(a.X)), EAdd(b.Y, ENeg(a.Y))
	cdx, cdy := EAdd(d.X, ENeg(c.X)), EAdd(d.Y, ENeg(c.Y))
	return Dot(abx, aby, cdx, cdy)
}

// IsCollinear returns the determinant residual with the sign convention of
// the proposed bridge. No division by a slope: vertical lines are supported.
// Repeated points also give zero, without establishing a distinct line.
func IsCollinear(a, b, c Point) *Expr {
	abx, aby := EAdd(b.X, ENeg(a.X)), EAdd(b.Y, ENeg(a.Y))
	bcx, bcy := EAdd(c.X, ENeg(b.X)), EAdd(c.Y, ENeg(b.Y))
	return EAdd(EMul(aby, bcx), ENeg(EMul(bcy, abx)))
}

// TwiceMidpoint is the NUMERATOR 2M = A+B, not the actual midpoint M.
// Distances from this scaled point are scaled too; never silently treat it as M.
func TwiceMidpoint(a, b Point) Point {
	return Point{X: EAdd(a.X, b.X), Y: EAdd(a.Y, b.Y)}
}

// MidpointResidual gives the two equations 2Mx-Ax-Bx = 0, 2My-Ay-By = 0.
// These are constraints only, not a general implication-proving mechanism.
func MidpointResidual(m, a, b Point) (*Expr, *Expr) {
	x := EAdd(EMul(C(2), m.X), ENeg(EAdd(a.X, b.X)))
	y := EAdd(EMul(C(2), m.Y), ENeg(EAdd(a.Y, b.Y)))
	return x, y
}

//go:embed geometry_levels.json
var embeddedGeometryLevels []byte

// LoadGeometryProblems loads the separate opt-in identity bank. It does not
// alter the historical curriculum, model, library, or CLI defaults. maxSteps
// is a path budget, not a claim that a proof has been found. Disk overrides
// follow the existing data-file convention; otherwise the bank is embedded.
func LoadGeometryProblems(level string, maxSteps int) []MState {
	data := readDataFile("geometry_levels.json")
	if data == nil {
		data = embeddedGeometryLevels
	}
	out := LoadProblemsBytes(data, level)
	if maxSteps <= 0 {
		maxSteps = 64
	}
	for i := range out {
		out[i].MaxSteps = maxSteps
	}
	return out
}
