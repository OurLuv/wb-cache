package service

import (
	"container/list"
	"sync"
	"time"
)

type ICache interface {
	Cap() int
	Len() int
	Clear() // удаляет все ключи
	Add(key, value interface{})
	AddWithTTL(key, value interface{}, ttl time.Duration) // добавляет ключ со сроком жизни ttl
	Get(key interface{}) (value interface{}, ok bool)
	Remove(key interface{})
}

type ICacheImpl struct {
	Capacity int
	LRU      *list.List
	Data     map[interface{}]*list.Element
	RWMutex  sync.RWMutex
}

type Item struct {
	key   interface{}
	value interface{}
}

func NewICache(cap int) ICache {
	return &ICacheImpl{
		Capacity: cap,
		LRU:      &list.List{},
		Data:     map[interface{}]*list.Element{},
		RWMutex:  sync.RWMutex{},
	}
}

func (c *ICacheImpl) Cap() int {
	return c.Capacity
}

func (c *ICacheImpl) Len() int {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	return len(c.Data)
}

func (c *ICacheImpl) Clear() {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	c.LRU.Init()
	c.Data = make(map[interface{}]*list.Element)
}

func (c *ICacheImpl) Add(key, value interface{}) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if elem, ok := c.Data[key]; ok {
		c.LRU.MoveToFront(elem)
		elem.Value.(*Item).value = value
		return
	}

	if c.Capacity <= c.LRU.Len() {
		elem := c.LRU.Back()
		if elem != nil {
			delete(c.Data, elem.Value.(*Item).key)
			c.LRU.Remove(elem)
		}
	}

	elem := c.LRU.PushFront(&Item{key: key, value: value})
	c.Data[key] = elem
}

func (c *ICacheImpl) AddWithTTL(key, value interface{}, ttl time.Duration) {
	c.Add(key, value)
	go func() {
		time.Sleep(ttl)
		c.Remove(key)
	}()

}

func (c *ICacheImpl) Get(key interface{}) (value interface{}, ok bool) {
	c.RWMutex.RLock()
	defer c.RWMutex.RUnlock()
	elem, ok := c.Data[key]
	if ok {
		c.LRU.MoveToFront(elem)
		return elem.Value.(*Item).value, ok
	} else {
		return nil, ok
	}
}

func (c *ICacheImpl) Remove(key interface{}) {
	c.RWMutex.Lock()
	defer c.RWMutex.Unlock()
	if elem, ok := c.Data[key]; ok {
		c.LRU.Remove(elem)
		delete(c.Data, key)
	}
}
