package components

import (
	"fmt"
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"sync/atomic"
)

type ScrollableBox struct {
	twin.Box
	vOffset atomic.Value // twin.Point
	vSize   atomic.Value // twin.Size
	sbs     ScrollableBoxStyle
}

type ScrollableBoxStyle struct {
	win         string
	style       *tcell.Style
	activeStyle *tcell.Style
	flags       *WindowFlags
}

func (sbs ScrollableBoxStyle) WithWin(win string) ScrollableBoxStyle {
	sbs.win = win
	return sbs
}

func (sbs ScrollableBoxStyle) WithStyle(style tcell.Style) ScrollableBoxStyle {
	sbs.style = &style
	return sbs
}

func (sbs ScrollableBoxStyle) Style() tcell.Style {
	if sbs.style != nil {
		return *sbs.style
	}
	ws := GetThemeValue[WindowTheme](sbs.win)
	return ws.NotActive
}

func (sbs ScrollableBoxStyle) WithActiveStyle(activeStyle tcell.Style) ScrollableBoxStyle {
	sbs.activeStyle = &activeStyle
	return sbs
}

func (sbs ScrollableBoxStyle) ActiveStyle() tcell.Style {
	if sbs.flags != nil {
		return *sbs.activeStyle
	}
	ws := GetThemeValue[WindowTheme](sbs.win)
	return ws.Active
}

func (sbs ScrollableBoxStyle) WithFlags(flags WindowFlags) ScrollableBoxStyle {
	sbs.flags = &flags
	return sbs
}

func (sbs ScrollableBoxStyle) Flags() WindowFlags {
	if sbs.flags != nil {
		return *sbs.flags
	}
	ws := GetThemeValue[WindowTheme](sbs.win)
	return ws.Flags
}

func (sbs ScrollableBoxStyle) windowStyle() WindowTheme {
	return GetThemeValue[WindowTheme](sbs.win)
}

func (sb *ScrollableBox) Init(owner, this twin.Component, sbs ScrollableBoxStyle) error {
	if sbs.win == "" {
		sbs.win = "win"
	}
	sb.sbs = sbs
	sb.vSize.Store(twin.Size{})
	sb.vOffset.Store(twin.Point{})
	return sb.Box.Init(owner, this)
}

func (sb *ScrollableBox) OnDraw(cc *twin.CanvasContext) {
	b := sb.Bounds().Normalized()
	stl := sb.sbs.Style()
	flags := sb.sbs.Flags()
	cc.FilledRectangle(b, stl)
	if flags&WindowFlagHasBorderBM != 0 {
		var crs twin.CanvasRectangleStyle
		if twin.IsActive(sb) {
			crs = sb.sbs.windowStyle().ActiveRectStyle
		} else {
			crs = sb.sbs.windowStyle().NaRectStyle
		}
		cc.Rectangle(b, crs, stl)
	}
	sbSize := b.Size()
	sbPoint := twin.Point{X: 0, Y: 0}
	hasV := sb.HasVBar()
	hasH := sb.HasHBar()

	if flags&WindowFlagHasBorderBM != 0 {
		sbSize = sbSize.Add(-2, -2)
		sbPoint = twin.Point{X: 1, Y: 1}
	} else {
		if hasV {
			sbSize = sbSize.Add(-1, 0)
		}
		if hasH {
			sbSize = sbSize.Add(0, -1)
		}
	}
	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))
	if hasV {
		cc.DrawVScrollBar(twin.Point{X: b.Width - 1, Y: sbPoint.Y}, sbSize.Height, virtSize.Height, visibleSize.Height, virtOffset.Y, stl)
	}
	if hasH {
		cc.DrawHScrollBar(twin.Point{X: sbPoint.X, Y: b.Height - 1}, sbSize.Width, virtSize.Width, visibleSize.Width, virtOffset.X, stl)
	}
}

func (sb *ScrollableBox) ChildrenCanvasBounds() twin.Rectangle {
	b := sb.Bounds()
	flags := sb.sbs.Flags()
	if flags&WindowFlagHasBorderBM != 0 {
		b.X, b.Y = b.X+1, b.Y+1
		b.Width = max(0, b.Width-2)
		b.Height = max(0, b.Height-2)
		return b
	}
	if sb.HasVBar() {
		b.Width = max(0, b.Width-1)
	}
	if sb.HasHBar() {
		b.Height = max(0, b.Height-1)
	}
	return b
}

func (sb *ScrollableBox) OnMousePressed(p twin.Point) bool {
	b := sb.Bounds().Normalized()
	flags := sb.sbs.Flags()
	sbSize := b.Size()
	sbPoint := twin.Point{X: 0, Y: 0}
	hasV := sb.HasVBar()
	hasH := sb.HasHBar()

	if flags&WindowFlagHasBorderBM != 0 {
		sbSize = sbSize.Add(-2, -2)
		sbPoint = twin.Point{X: 1, Y: 1}
	} else {
		if hasV {
			sbSize = sbSize.Add(-1, 0)
		}
		if hasH {
			sbSize = sbSize.Add(0, -1)
		}
	}

	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))

	fmt.Println(hasV, p, b, sbSize)
	if hasV && p.X == b.Width-1 && p.Y < sbPoint.Y+sbSize.Height && p.Y >= sbPoint.Y {
		maxY := virtSize.Height - visibleSize.Height
		if p.Y == sbPoint.Y {
			virtOffset.Y = virtOffset.Y - 1
		} else if p.Y == sbPoint.Y+sbSize.Height-1 {
			virtOffset.Y = virtOffset.Y + 1
		} else if sbSize.Height > 2 {
			virtOffset.Y = int(float64(p.Y-1) / float64(sbSize.Height-2) * float64(maxY))
		}
		virtOffset.Y = max(0, min(virtOffset.Y, maxY))
		sb.SetVirtualOffset(virtOffset)
		twin.Redraw(twin.This(sb))
		return true
	}

	if hasH && p.Y == b.Height-1 && p.X < sbPoint.X+sbSize.Width && p.X >= sbPoint.X {
		maxX := virtSize.Width - visibleSize.Width
		if p.X == sbPoint.X {
			virtOffset.X = virtOffset.X - 1
		} else if p.X == sbPoint.X+sbSize.Width-1 {
			virtOffset.X = virtOffset.X + 1
		} else if sbSize.Width > 2 {
			virtOffset.X = int(float64(p.X-1) / float64(sbSize.Width-2) * float64(maxX))
		}
		virtOffset.X = max(0, min(virtOffset.X, maxX))
		sb.SetVirtualOffset(virtOffset)
		twin.Redraw(twin.This(sb))
		return true
	}
	return false
}

