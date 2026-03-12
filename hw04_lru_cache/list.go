package hw04lrucache

type List interface {
	Len() int
	Front() *ListItem
	Back() *ListItem
	PushFront(v interface{}) *ListItem
	PushBack(v interface{}) *ListItem
	Remove(i *ListItem)
	MoveToFront(i *ListItem)
}

type ListItem struct {
	Value interface{}
	Next  *ListItem
	Prev  *ListItem
}

type list struct {
	len   int
	front *ListItem
	back  *ListItem
}

func NewList() List { return new(list) }

func (l *list) Len() int { return l.len }

func (l *list) Front() *ListItem { return l.front }

func (l *list) Back() *ListItem { return l.back }

func (l *list) PushFront(v interface{}) *ListItem {
	e, ok := l.insertIfEmpty(v)
	if ok {
		return e
	}

	l.front.Prev = &ListItem{v, l.front, nil}
	l.front = l.front.Prev
	l.len++
	return l.front
}

func (l *list) PushBack(v interface{}) *ListItem {
	e, ok := l.insertIfEmpty(v)
	if ok {
		return e
	}

	l.back.Next = &ListItem{v, nil, l.back}
	l.back = l.back.Next
	l.len++
	return l.back
}

func (l *list) Remove(i *ListItem) {
	l.swapper(i)
	l.len--
}

func (l *list) MoveToFront(i *ListItem) {
	l.swapper(i)
	i.Next = l.front
	i.Prev = nil
	l.front.Prev = i
	l.front = i
}

func (l *list) insertIfEmpty(v interface{}) (*ListItem, bool) {
	if l.len == 0 {
		node := &ListItem{Value: v, Next: nil, Prev: nil}
		l.front = node
		l.back = node
		l.len++
		return node, true
	}
	return nil, false
}

func (l *list) swapper(i *ListItem) {
	if i.Prev != nil {
		i.Prev.Next = i.Next
	} else {
		l.front = i.Next
	}

	if i.Next != nil {
		i.Next.Prev = i.Prev
	} else {
		l.back = i.Prev
	}
}
