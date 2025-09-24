package main

import "github.com/dspasibenko/twin-go/twin"

type (
	// Main panel contains the status bar (footer), ArgusDevicePanel and ArgusRegistersPanel
	Main struct {
		twin.BaseContainer
	}

	// ArgusDevicePanel contains different implementations of the Argus devices
	ArgusDevicePanel struct {
		twin.BaseContainer
	}

	// ArgusRegistersPanel contains different device registers views
	ArgusRegistersPanel struct {
		twin.BaseContainer
	}
)
