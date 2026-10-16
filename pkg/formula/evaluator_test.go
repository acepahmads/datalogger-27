package formula

import (
	"math"
	"testing"
)

func TestEvaluateFormulas(t *testing.T) {
	tests := []struct {
		name    string
		expr    string
		vars    map[string]float64
		want    float64
		wantErr bool
	}{
		{
			name: "Linear conversion (C to F)",
			expr: "x * 1.8 + 32",
			vars: map[string]float64{"x": 25.0},
			want: 77.0,
		},
		{
			name: "4-20mA current loop mapping",
			expr: "(raw - 4.0) * (100.0 / 16.0)",
			vars: map[string]float64{"raw": 12.0},
			want: 50.0,
		},
		{
			name: "Rounding function with precision",
			expr: "round(x, 2)",
			vars: map[string]float64{"x": 3.14159},
			want: 3.14,
		},
		{
			name: "Square root and power",
			expr: "sqrt(x^2 + y^2)",
			vars: map[string]float64{"x": 3, "y": 4},
			want: 5.0,
		},
		{
			name: "Double asterisk power syntax",
			expr: "x ** 3",
			vars: map[string]float64{"x": 2},
			want: 8.0,
		},
		{
			name: "Min and Max clamping",
			expr: "max(0, min(100, x))",
			vars: map[string]float64{"x": 120},
			want: 100.0,
		},
		{
			name: "Min and Max lower bound clamp",
			expr: "max(0, min(100, x))",
			vars: map[string]float64{"x": -15},
			want: 0.0,
		},
		{
			name: "Constants pi and e",
			expr: "round(pi, 4)",
			vars: nil,
			want: 3.1416,
		},
		{
			name: "Logarithm and exponential",
			expr: "round(exp(log(x)), 2)",
			vars: map[string]float64{"x": 10.0},
			want: 10.0,
		},
		{
			name: "Trigonometric sin and cos",
			expr: "round(sin(0), 1)",
			vars: nil,
			want: 0.0,
		},
		{
			name: "Unary minus and precedence",
			expr: "-x + 10 * 2",
			vars: map[string]float64{"x": 5},
			want: 15.0,
		},
		{
			name: "Modulo operator",
			expr: "17 % 5",
			vars: nil,
			want: 2.0,
		},
		{
			name:    "Division by zero returns error",
			expr:    "x / 0",
			vars:    map[string]float64{"x": 10},
			wantErr: true,
		},
		{
			name:    "Undefined variable returns error",
			expr:    "x + unknown_var",
			vars:    map[string]float64{"x": 10},
			wantErr: true,
		},
		{
			name:    "Missing closing parenthesis returns error",
			expr:    "(x + 5",
			vars:    map[string]float64{"x": 10},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Evaluate(tt.expr, tt.vars)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Evaluate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				if math.Abs(got-tt.want) > 0.0001 {
					t.Errorf("Evaluate() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestValidate(t *testing.T) {
	if err := Validate("x * 1.8 + 32"); err != nil {
		t.Errorf("expected valid formula, got %v", err)
	}
	if err := Validate(""); err != nil {
		t.Errorf("expected empty formula to be valid (no-op), got %v", err)
	}
	if err := Validate("x +* 2"); err == nil {
		t.Errorf("expected invalid syntax error, got nil")
	}
	if err := Validate("value + random(0.01, 0.09)"); err != nil {
		t.Errorf("expected valid formula with random and value, got %v", err)
	}
	if err := Validate("value + (random(0.01, 0.09))"); err != nil {
		t.Errorf("expected valid formula with nested parens and random, got %v", err)
	}
	if err := Validate("random(10, 2)"); err == nil {
		t.Errorf("expected error when min > max, got nil")
	}
	if err := Validate("random(1)"); err == nil {
		t.Errorf("expected error when random has only 1 arg, got nil")
	}
}

func TestRandomFunction(t *testing.T) {
	t.Run("Range and bounds [0.01, 0.09]", func(t *testing.T) {
		const minVal = 0.01
		const maxVal = 0.09
		uniqueValues := make(map[float64]bool)

		for i := 0; i < 200; i++ {
			val, err := Evaluate("random(0.01, 0.09)", nil)
			if err != nil {
				t.Fatalf("unexpected error on iteration %d: %v", i, err)
			}
			if val < minVal || val > maxVal {
				t.Fatalf("iteration %d: value %f outside range [%f, %f]", i, val, minVal, maxVal)
			}
			uniqueValues[val] = true
		}

		if len(uniqueValues) < 50 {
			t.Errorf("expected diverse random distribution, got only %d distinct values across 200 iterations", len(uniqueValues))
		}
	})

	t.Run("Negative range [-10.5, -2.1]", func(t *testing.T) {
		const minVal = -10.5
		const maxVal = -2.1
		for i := 0; i < 100; i++ {
			val, err := Evaluate("random(-10.5, -2.1)", nil)
			if err != nil {
				t.Fatalf("iteration %d error: %v", i, err)
			}
			if val < minVal || val > maxVal {
				t.Fatalf("iteration %d: value %f outside range [%f, %f]", i, val, minVal, maxVal)
			}
		}
	})

	t.Run("Equal bounds random(4.5, 4.5)", func(t *testing.T) {
		val, err := Evaluate("random(4.5, 4.5)", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if math.Abs(val-4.5) > 1e-9 {
			t.Errorf("got %v, want 4.5", val)
		}
	})

	t.Run("Expression arguments random(2*2, 3+5)", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			val, err := Evaluate("random(2 * 2, 3 + 5)", nil)
			if err != nil {
				t.Fatalf("iteration %d error: %v", i, err)
			}
			if val < 4.0 || val > 8.0 {
				t.Fatalf("iteration %d: value %f outside range [4, 8]", i, val)
			}
		}
	})
}

func TestRandomCombinations(t *testing.T) {
	t.Run("value + random(0.01, 0.09) with value=2.0", func(t *testing.T) {
		vars := map[string]float64{"value": 2.0}
		for i := 0; i < 100; i++ {
			val, err := Evaluate("value + random(0.01, 0.09)", vars)
			if err != nil {
				t.Fatalf("unexpected error on iteration %d: %v", i, err)
			}
			if val < 2.01 || val > 2.09 {
				t.Fatalf("iteration %d: got %f, want within [2.01, 2.09]", i, val)
			}
		}
	})

	t.Run("value + (random(0.01, 0.09)) with nested parentheses", func(t *testing.T) {
		vars := map[string]float64{"value": 5.0}
		for i := 0; i < 50; i++ {
			val, err := Evaluate("value + (random(0.01, 0.09))", vars)
			if err != nil {
				t.Fatalf("unexpected error on iteration %d: %v", i, err)
			}
			if val < 5.01 || val > 5.09 {
				t.Fatalf("iteration %d: got %f, want within [5.01, 5.09]", i, val)
			}
		}
	})

	t.Run("x and raw variables support", func(t *testing.T) {
		vars := map[string]float64{
			"x":   10.0,
			"raw": 100.0,
		}
		valX, err := Evaluate("x + random(0.5, 1.5)", vars)
		if err != nil || valX < 10.5 || valX > 11.5 {
			t.Errorf("x evaluation failed: val=%v, err=%v", valX, err)
		}

		valRaw, err := Evaluate("raw + random(1, 2)", vars)
		if err != nil || valRaw < 101.0 || valRaw > 102.0 {
			t.Errorf("raw evaluation failed: val=%v, err=%v", valRaw, err)
		}
	})

	t.Run("EvalWithXRaw helper support", func(t *testing.T) {
		for i := 0; i < 50; i++ {
			val, err := EvalWithXRaw("value + random(0.01, 0.09)", 25.0, 50.0)
			if err != nil {
				t.Fatalf("EvalWithXRaw error: %v", err)
			}
			if val < 25.01 || val > 25.09 {
				t.Fatalf("EvalWithXRaw value %v outside [25.01, 25.09]", val)
			}
		}
	})

	t.Run("Composed with round and min/max", func(t *testing.T) {
		vars := map[string]float64{"value": 1.234}
		val, err := Evaluate("round(value + random(0.01, 0.02), 2)", vars)
		if err != nil {
			t.Fatalf("composed round error: %v", err)
		}
		// 1.234 + [0.01, 0.02] = [1.244, 1.254], rounded to 2 decimals gives 1.24 or 1.25
		if val != 1.24 && val != 1.25 {
			t.Errorf("expected 1.24 or 1.25, got %v", val)
		}
	})
}

func TestRandomInvalidArguments(t *testing.T) {
	invalidCases := []struct {
		name string
		expr string
	}{
		{"min greater than max", "random(10, 2)"},
		{"negative min greater than max", "random(-2, -5)"},
		{"no arguments", "random()"},
		{"one argument", "random(1)"},
		{"three arguments", "random(1, 2, 3)"},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := Evaluate(tc.expr, nil)
			if err == nil {
				t.Fatalf("expected error for '%s', got nil", tc.expr)
			}
		})
	}
}
