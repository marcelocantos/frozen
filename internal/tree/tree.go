package tree

import (
	"container/heap"
	"math/bits"

	"github.com/arr-ai/frozen/internal/depth"
)

type Tree struct {
	root  noderef
	count int
}

func newTree(n noderef, count *int) Tree {
	return Tree{root: n, count: *count}
}

func newTreeNeg(n noderef, count *int) Tree {
	return Tree{root: n, count: -*count}
}

func (t Tree) Root() noderef {
	if t.root == nil {
		t.root = theEmptyNode
	}
	return t.root
}

func (t Tree) MutableRoot() noderef {
	if t.root == nil {
		t.root = newMutableLeaf().Node()
	}
	return t.root
}

func (t *Tree) Add(args *CombineArgs, v elementT) {
	count := -(t.count + 1)
	t.root = t.MutableRoot().Add(args, v, 0, newHasher(v, 0), &count)
	t.count = -count
}

func (t Tree) Count() int {
	return t.count
}

func (t Tree) Gauge() depth.Gauge {
	return depth.NewGauge(t.count)
}

func (t Tree) String() string {
	a := t.Root()
	return a.String()
}

func (t Tree) Combine(args *CombineArgs, u Tree) Tree {
	count := -(t.count + u.count)
	a := t.Root()
	b := u.Root()
	return newTreeNeg(a.Combine(args, b, 0, &count), &count)
}

func (t Tree) Difference(args *EqArgs, u Tree) Tree {
	count := -t.count
	a := t.Root()
	b := u.Root()
	return newTreeNeg(a.Difference(args, b, 0, &count), &count)
}

func (t Tree) Equal(args *EqArgs, u Tree) bool {
	a := t.Root()
	b := u.Root()
	return a.Equal(args, b, 0)
}

func (t Tree) Get(args *EqArgs, v elementT) *elementT {
	a := t.Root()
	h := newHasher(v, 0)
	return a.Get(args, v, h)
}

func (t Tree) Intersection(args *EqArgs, u Tree) Tree {
	if t.count > u.count {
		t, u = u, t
		args = args.flip
	}
	count := 0
	a := t.Root()
	b := u.Root()
	return newTree(a.Intersection(args, b, 0, &count), &count)
}

func (t Tree) Iterator() Iterator {
	a := t.Root()
	buf := packedIteratorBuf(t.count)
	return a.Iterator(buf)
}

func (t Tree) OrderedIterator(less Less, n int) Iterator {
	if n == -1 {
		n = t.count
	}
	o := &ordered{less: less, elements: make([]elementT, 0, n)}
	a := t.MutableRoot()
	for i := a.Iterator(packedIteratorBuf(t.count)); i.Next(); {
		heap.Push(o, i.Value())
		if o.Len() > n {
			heap.Pop(o)
		}
	}
	r := reverseO(o)
	heap.Init(r)
	return r.(Iterator)
}

func (t *Tree) Remove(args *EqArgs, v elementT) {
	count := -t.count
	a := t.MutableRoot()
	h := newHasher(v, 0)
	t.root = a.Remove(args, v, 0, h, &count)
	t.count = -count
}

func (t Tree) SubsetOf(args *EqArgs, u Tree) bool {
	a := t.Root()
	return a.SubsetOf(args, u.Root(), 0)
}

func (t Tree) Map(args *CombineArgs, f func(v elementT) elementT) Tree {
	count := 0
	a := t.Root()
	return newTree(a.Map(args, 0, &count, f), &count)
}

func (t Tree) Reduce(args NodeArgs, r func(values ...elementT) elementT) elementT {
	a := t.Root()
	return a.Reduce(args, 0, r)
}

func (t Tree) Where(args *WhereArgs) Tree {
	count := 0
	a := t.Root()
	return newTree(a.Where(args, 0, &count), &count)
}

func (t Tree) With(args *CombineArgs, v elementT) Tree {
	count := -(t.count + 1)
	a := t.Root()
	h := newHasher(v, 0)
	return newTreeNeg(a.With(args, v, 0, h, &count), &count)
}

func (t Tree) Without(args *EqArgs, v elementT) Tree {
	count := -t.count
	a := t.Root()
	h := newHasher(v, 0)
	return newTreeNeg(a.Without(args, v, 0, h, &count), &count)
}

func packedIteratorBuf(count int) [][]noderef {
	depth := (bits.Len64(uint64(count)) + 1) * 3 / 2 // 1.5 (log₈(count) + 1)
	return make([][]noderef, 0, depth)
}
