package parser

import (
	"dao/ast"
	"dao/lexer"
	"dao/token"
	"fmt"
	"strconv"
)

type Parser struct {
	l      *lexer.Lexer
	errors []string

	curTok  token.Token
	nextTok token.Token

	prefixFNs map[token.TokenType]prefixFN
	infixFNs  map[token.TokenType]infixFN
}

type (
	prefixFN func() ast.Expression
	infixFN  func(ast.Expression) ast.Expression
)

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l, errors: []string{}}
	p.prefixFNs = make(map[token.TokenType]prefixFN)
	p.infixFNs = make(map[token.TokenType]infixFN)

	p.registerPrefix(token.ID, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.TRUE, p.parseBoolean)
	p.registerPrefix(token.NIL, p.parseNil)
	p.registerPrefix(token.FALSE, p.parseBoolean)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.IF, p.parseIfExpression)
	p.registerPrefix(token.FUNC, p.parseFunctionLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)

	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)

	p.Next()
	p.Next()

	return p
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expect next token to be %s, got %s instead", t, p.nextTok.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) Next() {
	p.curTok = p.nextTok
	p.nextTok = p.l.Next()
}

func (p *Parser) Parse() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}

		p.Next()
	}
	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curTok.Type {
	case token.VAR:
		return p.parseVarStatement()
	case token.ID:
		if p.nextTok.Type == token.ASSIGN {
			return p.parseAssignStatement()
		}
		return p.parseExpressionStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.FOR:
		return p.parseForStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseVarStatement() *ast.VarStatement {
	stmt := &ast.VarStatement{Token: p.curTok}

	if !p.expectNext(token.ID) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if p.nextTokenIs(token.ID) {
		p.Next()
		stmt.Name.Type = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	}

	if !p.expectNext(token.ASSIGN) {
		return nil
	}

	p.Next()
	stmt.Value = p.parseExpression(LOWEST)

	if p.nextTokenIs(token.SEMICOLON) {
		p.Next()
	}

	return stmt
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curTok}
	stmt.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}

	if !p.expectNext(token.ASSIGN) {
		return nil
	}

	p.Next()

	stmt.Value = p.parseExpression(LOWEST)

	if p.nextTokenIs(token.SEMICOLON) {
		p.Next()
	}

	return stmt
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Token: p.curTok}

	p.Next()

	stmt.ReturnValue = p.parseExpression(LOWEST)

	if p.nextTokenIs(token.SEMICOLON) {
		p.Next()
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	forStmt := &ast.ForStatement{Token: p.curTok}
	p.Next()

	stmts := []ast.Statement{}
	if p.curTokenIs(token.LBRACE) {
		forStmt.Condition = stmts
		forStmt.Body = p.parseBlockStatement()
	} else {
		for !p.curTokenIs(token.LBRACE) {
			stmt := p.parseStatement()
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
			p.Next()
			if p.curTokenIs(token.SEMICOLON) {
				p.Next()
			}
		}

		forStmt.Condition = stmts
		forStmt.Body = p.parseBlockStatement()
	}

	if p.nextTokenIs(token.SEMICOLON) {
		p.Next()
	}
	return forStmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curTok}
	stmt.Expression = p.parseExpression(LOWEST)

	if p.nextTokenIs(token.SEMICOLON) {
		p.Next()
	}

	return stmt
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found", t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseExpression(pb int) ast.Expression {
	prefix := p.prefixFNs[p.curTok.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curTok.Type)
		return nil
	}
	leftExp := prefix()

	for !p.nextTokenIs(token.SEMICOLON) && pb < p.peekPowerBind() {
		infix := p.infixFNs[p.nextTok.Type]
		if infix == nil {
			return leftExp
		}

		p.Next()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curTok,
		Operator: p.curTok.Literal,
	}

	p.Next()

	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curTok,
		Operator: p.curTok.Literal,
		Left:     left,
	}

	precedence := p.curPowerBind()
	p.Next()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.Next()
	exp := p.parseExpression(LOWEST)
	if !p.expectNext(token.RPAREN) {
		return nil
	}
	return exp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseBoolean() ast.Expression {
	return &ast.BooleanLiteral{Token: p.curTok, Value: p.curTokenIs(token.TRUE)}
}

