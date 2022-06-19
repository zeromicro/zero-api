package parser

import "github.com/zeromicro/zero-api/token"

type bailout struct {
}

func (p *parser) error(pos token.Pos, msg string) {
	if p.trace {
		defer un(trace(p, "error: "+msg))
	}

	position := p.file.Position(pos)
	if p.mode&AllErrors == 0 {
		n := len(p.errors)
		if n > 0 && p.errors[n-1].Pos.Line == position.Line {
			return
		}
		if n > 10 {
			panic(bailout{})
		}
	}
	p.errors.Add(position, msg)
}

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok == token.SEMICOLON && p.lit == "\n":
			msg += ", found newline"
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.error(pos, msg)
}
