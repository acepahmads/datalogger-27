package calibration

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

// Supported calibration regression models
const (
	ModelNone      = "NONE"
	ModelOffset1Pt = "OFFSET_1PT"
	ModelSpan1Pt   = "SPAN_1PT"
	ModelLinear    = "LINEAR"
	ModelQuadratic = "QUADRATIC"

	DefaultMinR2 = 0.995
	MinPoints    = 1
	MaxPoints    = 10
)

// CalibrationPoint represents a single standard calibration point (1..10)
type CalibrationPoint struct {
	Index    int      `json:"index"`
	Active   bool     `json:"active"`
	Label    string   `json:"label,omitempty"`
	RawX     float64  `json:"raw_x"`
	RefY     float64  `json:"ref_y"`
	PredY    float64  `json:"pred_y"`
	ErrorAbs float64  `json:"error_abs"`
	ErrorPct *float64 `json:"error_pct,omitempty"`
}

// RegressionResult holds the computed coefficients, goodness-of-fit metrics, and evaluated points
type RegressionResult struct {
	Model           string             `json:"model"`
	C2              float64            `json:"c2"`
	C1              float64            `json:"c1"`
	C0              float64            `json:"c0"`
	Equation        string             `json:"equation"`
	ActiveCount     int                `json:"active_count"`
	TotalCount      int                `json:"total_count"`
	R2              *float64           `json:"r2"`
	MinR2           float64            `json:"min_r2"`
	PassedR2        bool               `json:"passed_r2"`
	RMSE            float64            `json:"rmse"`
	Points          []CalibrationPoint `json:"points"`
	AdvisoryMessage string             `json:"advisory_message,omitempty"`
}

// NormalizeModel canonicalizes calibration model strings
func NormalizeModel(m string) string {
	switch strings.ToUpper(strings.TrimSpace(m)) {
	case ModelOffset1Pt, "OFFSET", "1PT_OFFSET":
		return ModelOffset1Pt
	case ModelSpan1Pt, "SPAN", "1PT_SPAN", "FACTOR":
		return ModelSpan1Pt
	case ModelQuadratic, "POLY2", "POLYNOMIAL", "ORDER2":
		return ModelQuadratic
	case ModelNone:
		return ModelNone
	default:
		return ModelLinear
	}
}

// NormalizeMinR2 ensures per-parameter minimum R2 threshold is within a sensible range (default 0.995)
func NormalizeMinR2(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) || v <= 0 || v > 1.0 {
		return DefaultMinR2
	}
	if v < 0.5 {
		return 0.5
	}
	return v
}

// EvaluateCurve applies calibration curve y = c2*x^2 + c1*x + c0 when calibration is enabled
func EvaluateCurve(enabled bool, modelType string, c2, c1, c0, x float64) float64 {
	if !enabled || math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	switch NormalizeModel(modelType) {
	case ModelNone:
		return x
	case ModelOffset1Pt:
		return x + c0
	case ModelSpan1Pt:
		return c1 * x
	case ModelQuadratic:
		return c2*x*x + c1*x + c0
	default: // LINEAR
		return c1*x + c0
	}
}

// FormatEquation builds a clean human-readable equation string for display and audit trail
func FormatEquation(enabled bool, modelType string, c2, c1, c0 float64) string {
	if !enabled {
		return "y = x (Default)"
	}
	switch NormalizeModel(modelType) {
	case ModelNone:
		return "y = x (Default)"
	case ModelOffset1Pt:
		if math.Abs(c0) < 1e-9 {
			return "y = x + 0.0000"
		}
		if c0 >= 0 {
			return fmt.Sprintf("y = x + %.4f", c0)
		}
		return fmt.Sprintf("y = x - %.4f", math.Abs(c0))
	case ModelSpan1Pt:
		return fmt.Sprintf("y = %.4fx", c1)
	case ModelQuadratic:
		bSign := "+"
		if c1 < 0 {
			bSign = "-"
		}
		cSign := "+"
		if c0 < 0 {
			cSign = "-"
		}
		return fmt.Sprintf("y = %.5fx² %s %.4fx %s %.4f", c2, bSign, math.Abs(c1), cSign, math.Abs(c0))
	default: // LINEAR
		if c0 >= 0 {
			return fmt.Sprintf("y = %.4fx + %.4f", c1, c0)
		}
		return fmt.Sprintf("y = %.4fx - %.4f", c1, math.Abs(c0))
	}
}

