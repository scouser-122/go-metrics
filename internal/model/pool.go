package models

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool    sync.Pool
	newFunc func() T
}

func NewPool[T Resettable](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
		newFunc: newFunc,
	}
}

func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	if obj == nil {
		return p.newFunc()
	}
	return obj.(T)
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}
