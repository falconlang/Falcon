package ast

func (i Signature) String() string {
	switch i {
	case SignBool:
		return "boolean"
	case SignNumb:
		return "number"
	case SignText:
		return "text"
	case SignList:
		return "list"
	case SignDict:
		return "dictionary"
	case SignComponent:
		return "component"
	case SignHelper:
		return "helper"
	case SignAny:
		return "any"
	case SignOfEvent:
		return "event"
	case SignVoid:
		return "void"
	default:
		return "unknown"
	}
}
