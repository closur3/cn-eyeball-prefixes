package iputil

type SpanRecord interface{ GetLo() uint32; GetHi() uint32 }

type SpanHeap[T SpanRecord] struct {
	items []int
	Data  []T
}

func (h SpanHeap[T]) Len() int { return len(h.items) }
func (h SpanHeap[T]) Less(i, j int) bool {
	a, b := h.Data[h.items[i]], h.Data[h.items[j]]
	as, bs := uint64(a.GetHi())-uint64(a.GetLo()), uint64(b.GetHi())-uint64(b.GetLo())
	if as != bs {
		return as < bs
	}
	return h.items[i] < h.items[j]
}
func (h SpanHeap[T]) Swap(i, j int)      { h.items[i], h.items[j] = h.items[j], h.items[i] }
func (h *SpanHeap[T]) Push(v any)        { h.items = append(h.items, v.(int)) }
func (h *SpanHeap[T]) Pop() any          { x := h.items[len(h.items)-1]; h.items = h.items[:len(h.items)-1]; return x }
func (h *SpanHeap[T]) Top() int          { return h.items[0] }
