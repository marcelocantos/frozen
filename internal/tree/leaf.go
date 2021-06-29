package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/fu"
)

type leaf struct {
	data [2]elementT
}

func newLeaf(data ...elementT) *leaf {
	switch len(data) {
	case 1:
		return newLeaf1(data[0])
	case 2:
		return newLeaf2(data[0], data[1])
	default:
		panic(errors.Errorf("data wrong size (%d) for leaf", len(data)))
	}
}

func newLeaf1(a elementT) *leaf {
	return newLeaf2(a, zero)
}

func newLeaf2(a, b elementT) *leaf {
	if a == zero {
		panic(errors.WTF)
	}
	return &leaf{data: [2]elementT{a, b}}
}

// fmt.Formatter

func (l *leaf) Format(f fmt.State, verb rune) {
	fu.WriteString(f, "(")
	if l.data[0] != zero {
		fu.Format(l.data[0], f, verb)
		if l.data[1] != zero {
			fu.WriteString(f, ",")
			fu.Format(l.data[1], f, verb)
		}
	}
	fu.WriteString(f, ")")
}

// fmt.Stringer

func (l *leaf) String() string {
	return fmt.Sprintf("%s", l)
}

// node

func (l *leaf) Add(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	switch {
	case args.eq(l.data[0], v):
		l.data[0] = args.f(l.data[0], v)
		matches++
	case l.data[1] == zero:
		l.data[1] = v
	case args.eq(l.data[1], v):
		l.data[1] = args.f(l.data[1], v)
		matches++
	case depth >= maxTreeDepth:
		return newTwig(l.data[0], l.data[1], v), 0
	default:
		return newBranchFrom(depth, l.data[0], l.data[1], v), 0
	}
	return l, matches
}

func (l *leaf) Canonical(depth int) node {
	return l
}

func (l *leaf) Combine(args *CombineArgs, n node, depth int) (_ node, matches int) { //nolint:cyclop
	switch n := n.(type) {
	case *branch:
		return n.Combine(args.flip, l, depth)
	case *leaf:
		lr := func(a, b int) int { return a<<2 | b }
		masks := lr(l.mask(), n.mask())
		if masks == lr(3, 1) {
			masks, l, n, args = lr(1, 3), n, l, args.flip
		}
		l0, l1 := l.data[0], l.data[1]
		n0, n1 := n.data[0], n.data[1]
		if args.eq(l0, n0) { //nolint:nestif
			matches++
			switch masks {
			case lr(1, 1):
				return newLeaf1(args.f(l0, n0)), matches
			case lr(1, 3):
				return newLeaf2(args.f(l0, n0), n1), matches
			default:
				if args.eq(l1, n1) {
					matches++
					return newLeaf2(args.f(l0, n0), args.f(l1, n1)), matches
				}
			}
		} else {
			switch masks {
			case lr(1, 1):
				return newLeaf2(l0, n0), matches
			case lr(1, 3):
				if args.eq(l0, n1) {
					matches++
					return newLeaf2(n0, args.f(l0, n1)), matches
				}
				return newBranchFrom(depth, l0, n0, n1), matches
			default:
				if args.eq(l1, n1) {
					matches++
					return newBranchFrom(depth, l0, n0, args.f(l1, n1)), matches
				}
				if args.eq(l0, n1) {
					matches++
					if args.eq(l1, n0) {
						matches++
						return newLeaf2(args.f(l0, n1), args.f(l1, n0)), matches
					}
					return newBranchFrom(depth, args.f(l0, n1), l1, n0), matches
				}
				if args.eq(l1, n0) {
					matches++
					return newBranchFrom(depth, l0, n1, args.f(l1, n0)), matches
				}
			}
		}
		return newBranchFrom(depth, l0, l1, n0, n1), matches
	case *twig:
		return n.Combine(args.flip, l, depth)
	default:
		panic(errors.WTF)
	}
}

func (l *leaf) AppendTo(dest []elementT) []elementT {
	data := l.slice()
	if len(dest)+len(data) > cap(dest) {
		return nil
	}
	return append(dest, data...)
}

