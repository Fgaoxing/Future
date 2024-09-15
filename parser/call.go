package parser

import (
	"errors"
	"future/lexer"
	typeSys "future/type"
)

type CallBlock struct {
	Name string
	Args []*ArgBlock
	Func *FuncBlock
}

func (c *CallBlock) Parse(p *Parser) {
	// 找到定义位置
	oldThisBlock := p.ThisBlock
	for {
		if p.ThisBlock.Father == nil && p.ThisBlock.Value == nil {
			// 查找根级内容
			for i := 0; i < len(p.ThisBlock.Children); i++ {
				switch p.ThisBlock.Children[i].Value.(type) {
				case *FuncBlock:
					tmp := p.ThisBlock.Children[i].Value.(*FuncBlock)
					if tmp.Name == c.Name {
						c.Func = tmp
						goto end
					}
				}
			}
			p.Error.MissError("Syntax Error", p.Lexer.Cursor, "need define "+c.Name)
		}
		for i := 0; i < len(p.ThisBlock.Children); i++ {
			switch p.ThisBlock.Children[i].Value.(type) {
			case *FuncBlock:
				tmp := p.ThisBlock.Children[i].Value.(*FuncBlock)
				if tmp.Name == c.Name {
					c.Func = tmp
					goto end
				}
			}
		}
		p.ThisBlock = p.ThisBlock.Father
	}
end:
	p.ThisBlock = oldThisBlock
	// 解析括号
	rightBra := p.FindRightBracket(false)
	for p.Lexer.Cursor+1 < rightBra {
		oldCursor := p.Lexer.Cursor + 1
		sepCursor := p.Has(lexer.Token{Type: lexer.LexTokenType["SEPARATOR"], Value: ","}, rightBra)
		if sepCursor == -1 {
			arg := &ArgBlock{Value: p.ParseExpression(rightBra-1)}
			arg.Type = arg.Value.Type
			c.Args = append(c.Args, arg)
			if len(c.Func.Args) < 1 {
				p.Error.MissErrors("Call Error", oldCursor, rightBra+1, "Args length error")
			}
			arg.Defind = c.Func.Args[len(c.Args)-1]
			if typeSys.AutoType(arg.Type, arg.Defind.Type, true) {
				arg.Type = arg.Defind.Type
			} else {
				p.Error.MissErrors("Type Error", oldCursor, rightBra+1, "need type "+arg.Defind.Type.Type()+", not "+arg.Value.Type.Type())
			}
			break
		}
		arg := &ArgBlock{Value: p.ParseExpression(sepCursor-1)}
		arg.Type = arg.Value.Type
		p.Lexer.Cursor++
		c.Args = append(c.Args, arg)
		if len(c.Func.Args) <= len(c.Args) {
			p.Error.MissErrors("Call Error", oldCursor, rightBra+1, "Args length error")
		}
		arg.Defind = c.Func.Args[len(c.Args)-1]
		if typeSys.AutoType(arg.Type, arg.Defind.Type, true) {
			arg.Type = arg.Defind.Type
		} else {
			p.Error.MissErrors("Type Error", oldCursor, sepCursor+1, "need type "+arg.Defind.Type.Type()+", not "+arg.Value.Type.Type())
		}
	}
	if err := c.ParseArgsDefault(p); err != nil {
		p.Error.MissError("Call Error", rightBra-1, err.Error())
	}
	p.ThisBlock.AddChild(&Node{Value: c})
	// 查找父级内容，找到定义位置
	p.Lexer.Cursor++

}

func (c *CallBlock) ParseArgsDefault(p *Parser) error {
	for i := 0; i < len(c.Func.Args); i++ {
		if len(c.Args) <= i && c.Func.Args[i].Default == nil {
			return errors.New("Args length error")
		} else {
			c.Args = append(c.Args, &ArgBlock{Value: c.Func.Args[i].Default, Defind: c.Func.Args[i]})
		}
	}
	return nil
}
