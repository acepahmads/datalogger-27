package formula

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type TokenType int

const (
	TokenNumber TokenType = iota
	TokenIdent
	TokenPlus
	TokenMinus
	TokenMul
	TokenDiv
	TokenMod
	TokenPow
	TokenLParen
	TokenRParen
	TokenComma
	TokenEOF
)

type Token struct {
	Type  TokenType
	Text  string
	Value float64
}

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

func (l *Lexer) NextToken() (Token, error) {
	l.skipWhitespace()
	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF}, nil
	}

	ch := l.input[l.pos]

	if unicode.IsDigit(ch) || ch == '.' {
		start := l.pos
		hasDot := (ch == '.')
		l.pos++
		for l.pos < len(l.input) {
			c := l.input[l.pos]
			if unicode.IsDigit(c) {
				l.pos++
			} else if c == '.' && !hasDot {
				hasDot = true
				l.pos++
			} else {
				break
			}
		}
		str := string(l.input[start:l.pos])
		val, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return Token{}, fmt.Errorf("invalid number '%s'", str)
		}
		return Token{Type: TokenNumber, Value: val, Text: str}, nil
	}

	if unicode.IsLetter(ch) || ch == '_' {
		start := l.pos
		l.pos++
		for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
			l.pos++
		}
		str := string(l.input[start:l.pos])
		return Token{Type: TokenIdent, Text: str}, nil
	}

	l.pos++
	switch ch {
	case '+':
		return Token{Type: TokenPlus, Text: "+"}, nil
	case '-':
		return Token{Type: TokenMinus, Text: "-"}, nil
	case '*':
		if l.pos < len(l.input) && l.input[l.pos] == '*' {
			l.pos++
			return Token{Type: TokenPow, Text: "**"}, nil
		}
		return Token{Type: TokenMul, Text: "*"}, nil
	case '/':
		return Token{Type: TokenDiv, Text: "/"}, nil
	case '%':
		return Token{Type: TokenMod, Text: "%"}, nil
	case '^':
		return Token{Type: TokenPow, Text: "^"}, nil
	case '(':
		return Token{Type: TokenLParen, Text: "("}, nil
	case ')':
		return Token{Type: TokenRParen, Text: ")"}, nil
	case ',':
		return Token{Type: TokenComma, Text: ","}, nil
	default:
		return Token{}, fmt.Errorf("unexpected character '%c'", ch)
	}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

type Parser struct {
	tokens []Token
	pos    int
	vars   map[string]float64
}

func NewParser(tokens []Token, vars map[string]float64) *Parser {
	normalizedVars := make(map[string]float64)
	for k, v := range vars {
		normalizedVars[strings.ToLower(k)] = v
	}
	return &Parser{tokens: tokens, vars: normalizedVars}
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) consume(expected TokenType) (Token, error) {
	curr := p.current()
	if curr.Type != expected {
		return Token{}, fmt.Errorf("expected token type %v, got %v (%s)", expected, curr.Type, curr.Text)
	}
	p.pos++
	return curr, nil
}

func (p *Parser) Parse() (float64, error) {
	val, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	if p.current().Type != TokenEOF {
		return 0, fmt.Errorf("unexpected token '%s' after valid expression", p.current().Text)
	}
	return val, nil
}

func (p *Parser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}

	for {
		curr := p.current()
		if curr.Type == TokenPlus {
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left += right
		} else if curr.Type == TokenMinus {
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left -= right
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parseTerm() (float64, error) {
	left, err := p.parsePower()
	if err != nil {
		return 0, err
	}

	for {
		curr := p.current()
		if curr.Type == TokenMul {
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			left *= right
		} else if curr.Type == TokenDiv {
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		} else if curr.Type == TokenMod {
			p.pos++
			right, err := p.parsePower()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			left = math.Mod(left, right)
		} else {
			break
		}
	}
	return left, nil
}

func (p *Parser) parsePower() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}

	if p.current().Type == TokenPow {
		p.pos++
		right, err := p.parsePower() // Right-associative exponentiation
		if err != nil {
			return 0, err
		}
		return math.Pow(left, right), nil
	}
	return left, nil
}

func (p *Parser) parseFactor() (float64, error) {
	curr := p.current()
	if curr.Type == TokenPlus {
		p.pos++
		return p.parseFactor()
	}
	if curr.Type == TokenMinus {
		p.pos++
		val, err := p.parseFactor()
		if err != nil {
			return 0, err
		}
		return -val, nil
	}
	return p.parsePrimary()
}