func (l *leaf) Difference(args *EqArgs, n node, depth int) (_ node, matches int) {
	mask := l.mask()
	if n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil {
		matches++
		mask &^= 0b01
	}

	if l.data[1] != zero {
		if n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil {
			matches++
			mask &^= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf) Empty() bool {
	return false
}

func (l *leaf) Equal(args *EqArgs, n node, depth int) bool {
	if n, is := n.(*leaf); is {
		lm, nm := l.mask(), n.mask()
		if lm != nm {
			return false
		}
		l0, l1 := l.data[0], l.data[1]
		n0, n1 := n.data[0], n.data[1]
		if lm == 1 && nm == 1 {
			return args.eq(l0, n0)
		}
		return args.eq(l0, n0) && args.eq(l1, n1) ||
			args.eq(l0, n1) && args.eq(l1, n0)
	}
	return false
}

func (l *leaf) Get(args *EqArgs, v elementT, _ hasher) *elementT {
	for i, e := range l.slice() {
		if args.eq(e, v) {
			return &l.data[i]
		}
	}
	return nil
}

func (l *leaf) Intersection(args *EqArgs, n node, depth int) (_ node, matches int) {
	mask := 0
	if n.Get(args, l.data[0], newHasher(l.data[0], depth)) != nil {
		matches++
		mask |= 0b01
	}

	if l.data[1] != zero {
		if n.Get(args, l.data[1], newHasher(l.data[1], depth)) != nil {
			matches++
			mask |= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf) Iterator([][]node) Iterator {
	return newSliceIterator(l.slice())
}

func (l *leaf) Reduce(_ NodeArgs, _ int, r func(values ...elementT) elementT) elementT {
	return r(l.slice()...)
}

func (l *leaf) Remove(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	if args.eq(l.data[0], v) {
		matches++
		if l.data[1] == zero {
			return nil, matches
		}
		l.data = [2]elementT{l.data[1], zero}
	} else if l.data[1] != zero {
		if args.eq(l.data[1], v) {
			matches++
			l.data[1] = zero
		}
	}
	return l, matches
}

func (l *leaf) SubsetOf(args *EqArgs, n node, depth int) bool {
	a := l.data[0]
	h := newHasher(a, depth)
	if n.Get(args, a, h) == nil {
		return false
	}
	if b := l.data[1]; b != zero {
		h := newHasher(b, depth)
		if n.Get(args, b, h) == nil {
			return false
		}
	}
	return true
}

func (l *leaf) Map(args *CombineArgs, _ int, f func(e elementT) elementT) (_ node, matches int) {
	var nb Builder
	for _, e := range l.slice() {
		nb.Add(args, f(e))
	}
	t := nb.Finish()
	matches += t.count
	return t.root, matches
}

func (l *leaf) Vet() {
	if l.data[0] == zero {
		if l.data[1] != zero {
			panic(errors.Errorf("data only in leaf slot 1"))
		}
		panic(errors.Errorf("empty leaf"))
	}
}

func (l *leaf) Where(args *WhereArgs, depth int) (_ node, matches int) {
	var mask int
	if args.Pred(l.data[0]) {
		matches++
		mask ^= 0b01
	}
	if l.data[1] != zero {
		if args.Pred(l.data[1]) {
			matches++
			mask ^= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf) where(mask int) node {
	switch mask {
	case 0b00:
		return nil
	case 0b01:
		if l.data[1] == zero {
			return l
		}
		return newLeaf1(l.data[0])
	case 0b10:
		return newLeaf1(l.data[1])
	default:
		return l
	}
}

func (l *leaf) With(args *CombineArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	switch {
	case args.eq(l.data[0], v):
		matches++
		return newLeaf2(args.f(l.data[0], v), l.data[1]), matches
	case l.data[1] == zero:
		return newLeaf2(l.data[0], v), 0
	case args.eq(l.data[1], v):
		matches++
		return newLeaf2(l.data[0], args.f(l.data[1], v)), matches
	case depth >= maxTreeDepth:
		return newTwig(append(l.data[:], v)...), matches
	default:
		return newBranchFrom(depth, l.data[0], l.data[1], v), matches
	}
}

func (l *leaf) Without(args *EqArgs, v elementT, depth int, h hasher) (_ node, matches int) {
	mask := l.mask()
	if args.eq(l.data[0], v) {
		matches++
		mask ^= 0b01
	} else if l.data[1] != zero {
		if args.eq(l.data[1], v) {
			matches++
			mask ^= 0b10
		}
	}
	return l.where(mask), matches
}

func (l *leaf) count() int {
	if l.data[1] == zero {
		return 1
	}
	return 2
}

func (l *leaf) mask() int {
	if l.data[1] == zero {
		return 1
	}
	return 3
}

func (l *leaf) slice() []elementT {
	return l.data[:l.count()]
}
