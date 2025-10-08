package components

import (
	"fmt"
	"github.com/dspasibenko/twin-go/pkg/golibs/container"
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"sync/atomic"
)

type TextAlignment int

const (
	AllignLeft = TextAlignment(iota)
	AllignRight
	AllignCenter
)

type ButtonType string

const (
	NormalButtonType ButtonType = "button"
)

type WindowFlags int

const (
	WindowFlagHasBorderBM           = WindowFlags(1)
	WindowFlagHasVerticalScrollBM   = WindowFlags(2)
	WindowFlagHasHorizontalScrollBM = WindowFlags(4)
	WindowFlagHasBothScrollsBM      = WindowFlags(6)
	WindowFlagAutoHideScrollBM      = WindowFlags(8)
)

func init() {
	SetTheme(GetDefaultTheme())
}

type Theme map[string]any

var theme atomic.Value

func SetTheme(t Theme) {
	t = container.CopyMap(t)
	theme.Store(t)
}

// GeThemeValue for the name. It panics if there is no such name in the theme
func GetThemeValue[T any](name string) T {
	t := theme.Load().(Theme)
	v, ok := t[name]
	if !ok {
		panic(fmt.Sprintf("Theme value %s not found", name))
	}
	return v.(T)
}

func GetThemeValueIfNoVal[T any](val *T, name string) T {
	if val == nil {
		return GetThemeValue[T](name)
	}
	return *val
}

type (
	ButtonTheme struct {
		NotActive tcell.Style
		Active    tcell.Style
		Alignment TextAlignment
	}

	LabelTheme struct {
		Style     tcell.Style
		Alignment TextAlignment
	}

	WindowTheme struct {
		NotActive       tcell.Style
		Active          tcell.Style
		Flags           WindowFlags
		NaRectStyle     twin.CanvasRectangleStyle
		ActiveRectStyle twin.CanvasRectangleStyle
	}

	ListBoxTheme struct {
		WindowTheme
		SelStyle  tcell.Style
		SelActive tcell.Style
	}
)

func GetDefaultTheme() Theme {
	return Theme{
		"win": WindowTheme{
			NotActive:       tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
			Active:          tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
			Flags:           WindowFlagHasBorderBM | WindowFlagHasBothScrollsBM | WindowFlagAutoHideScrollBM,
			NaRectStyle:     twin.CanvasRectangleRounded,
			ActiveRectStyle: twin.CanvasRectangleDouble,
		},
		"button": ButtonTheme{
			NotActive: tcell.StyleDefault.Background(tcell.ColorGrey).Foreground(tcell.ColorWhite),
			Active:    tcell.StyleDefault.Background(tcell.ColorWhite).Foreground(tcell.ColorBlack),
			Alignment: AllignCenter,
		},
		"label": LabelTheme{
			Style:     tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorYellow),
			Alignment: AllignLeft,
		},
		"listbox": ListBoxTheme{
			SelStyle:  tcell.StyleDefault.Background(tcell.ColorDarkBlue).Foreground(tcell.ColorGrey),
			SelActive: tcell.StyleDefault.Background(tcell.ColorBlack).Foreground(tcell.ColorWhite),
		},
		"listboxWin": WindowTheme{
			NotActive:       tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
			Active:          tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorWhite),
			Flags:           WindowFlagHasBorderBM | WindowFlagHasBothScrollsBM | WindowFlagAutoHideScrollBM,
			NaRectStyle:     twin.CanvasRectangleSingle,
			ActiveRectStyle: twin.CanvasRectangleDouble,
		},
	}
}
