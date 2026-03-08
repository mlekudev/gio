// SPDX-License-Identifier: Unlicense OR MIT

//go:build android || (darwin && ios)

package app

type windowRendezvous struct {
	in      chan windowAndConfig
	out     chan windowAndConfig
	windows chan struct{}
}

type windowAndConfig struct {
	window  *callbacks
	options []Option
}

func newWindowRendezvous() *windowRendezvous {
	wr := &windowRendezvous{
		in:      make(chan windowAndConfig),
		out:     make(chan windowAndConfig),
		windows: make(chan struct{}),
	}
	go func() {
		in := wr.in
		var window windowAndConfig
		var out chan windowAndConfig
		for {
			select {
			case w := <-in:
				window = w
				out = wr.out
			case out <- window:
			}
		}
	}()
	return wr
}
