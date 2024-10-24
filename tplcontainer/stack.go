package tplcontainer

import "container/list"

type Stack[T any] struct {
	list *list.List
}

func NewStack[T any]() *Stack[T] {
	return &Stack[T]{
		list: list.New(),
	}
}

func (s *Stack[T]) Push(e T) {
	s.list.PushBack(e)
}

func (s *Stack[T]) Pop() (T, bool) {
	if back := s.list.Back(); back != nil {
		return s.list.Remove(back).(T), true
	}
	var zero T
	return zero, false
}

func (s *Stack[T]) Top() (T, bool) {
	if back := s.list.Back(); back != nil {
		return back.Value.(T), true
	}
	var zero T
	return zero, false
}

func (s *Stack[T]) Len() int {
	return s.list.Len()
}

func (s *Stack[T]) BottomIterator() StackIterator[T] {
	return StackIterator[T]{e: s.list.Front()}
}

type StackIterator[T any] struct {
	e *list.Element
}

func (it StackIterator[T]) Next() StackIterator[T] {
	if !it.Valid() {
		var zero StackIterator[T]
		return zero
	}

	return StackIterator[T]{e: it.e.Next()}
}

func (it StackIterator[T]) Valid() bool {
	return it.e != nil
}

func (it StackIterator[T]) Value() T {
	return it.e.Value.(T)
}
