package frozen

// IntLess dictates the order of two elements.
type IntLess func(a, b int) bool

type IntIterator interface {
	Next() bool
	Value() int
}

type intSetIterator struct {
	blockIter      *MapIterator
	block          []cellMask
	cell           cellMask
	firstIntInCell int
}

func (i *intSetIterator) Next() bool {
	if len(i.block) == 0 {
		if !i.blockIter.Next() {
			return false
		}
		i.firstIntInCell = i.blockIter.Key().(int) * blockBits
		block := i.blockIter.Value().(cellBlock)
		for i.block = block[:]; i.block[0] == 0; i.block = i.block[1:] {
			i.firstIntInCell += cellBits
		}
	} else if i.block[0] == 0 {
		for ; i.block[0] == 0; i.block = i.block[1:] {
			i.firstIntInCell += cellBits
		}
	} else {
		i.block[0] &= i.block[0] - 1
	}
	return true
}

func (i *intSetIterator) Value() int {
	return i.firstIntInCell + BitIterator(i.block[0]).Index()
}
