package components

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
)

type EditLine struct {
	twin.Box
	els EditLineSettings
}

type EditLineSettings struct {
	style      tcell.Style
	activeStye tcell.Style
	curStyle   tcell.CursorStyle
}
