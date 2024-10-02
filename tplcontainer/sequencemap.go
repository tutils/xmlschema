package tplcontainer

type SequenceMap[KEY comparable, VALUE any] struct {
	mp    map[KEY]VALUE
	order []KEY
}

func NewSequenceMap[KEY comparable, VALUE any]() *SequenceMap[KEY, VALUE] {
	return &SequenceMap[KEY, VALUE]{
		mp:    make(map[KEY]VALUE),
		order: make([]KEY, 0),
	}
}

func NewSequenceMapWithSizeHint[KEY comparable, VALUE any](sizeHint int) *SequenceMap[KEY, VALUE] {
	return &SequenceMap[KEY, VALUE]{
		mp:    make(map[KEY]VALUE, sizeHint),
		order: make([]KEY, 0, sizeHint),
	}
}

func (om *SequenceMap[KEY, VALUE]) Get(k KEY) (VALUE, bool) {
	v, ok := om.mp[k]
	return v, ok
}

func (om *SequenceMap[KEY, VALUE]) MustGet(k KEY) VALUE {
	v, ok := om.mp[k]
	if !ok {
		panic("invalid key")
	}
	return v
}

func (om *SequenceMap[KEY, VALUE]) Set(k KEY, v VALUE) {
	if _, ok := om.mp[k]; !ok {
		om.order = append(om.order, k)
	}
	om.mp[k] = v
}

func (om *SequenceMap[KEY, VALUE]) Len() int {
	return len(om.mp)
}

func (om *SequenceMap[KEY, VALUE]) Order() []KEY {
	return om.order
}
