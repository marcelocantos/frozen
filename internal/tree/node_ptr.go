package tree

func (n *node) String() string {
	if l := n.Leaf(); l != nil {
		return l.String()
	}
	return n.b.String()
}

func (n *node) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Add(args, v, depth, h, matches)
	}
	return n.b.Add(args, v, depth, h, matches)
}

func (n *node) AppendTo(dest []elementT) []elementT {
	if l := n.Leaf(); l != nil {
		return l.AppendTo(dest)
	}
	return n.b.AppendTo(dest)
}

func (n *node) Canonical(depth int) *node {
	if l := n.Leaf(); l != nil {
		return l.Canonical(depth)
	}
	return n.b.Canonical(depth)
}

func (n *node) Combine(args *CombineArgs, n2 *node, depth int, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Combine(args, n2, depth, matches)
	}
	return n.b.Combine(args, n2, depth, matches)
}

func (n *node) Difference(args *EqArgs, n2 *node, depth int, removed *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Difference(args, n2, depth, removed)
	}
	return n.b.Difference(args, n2, depth, removed)
}

func (n *node) Empty() bool {
	if l := n.Leaf(); l != nil {
		return l.Empty()
	}
	return n.b.Empty()
}

func (n *node) Equal(args *EqArgs, n2 *node, depth int) bool {
	if l := n.Leaf(); l != nil {
		return l.Equal(args, n2, depth)
	}
	return n.b.Equal(args, n2, depth)
}

func (n *node) Get(args *EqArgs, v elementT, h hasher) *elementT {
	if l := n.Leaf(); l != nil {
		return l.Get(args, v, h)
	}
	return n.b.Get(args, v, h)
}

func (n *node) Intersection(args *EqArgs, n2 *node, depth int, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Intersection(args, n2, depth, matches)
	}
	return n.b.Intersection(args, n2, depth, matches)
}

func (n *node) Iterator(buf [][]*node) Iterator {
	if l := n.Leaf(); l != nil {
		return l.Iterator(buf)
	}
	return n.b.Iterator(buf)
}

func (n *node) Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT {
	if l := n.Leaf(); l != nil {
		return l.Reduce(args, depth, r)
	}
	return n.b.Reduce(args, depth, r)
}

func (n *node) SubsetOf(args *EqArgs, n2 *node, depth int) bool {
	if l := n.Leaf(); l != nil {
		return l.SubsetOf(args, n2, depth)
	}
	return n.b.SubsetOf(args, n2, depth)
}

func (n *node) Transform(args *CombineArgs, depth int, count *int, f func(v elementT) elementT) *node {
	if l := n.Leaf(); l != nil {
		return l.Transform(args, depth, count, f)
	}
	return n.b.Transform(args, depth, count, f)
}

func (n *node) Where(args *WhereArgs, depth int, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Where(args, depth, matches)
	}
	return n.b.Where(args, depth, matches)
}

func (n *node) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.With(args, v, depth, h, matches)
	}
	return n.b.With(args, v, depth, h, matches)
}

func (n *node) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Without(args, v, depth, h, matches)
	}
	return n.b.Without(args, v, depth, h, matches)
}

func (n *node) Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) *node {
	if l := n.Leaf(); l != nil {
		return l.Remove(args, v, depth, h, matches)
	}
	return n.b.Remove(args, v, depth, h, matches)
}