package components

import (
	"fmt"
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"sync/atomic"
	"time"
)

type ScrollableBox struct {
	twin.BaseContainer
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
)

func (sb *ScrollableBox) Init(owner twin.Container, this twin.Container, style tcell.Style, flags int) error {
	sb.flags = flags
	sb.style = style
	sb.vSize.Store(twin.Size{})
	sb.vOffset.Store(twin.Point{})
	return sb.BaseContainer.Init(owner, this)
}

func (sb *ScrollableBox) Draw(cc *twin.CanvasContext) {
	fmt.Println("Fuck draw", sb.vOffset.Load(), time.Now())
	b := sb.Bounds().Normalized()
	cc.FilledRectangle(b, sb.style)
	if sb.flags&ScrollableBoxHasBorderBM != 0 {
		cc.Rectangle(b, false, sb.style)
	}
	sbSize := b.Size()
	if sb.flags&ScrollableBoxHasVerticalScrollBM != 0 {
		sbSize = sbSize.Add(-1, 0)
	}
	if sb.flags&ScrollableBoxHasHorizontalScrollBM != 0 {
		sbSize = sbSize.Add(0, -1)
	}
	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))
	if sb.flags&ScrollableBoxHasVerticalScrollBM != 0 {
		cc.DrawVScrollBar(b.TopRight(), sbSize.Height, virtSize.Height, visibleSize.Height, virtOffset.Y, sb.style)
	}
	if sb.flags&ScrollableBoxHasHorizontalScrollBM != 0 {
		cc.DrawHScrollBar(b.BottomLeft(), sbSize.Width, virtSize.Width, visibleSize.Width, virtOffset.X, sb.style)
	}
	cc.Print(twin.Point{X: 30 - virtOffset.X, Y: 10 - virtOffset.Y}, "LLLLL", sb.style)
}

func (sb *ScrollableBox) ChildrenCanvasBounds() twin.Rectangle {
	b := sb.Bounds()
	if sb.flags&ScrollableBoxHasBorderBM != 0 {
		b.X, b.Y = b.X+1, b.Y+1
		b.Width = max(0, b.Width-2)
		b.Height = max(0, b.Height-2)
		return b
	}
	if sb.flags&ScrollableBoxHasVerticalScrollBM != 0 {
		b.Width = max(0, b.Width-1)
	}
	if sb.flags&ScrollableBoxHasHorizontalScrollBM != 0 {
		b.Height = max(0, b.Height-1)
	}
	return b
}

func (sb *ScrollableBox) OnMousePressed(p twin.Point) bool {
	b := sb.Bounds().Normalized()
	sbSize := b.Size()
	if sb.flags&ScrollableBoxHasVerticalScrollBM != 0 {
		sbSize = sbSize.Add(-1, 0)
	}
	if sb.flags&ScrollableBoxHasHorizontalScrollBM != 0 {
		sbSize = sbSize.Add(0, -1)
	}
	virtSize := sb.VirtualSize()
	visibleSize := sb.ChildrenCanvasBounds().Size()
	virtSize.Width = max(visibleSize.Width, virtSize.Width)
	virtSize.Height = max(visibleSize.Height, virtSize.Height)
	virtOffset := sb.VirtualOffset()
	virtOffset.X = max(0, min(virtOffset.X, virtSize.Width-visibleSize.Width))
	virtOffset.Y = max(0, min(virtOffset.Y, virtSize.Height-visibleSize.Height))

	if p.X == sbSize.Width && p.Y < sbSize.Height {
		maxY := virtSize.Height - visibleSize.Height
		if p.Y == 0 {
			virtOffset.Y = virtOffset.Y - 1
		} else if p.Y == sbSize.Height-1 {
			virtOffset.Y = virtOffset.Y + 1
		} else if sbSize.Height > 2 {
			virtOffset.Y = int(float64(p.Y) / float64(sbSize.Height-2) * float64(maxY))
		}
		virtOffset.Y = max(0, min(virtOffset.Y, maxY))
		sb.SetVirtualOffset(virtOffset)
		twin.Redraw(sb.This())
	}

	if p.Y == sbSize.Height && p.X < sbSize.Width {
		maxX := virtSize.Width - visibleSize.Width
		if p.X == 0 {
			virtOffset.X = virtOffset.X - 1
		} else if p.X == sbSize.Width-1 {
			virtOffset.X = virtOffset.X + 1
		} else if sbSize.Width > 2 {
			virtOffset.X = int(float64(p.X) / float64(sbSize.Width-2) * float64(maxX))
		}
		virtOffset.X = max(0, min(virtOffset.X, maxX))
		sb.SetVirtualOffset(virtOffset)
		twin.Redraw(sb.This())
	}

	return true
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
