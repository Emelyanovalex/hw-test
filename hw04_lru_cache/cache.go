package hw04lrucache

type Key string

type Cache interface {
	Set(key Key, value interface{}) bool
	Get(key Key) (interface{}, bool)
	Clear()
}

type CacheItem struct {
	Key   Key
	Cache interface{}
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
	if item, exist := l.items[key]; exist {
		item.Value.(*CacheItem).Cache = value
		l.queue.MoveToFront(item)
		return Exist
	}

	item := l.queue.PushFront(&CacheItem{
		Key:   key,
		Cache: value,
	})
	l.items[key] = item

	if l.queue.Len() > l.capacity {
		l.removeLastCache()
	}

	return notExist
}

func (l *lruCache) Get(key Key) (interface{}, bool) {
	if item, exist := l.items[key]; exist {
		l.queue.MoveToFront(item)
		return item.Value.(*CacheItem).Cache, true
	}
	return nil, notExist
}

func (l *lruCache) Clear() {
	l.queue = NewList()
	l.items = make(map[Key]*ListItem, l.capacity)
}

func (l *lruCache) removeLastCache() {
	back := l.queue.Back()
	if back != nil {
		if ci, ok := back.Value.(*CacheItem); ok {
			delete(l.items, ci.Key)
			l.queue.Remove(back)
		}
	}
}
