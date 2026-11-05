// Package design bundles the canvas page: one HTML file with its styles and its
// script inline, compiled into the binary.
//
// No build step, no CDN, no remote asset -the same rule the artboards follow. A
// canvas that needs the network to draw a local file breaks on a plane.
package design

import _ "embed"

//go:embed canvas.html
var canvas []byte

//go:embed plugin.ts
var plugin []byte

// Canvas returns the page the server serves at /.
func Canvas() []byte { return canvas }

// Plugin returns the OpenCode plugin that delivers the canvas queue to the
// session: one file per machine, watched per project.
func Plugin() []byte { return plugin }
