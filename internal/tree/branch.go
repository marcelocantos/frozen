package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/errors"
	"github.com/arr-ai/frozen/internal/depth"
	"github.com/arr-ai/frozen/internal/fmtutil"
)

const (
	fanoutBits = depth.FanoutBits
	fanout     = depth.Fanout
)

var (
	// UseRHS returns its RHS arg.
	UseRHS = func(_, b elementT) elementT { return b }

	// UseLHS returns its LHS arg.
	UseLHS = func(a, _ elementT) elementT { return a }
)

type branch struct {
	p packer
}

func (b *branch) Canonical(_ int) node {
	var buf [maxLeafLen]elementT
	if data := b.CopyTo(buf[:0]); data != nil {
		return newLeaf(append([]elementT{}, data...)...)
	}
	return b
}

func (b *branch) Combine(args *CombineArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case *leaf:
		ret := node(b)
		for _, e := range n.data {
			ret = ret.With(args, e, depth, newHasher(e, depth), matches)
		}
		return ret
	case *branch:
		ret := &branch{}
		args.Parallel(depth, matches, func(i int, matches *int) bool {
			ret.p[i] = b.p.Get(i).Combine(args, n.p.Get(i), depth+1, matches)
			return true
		})
		return ret
	default:
		panic(errors.WTF)
	}
}

func (b *branch) CopyTo(dest []elementT) []elementT {
	for _, child := range b.p {
		if child != nil {
			if dest = child.CopyTo(dest); dest == nil {
				break
			}
		}
	}
	return dest
}

func (b *branch) Defrost() unNode {
	u := newUnBranch()
	for i, e := range b.p {
		if e != nil {
			u.p[i] = e.Defrost()
		}
	}
	return u
}

func (b *branch) Difference(args *EqArgs, n node, depth int, removed *int) node {
	switch n := n.(type) {
	case *leaf:
		result := node(b)
		for _, e := range n.data {
			result = result.Without(args, e, depth, newHasher(e, depth), removed)
		}
		return result
	case *branch:
		var ret branch
		args.Parallel(depth, removed, func(i int, removed *int) bool {
			ret.p[i] = b.p.Get(i).Difference(args, n.p.Get(i), depth+1, removed)
			return true
		})
		return ret.Canonical(depth)
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Empty() bool {
	return false
}

func (b *branch) Equal(args *EqArgs, n node, depth int) bool {
	if n, is := n.(*branch); is {
		return args.Parallel(depth, nil, func(i int, _ *int) bool {
			return b.p.Get(i).Equal(args, n.p.Get(i), depth+1)
		})
	}
	return false
}

func (b *branch) Get(args *EqArgs, v elementT, h hasher) *elementT {
	return b.p.Get(h.hash()).Get(args, v, h.next())
}

func (b *branch) Intersection(args *EqArgs, n node, depth int, matches *int) node {
	switch n := n.(type) {
	case *leaf:
		return n.Intersection(args.flip, b, depth, matches)
	case *branch:
		var ret branch
		args.Parallel(depth, matches, func(i int, matches *int) bool {
			ret.p[i] = b.p.Get(i).Intersection(args, n.p.Get(i), depth+1, matches)
			return true
		})
		return ret.Canonical(depth)
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Iterator(buf [][]node) Iterator {
	return b.p.Iterator(buf)
}

func (b *branch) Reduce(args NodeArgs, depth int, r func(values ...elementT) elementT) elementT {
	var results [fanout]elementT
	args.Parallel(depth, nil, func(i int, _ *int) bool {
		if n := b.p.Get(i); !n.Empty() {
			results[i] = n.Reduce(args, depth+1, r)
		}
		return true
	})

	results2 := results[:0]
	for _, r := range results {
		if !isBlank(r) {
			results2 = append(results2, r)
		}
	}
	return r(results2...)
}

func (b *branch) SubsetOf(args *EqArgs, n node, depth int) bool {
	switch n := n.(type) {
	case *leaf:
		return false
	case *branch:
		return args.Parallel(depth, nil, func(i int, _ *int) bool {
			return b.p.Get(i).SubsetOf(args, n.p.Get(i), depth+1)
		})
	default:
		panic(errors.WTF)
	}
}

func (b *branch) Transform(args *CombineArgs, depth int, count *int, f func(e elementT) elementT) node {
	var nodes [fanout]node
	args.Parallel(depth, count, func(i int, count *int) bool {
		nodes[i] = b.p.Get(i).Transform(args, depth+1, count, f)
		return true
	})

	// log.Printf("%*s%v", 4*depth, "", nodes[0])
	acc := nodes[0]
	var duplicates int
	for _, n := range nodes[1:] {
		acc = acc.Combine(args, n, 0, &duplicates)
		// log.Printf("%*s%v -> %v", 4*depth, "", n, acc)
	}
	*count -= duplicates
	return acc
}

func (b *branch) Where(args *WhereArgs, depth int, matches *int) node {
	var nodes [fanout]node
	args.Parallel(depth, matches, func(i int, matches *int) bool {
		nodes[i] = b.p.Get(i).Where(args, depth+1, matches)
		return true
	})
	return (&branch{p: nodes}).Canonical(depth)
}

func (b *branch) With(args *CombineArgs, v elementT, depth int, h hasher, matches *int) node {
	i := h.hash()
	return &branch{p: b.p.With(i, b.p.Get(i).With(args, v, depth+1, h.next(), matches))}
}

func (b *branch) Without(args *EqArgs, v elementT, depth int, h hasher, matches *int) node {
	i := h.hash()
	child := b.p.Get(i).Without(args, v, depth+1, h.next(), matches)
	return (&branch{p: b.p.With(i, child)}).Canonical(depth)
}

var branchStringIndices = []string{
	"⁰", "¹", "²", "³", "⁴", "⁵", "⁶", "⁷", "⁸", "⁹",
	"¹⁰", "¹¹", "¹²", "¹³", "¹⁴", "¹⁵",
}

func (b *branch) Format(f fmt.State, _ rune) {
	total := 0
	printf := func(format string, args ...interface{}) {
		n, err := fmt.Fprintf(f, format, args...)
		if err != nil {
			panic(err)
		}
		total += n
	}
	write := func(b []byte) {
		n, err := f.Write(b)
		if err != nil {
			panic(err)
		}
		total += n
	}
	write([]byte("⁅"))

	var buf [20]elementT
	deep := b.CopyTo(buf[:]) != nil

	if deep {
		write([]byte("\n"))
	}

	for i, child := range b.p {
		if b.p.Get(i).Empty() {
			continue
		}
		index := branchStringIndices[i]
		if deep {
			printf("   %s%s\n", index, fmtutil.IndentBlock(child.String()))
		} else {
			if i > 0 {
				write([]byte(" "))
			}
			printf("%s%v", index, child)
		}
	}
	write([]byte("⁆"))

	fmtutil.PadFormat(f, total)
}

func (b *branch) String() string {
	return fmt.Sprintf("%s", b)
}