// CalculateRegression computes 1-to-10 point sensor calibration regression and per-point diagnostics
func CalculateRegression(modelType string, inputPoints []CalibrationPoint, minR2 float64) (*RegressionResult, error) {
	total := len(inputPoints)
	if total < MinPoints {
		return nil, errors.New("minimal 1 titik kalibrasi diperlukan")
	}
	if total > MaxPoints {
		return nil, fmt.Errorf("maksimal %d titik kalibrasi diperbolehkan", MaxPoints)
	}

	normModel := NormalizeModel(modelType)
	if normModel == ModelNone {
		normModel = ModelLinear
	}
	effMinR2 := NormalizeMinR2(minR2)

	points := make([]CalibrationPoint, total)
	activeX := make([]float64, 0, total)
	activeY := make([]float64, 0, total)

	for i, pt := range inputPoints {
		if math.IsNaN(pt.RawX) || math.IsInf(pt.RawX, 0) || math.IsNaN(pt.RefY) || math.IsInf(pt.RefY, 0) {
			return nil, fmt.Errorf("nilai titik #%d tidak valid (NaN/Inf)", i+1)
		}
		cp := pt
		if cp.Index <= 0 {
			cp.Index = i + 1
		}
		points[i] = cp
		if cp.Active {
			activeX = append(activeX, cp.RawX)
			activeY = append(activeY, cp.RefY)
		}
	}

	n := len(activeX)
	if n == 0 {
		return nil, errors.New("minimal 1 titik aktif harus dicentang")
	}

	var c2, c1, c0 float64
	var advisory string

	switch normModel {
	case ModelOffset1Pt:
		// y = x + c0 -> optimal mean offset across active point(s)
		sumDiff := 0.0
		for i := 0; i < n; i++ {
			sumDiff += activeY[i] - activeX[i]
		}
		c2 = 0
		c1 = 1.0
		c0 = sumDiff / float64(n)
		if n > 1 {
			advisory = "Mode 1-Titik Offset menghitung rata-rata selisih (y - x) dari titik aktif."
		}

	case ModelSpan1Pt:
		// y = c1 * x -> least-squares slope through origin
		sumXX := 0.0
		sumXY := 0.0
		for i := 0; i < n; i++ {
			sumXX += activeX[i] * activeX[i]
			sumXY += activeX[i] * activeY[i]
		}
		if sumXX < 1e-12 {
			return nil, errors.New("bacaan sensor (x) tidak boleh bernilai 0 pada kalibrasi 1-Titik Faktor/Span")
		}
		c2 = 0
		c1 = sumXY / sumXX
		c0 = 0

	case ModelLinear:
		if n < 2 {
			return nil, errors.New("regresi linier membutuhkan minimal 2 titik aktif (gunakan model 1-Titik untuk 1 titik standar)")
		}
		meanX, meanY := 0.0, 0.0
		for i := 0; i < n; i++ {
			meanX += activeX[i]
			meanY += activeY[i]
		}
		meanX /= float64(n)
		meanY /= float64(n)

		sxx, sxy := 0.0, 0.0
		for i := 0; i < n; i++ {
			dx := activeX[i] - meanX
			dy := activeY[i] - meanY
			sxx += dx * dx
			sxy += dx * dy
		}
		if sxx < 1e-12 {
			return nil, errors.New("seluruh nilai bacaan sensor (x) identik; minimal 2 nilai x berbeda diperlukan untuk regresi linier")
		}
		c2 = 0
		c1 = sxy / sxx
		c0 = meanY - c1*meanX

	case ModelQuadratic:
		if n < 3 {
			return nil, errors.New("regresi polinomial kuadratik (orde-2) membutuhkan minimal 3 titik aktif")
		}
		var err error
		c2, c1, c0, err = solveQuadraticOLS(activeX, activeY)
		if err != nil {
			return nil, err
		}
		if n == 3 {
			advisory = "Dengan tepat 3 titik, kurva kuadratik melewati seluruh titik (R² = 1). Disarankan ≥ 4 titik untuk menguji residu."
		}
	}

	// Clean -0.0000 artifacts
	if math.Abs(c2) < 1e-15 {
		c2 = 0
	}
	if math.Abs(c1) < 1e-15 {
		c1 = 0
	}
	if math.Abs(c0) < 1e-15 {
		c0 = 0
	}

	// Evaluate all points (active and excluded outliers) and compute RMSE & R2 on active points
	sse := 0.0
	meanY := 0.0
	for i := 0; i < n; i++ {
		meanY += activeY[i]
	}
	meanY /= float64(n)

	syy := 0.0
	for i := range points {
		pred := EvaluateCurve(true, normModel, c2, c1, c0, points[i].RawX)
		errAbs := pred - points[i].RefY
		points[i].PredY = pred
		points[i].ErrorAbs = errAbs
		if math.Abs(points[i].RefY) >= 1e-6 {
			pct := (errAbs / math.Abs(points[i].RefY)) * 100.0
			points[i].ErrorPct = &pct
		} else {
			points[i].ErrorPct = nil
		}

		if points[i].Active {
			diff := points[i].RefY - pred
			sse += diff * diff
			dy := points[i].RefY - meanY
			syy += dy * dy
		}
	}

	rmse := math.Sqrt(sse / float64(n))
	var r2Ptr *float64
	passedR2 := true

	if n >= 2 && normModel != ModelOffset1Pt && normModel != ModelSpan1Pt {
		r2Val := 1.0
		if syy > 1e-12 {
			r2Val = 1.0 - (sse / syy)
			if r2Val > 1.0 {
				r2Val = 1.0
			}
		} else if sse > 1e-9 {
			r2Val = 0.0
		}
		r2Ptr = &r2Val
		passedR2 = r2Val >= effMinR2
	} else if n >= 2 && syy > 1e-12 {
		// Even for 1-point model applied across multiple checked points, compute R2 informatively
		r2Val := 1.0 - (sse / syy)
		if r2Val > 1.0 {
			r2Val = 1.0
		}
		r2Ptr = &r2Val
		passedR2 = r2Val >= effMinR2
	}

	return &RegressionResult{
		Model:           normModel,
		C2:              c2,
		C1:              c1,
		C0:              c0,
		Equation:        FormatEquation(true, normModel, c2, c1, c0),
		ActiveCount:     n,
		TotalCount:      total,
		R2:              r2Ptr,
		MinR2:           effMinR2,
		PassedR2:        passedR2,
		RMSE:            rmse,
		Points:          points,
		AdvisoryMessage: advisory,
	}, nil
}

