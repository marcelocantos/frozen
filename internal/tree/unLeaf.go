package tree

type unLeaf []elementT

var _ unNode = &unLeaf{}

func newUnLeaf() unLeaf {
	return make(unLeaf, 0, maxLeafLen)
}

func (l *unLeaf) Add(args *CombineArgs, v elementT, depth int, h hasher, matches *int) unNode {
	for i, e := range *l {
		if args.eq(e, v) {
			*matches++
			(*l)[i] = args.f(e, v)
			return l
		}
	}
	if len(*l) < cap(*l) || depth >= maxTreeDepth {
		*l = append(*l, v)
		return l
	}

	b := newUnBranch()
	for _, e := range *l {
		b.Add(args, e, depth, newHasher(e, depth), matches)
	}
	b.Add(args, v, depth, h, matches)

	return b
}

func (l unLeaf) appendTo(dest []elementT) []elementT {
	if len(dest)+len(l) > cap(dest) {
		return nil
	}
	return append(dest, l...)
}

func (l unLeaf) Freeze() node {
	return newLeaf(append(make([]elementT, 0, len(l)), l...)...)
}

func (l unLeaf) Get(args *EqArgs, v elementT, h hasher) *elementT {
	for i, e := range l {
		if args.eq(e, v) {
			return &(l)[i]
		}
	}
	return nil
}

func (l *unLeaf) Remove(args *EqArgs, v elementT, depth int, h hasher, matches *int) unNode {
	for i, e := range *l {
		if args.eq(e, v) {
			*matches++
			last := len(*l) - 1
			if last == 0 {
				return unEmptyNode{}
			}
			if i < last {
				(*l)[i] = (*l)[last]
			}
			*l = (*l)[:last]
			return l
		}
	}
	return l
}
