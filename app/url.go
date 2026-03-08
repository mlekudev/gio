// SPDX-License-Identifier: Unlicense OR MIT

//go:build windows || android || darwin

package app

import (
	"net/url"

	"github.com/mlekudev/gio/io/event"
	"golang.org/x/net/idna"
)

func processGlobalEvent(evt event.Event) {
	if yieldGlobalEvent == nil {
		return
	}
	if !yieldGlobalEvent(evt) {
		yieldGlobalEvent = nil
	}
}

// newURLEvent creates a URLEvent from a raw URL string, handling Punycode decoding.
func newURLEvent(rawurl string) (URLEvent, error) {
	u, err := url.Parse(rawurl)
	if err != nil {
		return URLEvent{}, err
	}
	u.Host, err = idna.Punycode.ToUnicode(u.Hostname())
	if err != nil {
		return URLEvent{}, err
	}
	u, err = url.Parse(u.String())
	if err != nil {
		return URLEvent{}, err
	}
	return URLEvent{URL: u}, nil
}
