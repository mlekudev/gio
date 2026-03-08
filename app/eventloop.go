// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows || android || darwin

package app

import (
	"github.com/mlekudev/gio/io/event"
	"github.com/mlekudev/gio/op"
)

// eventLoop implements the functionality required for drivers where
// window event loops must run on a separate thread.
type eventLoop struct {
	win *callbacks
	// wakeup is the callback to wake up the event loop.
	wakeup func()
	// driverFuncs is a channel of functions to run the next
	// time the window loop waits for events.
	driverFuncs chan func()
	// invalidates is notified when an invalidate is requested by the client.
	invalidates chan struct{}
	// immediateInvalidates is an optimistic invalidates that doesn't require a wakeup.
	immediateInvalidates chan struct{}
	// events is where the platform backend delivers events bound for the
	// user program.
	events   chan event.Event
	frames   chan *op.Ops
	frameAck chan struct{}
	// delivering avoids re-entrant event delivery.
	delivering bool
}

// flushEvent is sent to detect when the user program
// has completed processing of all prior events. Its an
// [io/event.Event] but only for internal use.
type flushEvent struct{}

func (t flushEvent) ImplementsEvent() {}

// theFlushEvent avoids allocating garbage when sending
// flushEvents.
var theFlushEvent flushEvent

func newEventLoop(w *callbacks, wakeup func()) *eventLoop {
	return &eventLoop{
		win:                  w,
		wakeup:               wakeup,
		events:               make(chan event.Event),
		invalidates:          make(chan struct{}, 1),
		immediateInvalidates: make(chan struct{}),
		frames:               make(chan *op.Ops),
		frameAck:             make(chan struct{}),
		driverFuncs:          make(chan func(), 1),
	}
}

// Frame receives a frame and waits for its processing. It is called by
// the client goroutine.
func (e *eventLoop) Frame(frame *op.Ops) {
	e.frames <- frame
	<-e.frameAck
}

// Event returns the next available event. It is called by the client
// goroutine.
func (e *eventLoop) Event() event.Event {
	for {
		evt := <-e.events
		// Receiving a flushEvent indicates to the platform backend that
		// all previous events have been processed by the user program.
		if _, ok := evt.(flushEvent); ok {
			continue
		}
		return evt
	}
}

// Invalidate requests invalidation of the window. It is called by the client
// goroutine.
func (e *eventLoop) Invalidate() {
	select {
	case e.immediateInvalidates <- struct{}{}:
		// The event loop was waiting, no need for a wakeup.
	case e.invalidates <- struct{}{}:
		// The event loop is sleeping, wake it up.
		e.wakeup()
	default:
		// A redraw is pending.
	}
}

// Run f in the window loop thread. It is called by the client goroutine.
func (e *eventLoop) Run(f func()) {
	e.driverFuncs <- f
	e.wakeup()
}

// FlushEvents delivers pending events to the client.
func (e *eventLoop) FlushEvents() {
	if e.delivering {
		return
	}
	e.delivering = true
	defer func() { e.delivering = false }()
	for {
		evt, ok := e.win.nextEvent()
		if !ok {
			break
		}
		e.deliverEvent(evt)
	}
}

func (e *eventLoop) deliverEvent(evt event.Event) {
	var frames <-chan *op.Ops
	for {
		select {
		case f := <-e.driverFuncs:
			f()
		case frame := <-frames:
			// The client called FrameEvent.Frame.
			frames = nil
			e.win.ProcessFrame(frame, e.frameAck)
		case e.events <- evt:
			switch evt.(type) {
			case flushEvent, DestroyEvent:
				// DestroyEvents are not flushed.
				return
			case FrameEvent:
				frames = e.frames
			}
			evt = theFlushEvent
		case <-e.invalidates:
			e.win.Invalidate()
		case <-e.immediateInvalidates:
			e.win.Invalidate()
		}
	}
}

func (e *eventLoop) Wakeup() {
	for {
		select {
		case f := <-e.driverFuncs:
			f()
		case <-e.invalidates:
			e.win.Invalidate()
		case <-e.immediateInvalidates:
			e.win.Invalidate()
		default:
			return
		}
	}
}
