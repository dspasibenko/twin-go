package components

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
)

type Button struct {
	twin.Box
	bs      ButtonStyle
	txtOffs int
}

type ButtonStyle struct {
	style       tcell.Style
	activeStyle tcell.Style
	text        string
	allignment  int
	rect        twin.Rectangle
	onEnter     func(b *Button)
}

func (bs ButtonStyle) Style(style tcell.Style) ButtonStyle {
	bs.style = style
	return bs
}

func (bs ButtonStyle) ActiveStyle(style tcell.Style) ButtonStyle {
	bs.activeStyle = style
	return bs
}

func (bs ButtonStyle) Text(text string) ButtonStyle {
	bs.text = text
	return bs
}

func (bs ButtonStyle) Allignment(a int) ButtonStyle {
	bs.allignment = a
	return bs
}

func (bs ButtonStyle) Rectangle(rect twin.Rectangle) ButtonStyle {
	bs.rect = rect
	return bs
}

func (bs ButtonStyle) OnEnter(f func(b *Button)) ButtonStyle {
	bs.onEnter = f
	return bs
}

func NewButton(owner twin.Component, bs ButtonStyle) (*Button, error) {
	b := &Button{bs: bs}
	err := b.Box.Init(owner, b)
	if err != nil {
		return nil, err
	}
	w := runewidth.StringWidth(bs.text)
	offs := 0
	switch bs.allignment {
	case AllignLeft:
		offs = 0
	case AllignRight:
		offs = bs.rect.Width - w
	case AllignCenter:
		offs = (bs.rect.Width - w) / 2
	}
	b.txtOffs = offs
	b.SetBounds(bs.rect)
	return b, nil
}

func (b *Button) CanBeFocused() bool { return true }

func (b *Button) OnMousePressed(p twin.Point) bool { return b.onEnter() }

func (b *Button) OnKeyPressed(ke *tcell.EventKey) bool {
	if ke.Key() == tcell.KeyEnter {
		return b.onEnter()
	}
	return false
}

func (b *Button) OnDraw(cc *twin.CanvasContext) {
	r := b.Bounds().Normalized()
	style := b.bs.style
	if twin.IsActive(b) {
		style = b.bs.activeStyle
	}
	cc.FilledRectangle(r, style)
	cc.Print(twin.Point{X: b.txtOffs, Y: 0}, b.bs.text, style)
}

func (b *Button) onEnter() bool {
	if b.bs.onEnter != nil {
		b.bs.onEnter(b)
		return true
	}
	return false
}
