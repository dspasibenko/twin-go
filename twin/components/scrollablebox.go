package components

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"sync/atomic"
)

type ScrollableBox struct {
	twin.Box
	vOffset atomic.Value // twin.Point
	vSize   atomic.Value // twin.Size
	flags   int
	style   tcell.Style
}

const (
	ScrollableBoxHasBorderBM           = 1
	ScrollableBoxHasVerticalScrollBM   = 2
	ScrollableBoxHasHorizontalScrollBM = 4
	ScrollableBoxHasBothScrollsBM      = 6
	ScrollableBoxAutoHide              = 8
)

func (sb *ScrollableBox) Init(owner, this twin.Component, style tcell.Style, flags int) error {
	sb.flags = flags
	sb.style = style
	sb.vSize.Store(twin.Size{})
	sb.vOffset.Store(twin.Point{})
	return sb.Box.Init(owner, this)
}

func (sb *ScrollableBox) OnDraw(cc *twin.CanvasContext) {
	b := sb.Bounds().Normalized()
	cc.FilledRectangle(b, sb.style)
	if sb.flags&ScrollableBoxHasBorderBM != 0 {
		cc.Rectangle(b, twin.IsActive(sb), sb.style)
	}
	sbSize := b.Size()
	hasV := sb.hasVBar()
	hasH := sb.hasHBar()
	if hasV {
		sbSize = sbSize.Add(-1, 0)
	}
	if hasH {
		sbSize = sbSize.Add(0, -1)
	}
	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))
	if hasV {
		cc.DrawVScrollBar(b.TopRight(), sbSize.Height, virtSize.Height, visibleSize.Height, virtOffset.Y, sb.style)
	}
	if hasH {
		cc.DrawHScrollBar(b.BottomLeft(), sbSize.Width, virtSize.Width, visibleSize.Width, virtOffset.X, sb.style)
	}
}

func (sb *ScrollableBox) ChildrenCanvasBounds() twin.Rectangle {
	b := sb.Bounds()
	if sb.flags&ScrollableBoxHasBorderBM != 0 {
		b.X, b.Y = b.X+1, b.Y+1
		b.Width = max(0, b.Width-2)
		b.Height = max(0, b.Height-2)
		return b
	}
	if sb.hasVBar() {
		b.Width = max(0, b.Width-1)
	}
	if sb.hasHBar() {
		b.Height = max(0, b.Height-1)
	}
	return b
}

func (sb *ScrollableBox) OnMousePressed(p twin.Point) bool {
	b := sb.Bounds().Normalized()
	sbSize := b.Size()
	hasV := sb.hasVBar()
	hasH := sb.hasHBar()
	if hasV {
		sbSize = sbSize.Add(-1, 0)
	}
	if hasH {
		sbSize = sbSize.Add(0, -1)
	}
	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))

	if hasV && p.X == sbSize.Width && p.Y < sbSize.Height {
		maxY := virtSize.Height - visibleSize.Height
		if p.Y == 0 {
			virtOffset.Y = virtOffset.Y - 1
		} else if p.Y == sbSize.Height-1 {
			virtOffset.Y = virtOffset.Y + 1
		} else if sbSize.Height > 2 {
			virtOffset.Y = int(float64(p.Y-1) / float64(sbSize.Height-2) * float64(maxY))
		}
		virtOffset.Y = max(0, min(virtOffset.Y, maxY))
		sb.SetVirtualOffset(virtOffset)
		twin.Redraw(twin.This(sb))
		return true
	}

	if hasH && p.Y == sbSize.Height && p.X < sbSize.Width {
		maxX := virtSize.Width - visibleSize.Width
		if p.X == 0 {
			virtOffset.X = virtOffset.X - 1
		} else if p.X == sbSize.Width-1 {
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

func (sb *ScrollableBox) hasVBar() bool {
	noBar := sb.flags&ScrollableBoxHasVerticalScrollBM == 0
	if sb.flags&ScrollableBoxAutoHide == 0 || noBar {
		return !noBar
	}
	s := sb.Bounds().Size()
	vs := sb.VirtualSize()
	if vs.Height > s.Height {
		return true
	}
	if sb.flags&ScrollableBoxHasBorderBM != 0 {
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

func (sb *ScrollableBox) hasHBar() bool {
	noBar := sb.flags&ScrollableBoxHasHorizontalScrollBM == 0
	if sb.flags&ScrollableBoxAutoHide == 0 || noBar {
		return !noBar
	}
	s := sb.Bounds().Size()
	vs := sb.VirtualSize()
	if vs.Width > s.Width {
		return true
	}
	if sb.flags&ScrollableBoxHasBorderBM != 0 {
		return vs.Width > s.Width-2
	}
	if vs.Height > s.Height {
		// we definitely have the vert bar
		return vs.Width == s.Width
	}
	// vs.Width <= s.Width && vs.Height <= s.Height
	return false
}
