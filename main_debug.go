//go:build debug

package main

import (
	"NordVPNGUI/nordvpn"
)

func main() {
	nordvpn.Init()
	open(true)
}