func (p *Parser) parseNil() ast.Expression {
	return &ast.NilLiteral{Token: p.curTok, Value: p.curTok.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curTok}

	value, err := strconv.ParseInt(p.curTok.Literal, 0, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curTok.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value

	return lit
}

func (p *Parser) parseIfExpression() ast.Expression {
	expression := &ast.IfExpression{Token: p.curTok}

	p.Next()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectNext(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()

	// for else block
	if p.nextTokenIs(token.ELSE) {
		p.Next()

		if p.nextTokenIs(token.IF) {
			p.Next()
			expression = p.parseElifExpression(expression)
		} else if !p.expectNext(token.LBRACE) {
			return nil
		} else {
			expression.Alternative = p.parseBlockStatement()
		}
	}

	return expression
}

func (p *Parser) parseElifExpression(ifexp *ast.IfExpression) *ast.IfExpression {
	expression := &ast.IfExpression{Token: p.curTok}

	p.Next()
	expression.Condition = p.parseExpression(LOWEST)

	if !p.expectNext(token.LBRACE) {
		return nil
	}

	expression.Consequence = p.parseBlockStatement()
	ifexp.Options = append(ifexp.Options, expression)

	if p.nextTokenIs(token.ELSE) {
		p.Next()

		if p.nextTokenIs(token.IF) {
			p.Next()
			return p.parseElifExpression(ifexp)
		} else if !p.expectNext(token.LBRACE) {
			return nil
		} else {
			ifexp.Alternative = p.parseBlockStatement()
		}
	}

	return ifexp
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curTok}
	block.Statements = []ast.Statement{}

	p.Next()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.Next()
	}

	if p.nextTokenIs(token.SEMICOLON) {
		p.Next()
	}
	return block
}

func (p *Parser) parseFunctionLiteral() ast.Expression {
	fn := &ast.FunctionLiteral{Token: p.curTok}

	if p.nextTokenIs(token.ID) {
		p.Next()
		fn.Name = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	}

	if !p.expectNext(token.LPAREN) {
		return nil
	}

	fn.Args = p.parseFunctionArgs()

	if p.nextTokenIs(token.ID) {
		p.Next()
		fn.ReturnType = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	}

	if !p.expectNext(token.LBRACE) {
		return nil
	}

	fn.Body = p.parseBlockStatement()

	return fn
}

func (p *Parser) parseFunctionArgs() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	if p.nextTokenIs(token.RPAREN) {
		p.Next()
		return identifiers
	}

	p.Next() // curTok => id
	ident := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	if !p.expectNext(token.ID) { // require type for arg
		return nil
	}

	// current token is type for the previous id
	ident.Type = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
	identifiers = append(identifiers, ident)

	for p.nextTokenIs(token.COMMA) {

		p.Next()
		p.Next()
		ident := &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		if !p.expectNext(token.ID) {
			return nil
		}
		ident.Type = &ast.Identifier{Token: p.curTok, Value: p.curTok.Literal}
		identifiers = append(identifiers, ident)
	}

	if !p.expectNext(token.RPAREN) {
		return nil
	}

	return identifiers
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curTok, Function: function}
	exp.Args = p.parseCallArguments()
	return exp
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.nextTokenIs(token.RPAREN) {
		p.Next()
		return args
	}

	p.Next()
	args = append(args, p.parseExpression(LOWEST))

	for p.nextTokenIs(token.COMMA) {
		p.Next()
		p.Next()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectNext(token.RPAREN) {
		return nil
	}

	// fmt.Println(args)
	return args
}

func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curTok.Type == t
}

func (p *Parser) nextTokenIs(t token.TokenType) bool {
	return p.nextTok.Type == t
}

func (p *Parser) expectNext(t token.TokenType) bool {
	if p.nextTokenIs(t) {
		p.Next()
		return true
	} else {
		p.peekError(t)
		return false
	}
}

func (p *Parser) peekPowerBind() int {
	if p, ok := powerBind[p.nextTok.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) curPowerBind() int {
	if p, ok := powerBind[p.curTok.Type]; ok {
		return p
	}

	return LOWEST
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixFN) {
	p.prefixFNs[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixFN) {
	p.infixFNs[tokenType] = fn
}

var powerBind = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.MOD:      PRODUCT,
	token.LPAREN:   CALL,
}

const (
	_ int = iota
	LOWEST
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
)