// solveQuadraticOLS solves y = c2*x^2 + c1*x + c0 using standardized variable u = (x - mu)/scale
// for high numerical stability across wide sensor spans (e.g. 0..10,000 uS/cm or mg/L).
func solveQuadraticOLS(xs, ys []float64) (c2, c1, c0 float64, err error) {
	n := len(xs)
	mu := 0.0
	for _, x := range xs {
		mu += x
	}
	mu /= float64(n)

	maxAbs := 0.0
	for _, x := range xs {
		d := math.Abs(x - mu)
		if d > maxAbs {
			maxAbs = d
		}
	}
	if maxAbs < 1e-12 {
		return 0, 0, 0, errors.New("nilai bacaan sensor (x) kurang bervariasi untuk regresi kuadratik")
	}
	scale := maxAbs

	// Normal equations in u = (x - mu) / scale: y = a2*u^2 + a1*u + a0
	var s0, s1, s2, s3, s4 float64
	var t0, t1, t2 float64
	for i := 0; i < n; i++ {
		u := (xs[i] - mu) / scale
		u2 := u * u
		y := ys[i]
		s0 += 1.0
		s1 += u
		s2 += u2
		s3 += u2 * u
		s4 += u2 * u2

		t0 += y
		t1 += u * y
		t2 += u2 * y
	}

	mat := [3][4]float64{
		{s0, s1, s2, t0},
		{s1, s2, s3, t1},
		{s2, s3, s4, t2},
	}

	// Gaussian elimination with partial pivoting
	for col := 0; col < 3; col++ {
		pivot := col
		maxPivot := math.Abs(mat[col][col])
		for row := col + 1; row < 3; row++ {
			if v := math.Abs(mat[row][col]); v > maxPivot {
				maxPivot = v
				pivot = row
			}
		}
		if maxPivot < 1e-11 {
			return 0, 0, 0, errors.New("minimal 3 nilai bacaan sensor (x) yang berbeda diperlukan untuk regresi kuadratik")
		}
		if pivot != col {
			mat[col], mat[pivot] = mat[pivot], mat[col]
		}
		for row := col + 1; row < 3; row++ {
			factor := mat[row][col] / mat[col][col]
			for k := col; k < 4; k++ {
				mat[row][k] -= factor * mat[col][k]
			}
		}
	}

	var sol [3]float64 // a0, a1, a2
	for row := 2; row >= 0; row-- {
		sum := mat[row][3]
		for k := row + 1; k < 3; k++ {
			sum -= mat[row][k] * sol[k]
		}
		sol[row] = sum / mat[row][row]
	}

	a0, a1, a2 := sol[0], sol[1], sol[2]

	// Transform back from u = (x - mu)/scale to x:
	// y = a2*((x - mu)/scale)^2 + a1*((x - mu)/scale) + a0
	invS := 1.0 / scale
	invS2 := invS * invS

	c2 = a2 * invS2
	c1 = a1*invS - 2.0*a2*mu*invS2
	c0 = a0 - a1*mu*invS + a2*mu*mu*invS2
	return c2, c1, c0, nil
}

