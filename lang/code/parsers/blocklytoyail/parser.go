package blocklytoyail

type Parser struct {
	xmlContent string
}

func NewParser(xmlContent string) *Parser {
	return &Parser{xmlContent: xmlContent}
}
