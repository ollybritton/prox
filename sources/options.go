package sources

import (
	"fmt"
	"time"
)

// Options represents a set of options that could be passed when getting a proxy.
type Options struct {
	Type      string        // Either "HTTP", "HTTPS", "SOCKS4" or "SOCKS5".
	Country   string        // Alpha2 code or "all".
	Anonymity string        // "elite", "anonymous", "transparent" or "all". Only applicable to HTTP proxies.
	Timeout   time.Duration // Maximum amount of time to wait for proxies.
}

// OptionsErr is returned when there's an issue with the options specified.
type OptionsErr struct {
	options Options
	msg     string
}

// Error returns the corresponding error message for an OptionsErr.
func (oe OptionsErr) Error() string {
	return fmt.Sprintf("invalid options %+v: %s", oe.options, oe.msg)
}

// NewOptionsErr returns a new options error.
func NewOptionsErr(options Options, msg string) OptionsErr {
	return OptionsErr{options: options, msg: msg}
}
