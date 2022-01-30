package tree

import (
	"fmt"

	"github.com/arr-ai/frozen/v2/internal/depth"
)

func DefaultNPKeyEqArgs[T comparable]() *EqArgs[T] {
	return NewDefaultKeyEqArgs[T](depth.NonParallel)
}

func DefaultNPKeyCombineArgs[T comparable]() *CombineArgs[T] {
	return NewCombineArgs(DefaultNPKeyEqArgs[T](), UseRHS[T])
}

func NewDefaultKeyEqArgs[T comparable](gauge depth.Gauge) *EqArgs[T] {
	return NewEqArgs[T](gauge, elementEqual[T], hashValue[T], hashValue[T])
}

// Builder[T] provides a more efficient way to build nodes incrementally.
type Builder[T comparable] struct {
	t Tree[T]
}

func NewBuilder[T comparable](capacity int) *Builder[T] {
	return &Builder[T]{}
}

func (b *Builder[T]) Count() int {
	return b.t.count
}

func (b *Builder[T]) Add(args *CombineArgs[T], v T) {
	if b.t.root == nil {
		b.t.root = newLeaf1(v)
		b.t.count = 1
	} else {
		h := newHasher(v, 0)
		if vetting {
			backup := b.clone()
			defer vet(func() { backup.Add(args, v) }, &b.t)(nil)
		}
		var matches int
		b.t.root, matches = b.t.root.Add(args, v, 0, h)
		b.t.count += 1 - matches
	}
}

func (b *Builder[T]) Remove(args *EqArgs[T], v T) {
	if b.t.root != nil {
		h := newHasher(v, 0)
		if vetting {
			backup := b.clone()
			defer vet(func() { backup.Remove(args, v) }, &b.t)(nil)
		}
		var matches int
		b.t.root, matches = b.t.root.Remove(args, v, 0, h)
		b.t.count -= matches
	}
}

func (b *Builder[T]) Get(args *EqArgs[T], el T) *T {
	return b.t.Get(args, el)
}

func (b *Builder[T]) Finish() Tree[T] {
	t := b.Borrow()
	*b = Builder[T]{}
	return t
}

func (b *Builder[T]) Borrow() Tree[T] {
	return b.t
}

func (b Builder[T]) String() string {
	return b.Borrow().String()
}

func (b Builder[T]) Format(state fmt.State, verb rune) {
	b.Borrow().Format(state, verb)
}

func (b *Builder[T]) clone() *Builder[T] {
	return &Builder[T]{t: b.t.clone()}
}