func (sb *ScrollableBox) OnKeyPressed(ke *tcell.EventKey) bool {
	visibleSize := sb.ChildrenCanvasBounds().Size()
	switch ke.Key() {
	case tcell.KeyLeft:
		return sb.scroll(twin.Point{X: -1, Y: 0})
	case tcell.KeyRight:
		return sb.scroll(twin.Point{X: +1, Y: 0})
	case tcell.KeyUp:
		return sb.scroll(twin.Point{X: 0, Y: -1})
	case tcell.KeyDown:
		return sb.scroll(twin.Point{X: 0, Y: +1})
	case tcell.KeyPgUp:
		return sb.scroll(twin.Point{X: 0, Y: -visibleSize.Height})
	case tcell.KeyPgDn:
		return sb.scroll(twin.Point{X: 0, Y: visibleSize.Height})
	case tcell.KeyHome:
		return sb.scroll(twin.Point{X: -visibleSize.Width, Y: 0})
	case tcell.KeyEnd:
		return sb.scroll(twin.Point{X: +visibleSize.Width, Y: 0})
	}
	return sb.Box.OnKeyPressed(ke)
}

func (sb *ScrollableBox) OnMouseWheel(p twin.Point, wheel twin.MouseWheel) bool {
	switch wheel {
	case twin.MouseWheelUp:
		return sb.scroll(twin.Point{X: 0, Y: -1})
	case twin.MouseWheelDown:
		return sb.scroll(twin.Point{X: 0, Y: 1})
	case twin.MouseWheelLeft:
		return sb.scroll(twin.Point{X: -1, Y: 0})
	case twin.MouseWheelRight:
		return sb.scroll(twin.Point{X: 1, Y: 0})
	}
	return false
}

func (sb *ScrollableBox) scroll(p twin.Point) bool {
	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	maxVOX, maxVOY := virtSize.Width-visibleSize.Width, virtSize.Height-visibleSize.Height
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))
	if p.Y != 0 {
		virtOffset.Y = max(0, min(maxVOY, virtOffset.Y+p.Y))
	}
	if p.X != 0 {
		virtOffset.X = max(0, min(maxVOX, virtOffset.X+p.X))
	}
	if p.Y != 0 || p.X != 0 {
		sb.SetVirtualOffset(virtOffset)
		twin.Redraw(twin.This(sb))
		return true
	}
	return false
}

func (sb *ScrollableBox) VirtualSize() twin.Size {
	return sb.vSize.Load().(twin.Size)
}

func (sb *ScrollableBox) SetVirtualSize(size twin.Size) {
	sb.vSize.Store(size)
}

func (sb *ScrollableBox) VirtualOffset() twin.Point {
	return sb.vOffset.Load().(twin.Point)
}

func (sb *ScrollableBox) SetVirtualOffset(offset twin.Point) {
	sb.vOffset.Store(offset)
}

func (sb *ScrollableBox) CanBeFocused() bool { return true }

func (sb *ScrollableBox) Style() tcell.Style { return sb.sbs.Style() }

func (sb *ScrollableBox) HasBorder() bool {
	return sb.sbs.Flags()&WindowFlagHasBorderBM != 0
}

func (sb *ScrollableBox) HasVBar() bool {
	flags := sb.sbs.Flags()
	noBar := flags&WindowFlagHasVerticalScrollBM == 0
	if flags&WindowFlagAutoHideScrollBM == 0 || noBar {
		return !noBar
	}
	s := sb.Bounds().Size()
	vs := sb.VirtualSize()
	if vs.Height > s.Height {
		return true
	}
	if flags&WindowFlagHasBorderBM != 0 {
		// border has the bars or not, doesn't change visible size anymore
		return vs.Height > s.Height-2
	}
	if vs.Width > s.Width {
		// we definitely have horz bar
		return vs.Height == s.Height
	}
	// vs.Width <= s.Width && vs.Height <= s.Height
	return false
}

func (sb *ScrollableBox) HasHBar() bool {
	flags := sb.sbs.Flags()
	noBar := flags&WindowFlagHasHorizontalScrollBM == 0
	if flags&WindowFlagAutoHideScrollBM == 0 || noBar {
		return !noBar
	}
	s := sb.Bounds().Size()
	vs := sb.VirtualSize()
	if vs.Width > s.Width {
		return true
	}
	if flags&WindowFlagHasBorderBM != 0 {
		return vs.Width > s.Width-2
	}
	if vs.Height > s.Height {
		// we definitely have the vert bar
		return vs.Width == s.Width
	}
	// vs.Width <= s.Width && vs.Height <= s.Height
	return false
}
