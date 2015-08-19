package ds

//EdgeAttr provides a evaluator for checking a section attribute existence
func EdgeAttr(attr string) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil && soc.Attrs.Has(attr) {
			return true
		}
		return false
	}
}

//EdgeKeyValue provides a evaluator for checking a section attribute existence
func EdgeKeyValue(key string, val interface{}) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return soc.HasMatch(key, val)
		}
		return false
	}
}

//EdgeKey provides a evaluator for checking a section attribute existence
func EdgeKey(key string) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return soc.Has(key)
		}
		return false
	}
}

//EdgeWeight provides a evaluator for checking a section attribute existence
func EdgeWeight(w int) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return soc.Weight == w
		}
		return false
	}
}

//OnlyDepth provides a evaluator for checking a section attribute existence
func OnlyDepth(w int) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return depth == w
		}
		return false
	}
}

//MaxDepth provides a evaluator for checking a section attribute existence
func MaxDepth(w int) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return depth <= w
		}
		return false
	}
}

//MinDepth provides a evaluator for checking a section attribute existence
func MinDepth(w int) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return depth >= w
		}
		return false
	}
}

//WithinDepth provides a evaluator for checking a section attribute existence
func WithinDepth(min, max int) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return (min > depth && depth < max)
		}
		return false
	}
}

//DepthRange provides a evaluator for checking a section attribute existence
func DepthRange(min, max int) NodeEval {
	return func(n Nodes, soc *Socket, depth int) bool {
		if soc != nil {
			return (min >= depth && depth <= max)
		}
		return false
	}
}
