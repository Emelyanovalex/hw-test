package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type lruCache struct {
	capacity int
	queue    List
	items    map[Key]*ListItem
}

func NewCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		queue:    NewList(),
		items:    make(map[Key]*ListItem, capacity),
	}
}

const (
	Exist    = true
	notExist = false
)

func (l *lruCache) Set(key Key, value interface{}) bool {
	if item, ok := l.items[key]; ok {
		item.Value = value
		l.queue.MoveToFront(item)
		return Exist
	}

	q := l.queue.PushFront(value)
	l.items[key] = q

	if l.queue.Len() > l.capacity {
		l.removeLastCache()
	}

	return notExist
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	if item, ok := l.items[key]; ok {
		l.queue.MoveToFront(item)
		return item.Value, Exist
	}
	return nil, notExist
}

func (l *lruCache) Clear() {
	l.queue = NewList()
	l.items = make(map[Key]*ListItem, l.capacity)
}

func (l *lruCache) removeLastCache() {
	back := l.queue.Back()
	for k, v := range l.items {
		if v.Value == back.Value {
			delete(l.items, k)
			break
		}
	}
	l.queue.Remove(back)
}
