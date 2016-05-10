// This is a dumbed down version of the Tomb package by Gustavo Niemeyer.
//
// Copyright (c) 2011 - Gustavo Niemeyer <gustavo@niemeyer.net>
//
// All rights reserved.
//
// Redistribution and use in source and binary forms, with or without
// modification, are permitted provided that the following conditions are met:
//
//     * Redistributions of source code must retain the above copyright notice,
//       this list of conditions and the following disclaimer.
//     * Redistributions in binary form must reproduce the above copyright notice,
//       this list of conditions and the following disclaimer in the documentation
//       and/or other materials provided with the distribution.
//     * Neither the name of the copyright holder nor the names of its
//       contributors may be used to endorse or promote products derived from
//       this software without specific prior written permission.
//
// THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS
// "AS IS" AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT
// LIMITED TO, THE IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR
// A PARTICULAR PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT OWNER OR
// CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
// EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
// PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
// PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
// LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
// NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
// SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.

package sdbot

import (
	"sync"
)

// Killable is a struct used to control the termination of goroutines.
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

// Dying tells us if the Killable has been requested to terminate.
func (k *Killable) Dying() <-chan struct{} {
	k.init()
	return k.dying
}

// Wait waits until the goroutine has been terminated.
func (k *Killable) Wait() {
	k.init()
	<-k.dead
}

// Done requests termination and closes the dead channel.
func (k *Killable) Done() {
	k.Kill()
	close(k.dead)
}

// Kill requests termination but does not close the dead channel.
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
