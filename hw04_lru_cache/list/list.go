// Package list provides implementation of linked list structure
package list

// List provides some contract for implementation.
type List interface {
	Len() int
	Front() *Item
	Back() *Item
	PushFront(v any) *Item
	PushBack(v any) *Item
	Remove(i *Item)
	MoveToFront(i *Item)
	// Debug() string
}

// Item is an item of list.
type Item struct {
	Value any
	Next  *Item
	Prev  *Item
}

type list struct {
	// Place your code here.
	head *Item
	tail *Item
	len  int
}

// NewList create a new list of linked list.
func NewList() List {
	l := new(list)
	// bc it will be nil anywayu
	// l.head.Prev = nil

	// l.head.Next = l.tail
	// l.tail.Prev = l.head
	return l
}

// Back return a pointer to the last element of List.
func (l *list) Back() *Item {
	return l.tail
}

// Front return a pointer to the first element of List.
func (l *list) Front() *Item {
	return l.head
}

// Len return the length of List.
func (l *list) Len() int {
	return l.len
}

// MoveToFront .
func (l *list) MoveToFront(i *Item) {
	if i == nil || i == l.head {
		return
	}

	if i == l.tail {
		l.tail = i.Prev
	} else {
		i.Next.Prev = i.Prev
	}

	i.Prev.Next = i.Next
	i.Prev = nil
	i.Next = l.head
	l.head.Prev = i
	l.head = i
}

// PushBack add element in the end of list.
func (l *list) PushBack(v any) *Item {
	data := &Item{
		Value: v,
	}
	if l.tail == nil || l.tail.Value == nil {
		// if we had only one element -> head and tail is the same
		l.tail = data
		l.head = data
	} else {
		l.tail.Next = data
		data.Prev = l.tail
		l.tail = data
	}
	l.len++
	return l.tail
}

// PushFront add `v` data in front of list.
func (l *list) PushFront(v any) *Item {
	data := &Item{
		Value: v,
	}
	// first -> let imagine that our list will be empty
	if l.head == nil || l.head.Value == nil {
		l.head = data
		// if we have only one element -> head and tail it's the same
		l.tail = data
	} else {
		data.Next = l.head
		l.head.Prev = data
		l.head = data
	}
	l.len++
	return l.head
}

// Remove the given (i) from List.
func (l *list) Remove(i *Item) {
	if i == nil {
		return
	}

	if i == l.head {
		l.head = i.Next
		if l.head != nil {
			l.head.Prev = nil
		} else {
			l.tail = nil
		}
		l.len--
		return
	}
	if i == l.tail {
		l.tail = i.Prev
		if l.tail != nil {
			l.tail.Next = nil
		}
		l.len--
		return
	}
	n := i.Next
	p := i.Prev
	n.Prev = p
	p.Next = n
	i.Next = nil
	i.Prev = nil
	l.len--
}
