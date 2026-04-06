// TOKENS --> AST

// This is the Pratt Parser approach to parsing expressions
// The key insight: each token type can have parsing functions associated with it
// When we encounter that token, we use those functions to parse the expression
package parser

import (
	"fmt"
	"monkey/ast"
	"monkey/lexer"
	"monkey/token"
	"strconv"
)

// Numbering consts to determine order and relation to each other
// for precedence - BEDMAS
const (
	_           int = iota // give the following constants incrementing numbers as values
	LOWEST                 //1
	EQUALS                 //2 ==
	LESSGREATER            //3 > or <
	SUM                    //4 +
	PRODUCT                //5 *
	PREFIX                 //6 -x or !x
	CALL                   //7 fn(x)
)

// Precedence table
// associates token types with their precedence.
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
}

// Implementing the Pratt Parser
// Whenever this token type is encountered, the parsing functions are called to parse the appropriate expression and return an AST node that represents it
// Each token type can have up to two parsing functions associated
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Parser takes a stream of tokens from the Lexer and builds an AST
type Parser struct {
	l      *lexer.Lexer //pointer to an instance of the lexer
	errors []string

	curToken  token.Token
	peekToken token.Token

	// Maps that store which parsing function to call for each token type
	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// New creates a new parser and sets it up to parse
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	// Create maps to store parsing functions
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)

	//register prefix parse functions
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)

	//register one infix parse function for all of  infix operators
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)

	//register boolean
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.FALSE, p.parseBoolean)

	//register grouping by ()
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	// register if
	p.registerPrefix(token.IF, p.parseIfExpression)

	// register function
	p.registerPrefix(token.FUNCTION, p.parseFunctionLiteral)

	//  register an infixParseFn for token.LPAREN. This way we parse the
	// expression that is the function (either an identifier, or a function literal)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	// Read two tokens, so curToken and peekToken are both set
	p.nextToken()
	p.nextToken()

	return p
}

// Creates an Identifier AST node with the token and its value
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

// This handles operators that come BEFORE a value
func (p *Parser) parsePrefixExpression() ast.Expression {
	defer untrace(trace("parsePrefixExpression"))

	// Create a PrefixExpression node to hold the operator and the thing it operates on
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()

	// Recursively parse the right side with high precedence (PREFIX)
	// -5*2 --? (-5)*2
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseInfixExpression parses an infix expression.
// takes an argument, an ast.Expression called left. It uses this argument to construct an
// *ast.InfixExpression node, with left being in the Left field. Then it assigns the precedence of
// the current token (which is the operator of the infix expression) to the local variable precedence,
// before advancing the tokens by calling nextToken and filling the Right field of the node with
// another call to parseExpression - this time passing in the precedence of the operator token.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	defer untrace(trace("parseInfixExpression"))

	// Create an InfixExpression node to hold both sides and the operator
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	// Get the precedence of this operator
	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence) //pass in the operator's precedence so operations with equal or lower precedence stop here

	return expression
}

// Parse boolean
func (p *Parser) parseBoolean() ast.Expression {
	return &ast.Boolean{Token: p.curToken, Value: p.curTokenIs(token.TRUE)}
}

// Parse grouped expression ()
func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	// Parse the expression inside the parentheses with LOWEST precedence
	exp := p.parseExpression(LOWEST)

	// Make sure there's a closing parenthesis ")"
	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

// parse if expression
// (<condition>) { <consequence> } else { <alternative> }
func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		expression.Alternative = p.parseBlockStatement()
	}
	return expression
}

// This handles a block of statements surrounded by { }
func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// This handles function definitions like: fn(x, y) { return x + y; }
func (p *Parser) parseFunctionLiteral() ast.Expression {
	lit := &ast.FunctionLiteral{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	lit.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	lit.Body = p.parseBlockStatement()

	return lit
}

// parseFunctionParameters constructs the slice of parameters by
// repeatedly building identifiers from the comma separated list.
// It also makes an early exit if the list is empty and it carefully
// handles lists of varying sizes.

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return identifiers
	}

	p.nextToken()

	ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
	identifiers = append(identifiers, ident)

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

// CHECKING FOR PRECEDENCE
// returns the precedence associated with the token type ofp.peekToken.
// If it doesn’t find a precedence for p.peekToken it defaults to LOWEST
func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}

	return LOWEST
}

// returns precedence for current token
func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}

	return LOWEST
}

// Errors
func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead",
		t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

// This is the main entry point - it parses an entire program
// It keeps parsing statements until it hits the end of file (EOF)
func (p *Parser) ParseProgram() *ast.Program {
	// Create an empty Program node
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

// This decides what KIND of statement
// and calls the appropriate parsing function
func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.LET:
		return p.parseLetStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// PRATT PARSER
// It uses precedence to decide how to combine operations
//
// How it works:
// 1. Find and call the prefix parsing function for the current token
// 2. Use that result as the left side
// 3. While the next token has higher precedence:
//   - Call the infix parsing function for that token
//   - Use the result as the new left side
//
// 4. When we hit a lower-precedence token, stop and return
//
// Example: parsing "5 + 3 * 2" with precedence 0
// - Parse 5 (prefix) -> leftExp = 5
// - See +, check precedence 4 > 0, so continue
// - Parse + (infix) with left=5 -> leftExp = (5 + ...)
// - Parse 3 * 2 with precedence 4
//   - Parse 3 (prefix) -> leftExp = 3
//   - See *, check precedence 5 > 4, so continue
//   - Parse * (infix) -> leftExp = (3 * 2)
//   - No more tokens with precedence > 4, stop
//
// - Combine: (5 + (3 * 2)) = (5 + 6) = 11
func (p *Parser) parseExpression(precedence int) ast.Expression {
	defer untrace(trace("parseExpression"))

	// Find the prefix parsing function for the current token
	prefix := p.prefixParseFns[p.curToken.Type]

	// If there's no prefix function, we can't parse this as an expression
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}

	// Call the prefix function to parse the left side
	leftExp := prefix()

	//In the loop’s body the method tries to find infixParseFns for the next token.
	//If it finds such a function, it calls it, passing in the expression returned by a prefixParseFn as an argument.
	//And it does all this again and again until it encounters a token that has a higher precedence.
	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		// Try to find an infix parsing function for the next token
		infix := p.infixParseFns[p.peekToken.Type]

		// If there's no infix function, we can't continue
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		// Call the infix function, passing the left side we built so far
		// It will build the right side and combine them
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// <identifier> = <expression>
func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	defer untrace(trace("parseExpressionStatement"))

	stmt := &ast.ExpressionStatement{Token: p.curToken}

	stmt.Expression = p.parseExpression(LOWEST)

	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	defer untrace(trace("parseIntegerLiteral"))

	lit := &ast.IntegerLiteral{Token: p.curToken}

	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

// Parsing function calls
// parseCallExpression receives the already parsed function as argument and uses it to construct
// an *ast.CallExpression node. To parse the argument list we call parseCallArguments, which
// looks strikingly similar to parseFunctionParameters, except that it’s more generic and returns
// a slice of ast.Expression and not *ast.Identifier.

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}
	if p.peekTokenIs(token.RPAREN) {
		p.nextToken()
		return args
	}
	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))
	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}
	if !p.expectPeek(token.RPAREN) {
		return nil
	}
	return args
}

// Methods that add entries to the maps
func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
