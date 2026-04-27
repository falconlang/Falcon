package ast

type Expr interface {
	String() string
	Blockly(flags ...bool) Block
	Continuous() bool
	Consumable() bool
	Signature() []Signature
}

func (b *Block) String() string {
	return "<" + b.Type + ">"
}

func (b *Block) GetType() string {
	return b.Type
}

func (b *Block) Order() int {
	return 100
}

func (b *Block) SingleValue() Block {
	return b.Values[0].Block
}

func (b *Block) SingleField() string {
	return b.Fields[0].Value
}

func (b *Block) SingleStatement() Statement {
	return b.Statements[0]
}

func (b *Block) Statement() Statement {
	return b.Statements[0]
}
