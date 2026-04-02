package method

import (
	"Falcon/code/ast"
)

func (c *Call) dictMethods(signature *CallSignature) ast.Block {
	switch signature.BlocklyName {
	case "dictionaries_length", "dictionaries_dict_to_alist":
		return c.simpleOperand(signature.BlocklyName, "DICT")
	case "dictionaries_lookup":
		return c.dictGet()
	case "dictionaries_set_pair":
		return c.dictSet()
	case "dictionaries_delete_pair":
		return c.dictDelete()
	case "dictionaries_recursive_lookup":
		return c.dictGetAtPath()
	case "dictionaries_recursive_set":
		return c.dictSetAtPath()
	case "dictionaries_is_key_in":
		return c.dictContainsKey()
	case "dictionaries_combine_dicts":
		return c.dictMergeInto()
	case "dictionaries_walk_tree":
		return c.dictWalkTree()
	case "dictionaries_getters":
		return c.dictGetters()
	default:
		c.Where.Error("Unknown dict method: %", signature.BlocklyName)
		panic("")
	}
}

func (c *Call) dictGetters() ast.Block {
	var fieldOp string
	if c.Name == "values" {
		fieldOp = "VALUES"
	} else {
		fieldOp = "KEYS"
	}
	return ast.Block{
		Type:   "dictionaries_getters",
		Fields: []ast.Field{{Name: "OP", Value: fieldOp}},
		Values: []ast.Value{{Name: "DICT", Block: c.On.Blockly(false)}},
	}
}

func (c *Call) dictWalkTree() ast.Block {
	return ast.Block{
		Type:   "dictionaries_walk_tree",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "PATH"),
	}
}

func (c *Call) dictMergeInto() ast.Block {
	return ast.Block{
		Type: "dictionaries_combine_dicts",
		Values: []ast.Value{
			{Name: "DICT1", Block: c.Args[0].Blockly(false)},
			{Name: "DICT2", Block: c.On.Blockly(false)},
		},
	}
}

func (c *Call) dictContainsKey() ast.Block {
	return ast.Block{
		Type:   "dictionaries_is_key_in",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "KEY"),
	}
}

func (c *Call) dictSetAtPath() ast.Block {
	return ast.Block{
		Type:   "dictionaries_recursive_set",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "KEYS", "VALUE"),
	}
}

func (c *Call) dictGetAtPath() ast.Block {
	return ast.Block{
		Type:   "dictionaries_recursive_lookup",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "KEYS", "NOTFOUND"),
	}
}

func (c *Call) dictDelete() ast.Block {
	return ast.Block{
		Type:   "dictionaries_delete_pair",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "KEY"),
	}
}

func (c *Call) dictSet() ast.Block {
	return ast.Block{
		Type:   "dictionaries_set_pair",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "KEY", "VALUE"),
	}
}

func (c *Call) dictGet() ast.Block {
	return ast.Block{
		Type:   "dictionaries_lookup",
		Values: ast.MakeValueArgs(c.On, "DICT", c.Args, "KEY", "NOTFOUND"),
	}
}
