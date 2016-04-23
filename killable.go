package sdbot

import (
	"sync"
)

type Killable struct {
	mutex sync.Mutex
	dying chan struct{}
	dead  chan struct{}
}

func (k *Killable) init() {
	k.mutex.Lock()
	if k.dead == nil {
		k.dying = make(chan struct{})
		k.dead = make(chan struct{})
	}
	k.mutex.Unlock()
}

func (k *Killable) Dying() <-chan struct{} {
	k.init()
	return k.dying
}

func (k *Killable) Wait() {
	k.init()
	<-k.dead
}

func (k *Killable) Done() {
	k.Kill()
	close(k.dead)
}

func (k *Killable) Kill() {
	k.init()
	k.mutex.Lock()
	defer k.mutex.Unlock()
	select {
	case <-k.dying:
	default:
		close(k.dying)
	}
}
