package components

import (
	"fmt"
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"strings"
)

type ListBox struct {
	ScrollableBox
	lbs      ListBoxStyle
	selected int
}

type ListBoxStyle struct {
	ScrollableBoxStyle
	listbox   string
	selStyle  *tcell.Style
	selActive *tcell.Style
}

func (lbs ListBoxStyle) WithListbox(lb string) ListBoxStyle {
	lbs.listbox = lb
	return lbs
}

func (lbs ListBoxStyle) WithSelStyle(ss tcell.Style) ListBoxStyle {
	lbs.selStyle = &ss
	return lbs
}

func (lbs ListBoxStyle) SelStyle() tcell.Style {
	if lbs.selStyle != nil {
		return *lbs.selStyle
	}
	lbt := GetThemeValue[ListBoxTheme](lbs.listbox)
	return lbt.SelStyle
}

func (lbs ListBoxStyle) WithSelActiveStyle(sas tcell.Style) ListBoxStyle {
	lbs.selActive = &sas
	return lbs
}

func (lbs ListBoxStyle) SelActiveStyle() tcell.Style {
	if lbs.selStyle != nil {
		return *lbs.selActive
	}
	lbt := GetThemeValue[ListBoxTheme](lbs.listbox)
	return lbt.SelActive
}

func (lb *ListBox) Init(owner, this twin.Component, lbs ListBoxStyle) error {
	if lbs.listbox == "" {
		lbs.listbox = "listbox"
	}
	lbs.ScrollableBoxStyle.win = lbs.listbox + "Win"
	lb.selected = 0
	lb.lbs = lbs
	return lb.ScrollableBox.Init(owner, this, lbs.ScrollableBoxStyle)
}

func (lb *ListBox) OnDraw(cc *twin.CanvasContext) {
	lb.ScrollableBox.OnDraw(cc)
	b := lb.listBounds()
	idx := lb.VirtualOffset().Y
	vs := lb.VirtualSize()
	for y := b.Y; y <= b.Height; y++ {
		if idx >= vs.Height {
			break
		}
		var s tcell.Style
		str := lb.Line(idx)
		if lb.selected == idx {
			if twin.IsActive(lb) {
				s = lb.lbs.SelActiveStyle()
			} else {
				s = lb.lbs.SelStyle()
			}
			if len(str) < b.Width {
				str += strings.Repeat(" ", b.Width-len(str))
			}
		} else {
			s = lb.lbs.Style()
		}
		cc.PrintL(twin.Point{b.X, y}, str, b.Width, s)
		idx++
	}
}

func (lb *ListBox) OnMousePressed(p twin.Point) bool {
	b := lb.listBounds()
	if b.Contains(p) {
		idx := p.Y - b.Y
		idx += lb.VirtualOffset().Y
		vs := lb.VirtualSize()
		if idx < vs.Height && idx != lb.selected {
			lb.selected = idx
			twin.Redraw(twin.This(lb))
		}
		return true
	}
	return lb.ScrollableBox.OnMousePressed(p)
}

func (lb *ListBox) OnKeyPressed(ke *tcell.EventKey) bool {
	vs := lb.VirtualSize()
	b := lb.listBounds()
	idx := lb.selected
	switch ke.Key() {
	case tcell.KeyUp:
		idx--
	case tcell.KeyDown:
		idx++
	case tcell.KeyPgUp:
		idx -= b.Height
	case tcell.KeyPgDn:
		idx += b.Height
	case tcell.KeyHome:
		idx = 0
	case tcell.KeyEnd:
		idx = vs.Height - 1
	}
	if idx != lb.selected {
		idx2 := max(0, min(idx, vs.Height-1))
		if lb.selected != idx2 {
			lb.selected = idx2
			vo := lb.VirtualOffset()
			if lb.selected < vo.Y {
				vo.Y = lb.selected
				lb.SetVirtualOffset(vo)
			}
			if lb.selected >= vo.Y+b.Height {
				vo.Y = lb.selected - b.Height + 1
				lb.SetVirtualOffset(vo)
			}
			twin.Redraw(twin.This(lb))
		}
		return true
	}
	return lb.ScrollableBox.OnKeyPressed(ke)
}

func (lb *ListBox) Line(idx int) string {
	return fmt.Sprintf("%d", idx)
}

func (lb *ListBox) listBounds() twin.Rectangle {
	b := lb.Bounds().Normalized()
	if lb.HasBorder() {
		b.X, b.Y = b.X+1, b.Y+1
		b.Width, b.Height = b.Width-2, b.Height-2
	} else {
		if lb.HasVBar() {
			b.Width--
		}
		if lb.HasHBar() {
			b.Height--
		}
	}
	return b
}
