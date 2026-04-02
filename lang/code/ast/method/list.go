package method

import (
	"Falcon/code/ast"
)

func (c *Call) listMethods(signature *CallSignature) ast.Block {
	switch signature.BlocklyName {
	case "lists_length", "lists_pick_random_item", "lists_reverse", "lists_to_csv_row",
		"lists_to_csv_table", "lists_sort", "lists_but_first", "lists_but_last":
		return c.simpleOperand(signature.BlocklyName, "LIST")
	case "dictionaries_alist_to_dict":
		return c.simpleOperand(signature.BlocklyName, "PAIRS")
	case "lists_add_items":
		return c.listAdd()
	case "lists_is_in":
		return c.listContainsItem()
	case "lists_position_in":
		return c.listIndexOf()
	case "lists_insert_item":
		return c.listInsertItem()
	case "lists_remove_item":
		return c.listRemoveAt()
	case "lists_append_list":
		return c.listAppendList()
	case "lists_lookup_in_pairs":
		return c.listLookupInPairs()
	case "lists_join_with_separator":
		return c.listJoin()
	case "lists_slice":
		return c.listSlice()
	default:
		c.Where.Error("Unknown list method: %", signature.BlocklyName)
		panic("")
	}
}

func (c *Call) listSlice() ast.Block {
	return ast.Block{
		Type:   "lists_slice",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "INDEX1", "INDEX2"),
	}
}

func (c *Call) listJoin() ast.Block {
	return ast.Block{
		Type:   "lists_join_with_separator",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "SEPARATOR"),
	}
}

func (c *Call) listLookupInPairs() ast.Block {
	return ast.Block{
		Type:   "lists_lookup_in_pairs",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "KEY", "NOTFOUND"),
	}
}

func (c *Call) listAppendList() ast.Block {
	return ast.Block{
		Type:   "lists_append_list",
		Values: ast.MakeValueArgs(c.On, "LIST0", c.Args, "LIST1"),
	}
}

func (c *Call) listRemoveAt() ast.Block {
	return ast.Block{
		Type:   "lists_remove_item",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "INDEX"),
	}
}

func (c *Call) listInsertItem() ast.Block {
	return ast.Block{
		Type:   "lists_insert_item",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "INDEX", "ITEM"),
	}
}

func (c *Call) listIndexOf() ast.Block {
	return ast.Block{
		Type:   "lists_position_in",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "ITEM"),
	}
}

func (c *Call) listContainsItem() ast.Block {
	return ast.Block{
		Type:   "lists_is_in",
		Values: ast.MakeValueArgs(c.On, "LIST", c.Args, "ITEM"),
	}
}

func (c *Call) listAdd() ast.Block {
	return ast.Block{
		Type:     "lists_add_items",
		Mutation: &ast.Mutation{ItemCount: len(c.Args)},
		Values:   ast.ValueArgsByPrefix(c.On, "LIST", "ITEM", c.Args),
	}
}
