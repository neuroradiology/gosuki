//
//  Copyright (c) 2025 Chakib Ben Ziane <contact@blob42.xyz>  and [`gosuki` contributors](https://github.com/blob42/gosuki/graphs/contributors).
//  All rights reserved.
//
//  SPDX-License-Identifier: AGPL-3.0-or-later
//
//  This file is part of GoSuki.
//
//  GoSuki is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Affero General Public License as
//  published by the Free Software Foundation, either version 3 of the
//  License, or (at your option) any later version.
//
//  GoSuki is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Affero General Public License for more details.
//
//  You should have received a copy of the GNU Affero General Public License
//  along with gosuki.  If not, see <http://www.gnu.org/licenses/>.
//

package modules

// Implements a channel based inter module communication ()

import (
	"context"
	"fmt"
	"time"

	"github.com/blob42/gosuki/internal/database"
	"github.com/blob42/gosuki/pkg/manager"
)

const DispatcherID = "dispatcher"

var (
	MsgDispatcher = modMsgDispatcher{
		listeners: map[ModID]listener{},
	}
)

type ModMsgType int

// types of messages passed between modules
const (
	MsgTriggerSync = iota
	MsgHello
	MsgPanic

	MsgSyncPeers
)

// ModMsg is a message exchanged between modules
type ModMsg struct {
	Type    ModMsgType
	To      ModID
	Payload any
}

var ModMsgBus = make(chan ModMsg) // Channel for intra-process message passing

// MsgListener is a module that can listen to messages from other mods
type MsgListener interface {
	MsgListen(context.Context, <-chan ModMsg)
}

// Listener is a Work unit for modules that implement the MsgListener interface
type Listener struct {
	Ctx   context.Context
	Queue chan ModMsg
	MsgListener
}

func (lw Listener) Run(m manager.UnitManager) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				m.Panic(fmt.Errorf("%v", err))
			}
		}()
		lw.MsgListen(lw.Ctx, lw.Queue)
	}()

	<-m.ShouldStop()
	m.Done()
}

type listener struct {
	id    ModID
	queue chan<- ModMsg
}

// dispatchers messages between modules
type modMsgDispatcher struct {
	listeners map[ModID]listener
}

func (mm *modMsgDispatcher) AddListener(id ModID, queue chan<- ModMsg) {
	mm.listeners[id] = listener{id, queue}
}

func (mm *modMsgDispatcher) Run(m manager.UnitManager) {

	// interval at which we check if we need to trigger a sync
	checkSyncTicker := time.NewTicker(time.Second * 5)
	go func() {
		log.Debug("dispatching module messages")
		for {
			select {
			case msg := <-ModMsgBus:
				log.Debug("dispatching mod message", "msg", msg.Type, "to", msg.To)
				if dst, ok := mm.listeners[msg.To]; ok {
					log.Trace("sending", "msg", msg.Type, "to-mod", msg.To)
					dst.queue <- msg
				} else { // discard
					log.Debugf("target %s not available, discarding msg=%#v", msg.To, msg.Type)
				}
			case <-checkSyncTicker.C:
				trigger := database.SyncTrigger.Load()
				if dst, ok := mm.listeners["p2p-sync"]; ok && trigger {
					dst.queue <- ModMsg{MsgTriggerSync, "", nil}
				}
				database.SyncTrigger.Store(false)
			}
		}
	}()

	//DEBUG:
	// time.Sleep(5 * time.Second)
	// ModMsgBus <- ModMsg{MsgHello, "tui"}

	// Wait for stop signal
	<-m.ShouldStop()
	m.Done()
}

var _ manager.WorkUnit = (*modMsgDispatcher)(nil)
