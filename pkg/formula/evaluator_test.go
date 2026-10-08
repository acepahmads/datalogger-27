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
}