func (p *Parser) parsePrimary() (float64, error) {
	curr := p.current()
	if curr.Type == TokenNumber {
		p.pos++
		return curr.Value, nil
	}

	if curr.Type == TokenLParen {
		p.pos++
		val, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		if _, err := p.consume(TokenRParen); err != nil {
			return 0, fmt.Errorf("missing closing parenthesis ')'")
		}
		return val, nil
	}

	if curr.Type == TokenIdent {
		name := curr.Text
		p.pos++

		// Check if function call
		if p.current().Type == TokenLParen {
			p.pos++
			args := []float64{}
			if p.current().Type != TokenRParen {
				for {
					arg, err := p.parseExpression()
					if err != nil {
						return 0, err
					}
					args = append(args, arg)
					if p.current().Type == TokenComma {
						p.pos++
					} else {
						break
					}
				}
			}
			if _, err := p.consume(TokenRParen); err != nil {
				return 0, fmt.Errorf("missing ')' after function arguments in %s", name)
			}
			return evalFunction(strings.ToLower(name), args)
		}

		// Otherwise variable or constant
		lower := strings.ToLower(name)
		if lower == "pi" {
			return math.Pi, nil
		}
		if lower == "e" {
			return math.E, nil
		}
		if val, ok := p.vars[lower]; ok {
			return val, nil
		}
		return 0, fmt.Errorf("undefined variable or function '%s'", name)
	}

	return 0, fmt.Errorf("unexpected token '%s'", curr.Text)
}

func evalFunction(fn string, args []float64) (float64, error) {
	switch fn {
	case "abs":
		if len(args) != 1 {
			return 0, fmt.Errorf("abs() requires 1 argument")
		}
		return math.Abs(args[0]), nil
	case "round":
		if len(args) == 1 {
			return math.Round(args[0]), nil
		} else if len(args) == 2 {
			prec := int(args[1])
			factor := math.Pow10(prec)
			return math.Round(args[0]*factor) / factor, nil
		}
		return 0, fmt.Errorf("round() requires 1 or 2 arguments")
	case "floor":
		if len(args) != 1 {
			return 0, fmt.Errorf("floor() requires 1 argument")
		}
		return math.Floor(args[0]), nil
	case "ceil":
		if len(args) != 1 {
			return 0, fmt.Errorf("ceil() requires 1 argument")
		}
		return math.Ceil(args[0]), nil
	case "sqrt":
		if len(args) != 1 {
			return 0, fmt.Errorf("sqrt() requires 1 argument")
		}
		if args[0] < 0 {
			return 0, fmt.Errorf("sqrt() of negative number")
		}
		return math.Sqrt(args[0]), nil
	case "min":
		if len(args) < 2 {
			return 0, fmt.Errorf("min() requires at least 2 arguments")
		}
		m := args[0]
		for _, v := range args[1:] {
			if v < m {
				m = v
			}
		}
		return m, nil
	case "max":
		if len(args) < 2 {
			return 0, fmt.Errorf("max() requires at least 2 arguments")
		}
		m := args[0]
		for _, v := range args[1:] {
			if v > m {
				m = v
			}
		}
		return m, nil
	case "pow":
		if len(args) != 2 {
			return 0, fmt.Errorf("pow() requires 2 arguments")
		}
		return math.Pow(args[0], args[1]), nil
	case "log", "ln":
		if len(args) != 1 {
			return 0, fmt.Errorf("log() requires 1 argument")
		}
		if args[0] <= 0 {
			return 0, fmt.Errorf("log() argument must be positive")
		}
		return math.Log(args[0]), nil
	case "log10":
		if len(args) != 1 {
			return 0, fmt.Errorf("log10() requires 1 argument")
		}
		if args[0] <= 0 {
			return 0, fmt.Errorf("log10() argument must be positive")
		}
		return math.Log10(args[0]), nil
	case "exp":
		if len(args) != 1 {
			return 0, fmt.Errorf("exp() requires 1 argument")
		}
		return math.Exp(args[0]), nil
	case "sin":
		if len(args) != 1 {
			return 0, fmt.Errorf("sin() requires 1 argument")
		}
		return math.Sin(args[0]), nil
	case "cos":
		if len(args) != 1 {
			return 0, fmt.Errorf("cos() requires 1 argument")
		}
		return math.Cos(args[0]), nil
	case "tan":
		if len(args) != 1 {
			return 0, fmt.Errorf("tan() requires 1 argument")
		}
		return math.Tan(args[0]), nil
	default:
		return 0, fmt.Errorf("unknown function '%s'", fn)
	}
}

// Evaluate evaluates a mathematical formula with given variables.
// Common variables supported: x, val, value, raw.
func Evaluate(expr string, vars map[string]float64) (float64, error) {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return 0, fmt.Errorf("empty expression")
	}

	lexer := NewLexer(expr)
	var tokens []Token
	for {
		tok, err := lexer.NextToken()
		if err != nil {
			return 0, err
		}
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}

	parser := NewParser(tokens, vars)
	return parser.Parse()
}

// Validate checks if a formula expression is syntactically valid with dummy variables.
func Validate(expr string) error {
	expr = strings.TrimSpace(expr)
	if expr == "" {
		return nil // Empty formula is allowed (no formula)
	}
	dummyVars := map[string]float64{
		"x":     1.0,
		"val":   1.0,
		"value": 1.0,
		"raw":   1.0,
	}
	_, err := Evaluate(expr, dummyVars)
	return err
}
