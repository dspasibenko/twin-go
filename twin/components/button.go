package components

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"sync"
	"sync/atomic"
)

type Button struct {
	twin.Box
	bs      ButtonStyle
	lock    sync.Mutex
	txtOffs atomic.Int32
}

type ButtonStyle struct {
	button      ButtonType
	style       *tcell.Style
	activeStyle *tcell.Style
	text        string
	allignment  *TextAlignment
	rect        twin.Rectangle
	onEnter     func(b *Button)
}

func (bs ButtonStyle) WithButtonType(buttonType ButtonType) {
	bs.button = buttonType
}

func (bs ButtonStyle) WithStyle(style tcell.Style) ButtonStyle {
	bs.style = &style
	return bs
}

func (bs ButtonStyle) Style() tcell.Style {
	if bs.style != nil {
		return *bs.style
	}
	return GetThemeValue[ButtonTheme](string(bs.button)).NotActive
}

func (bs ButtonStyle) WithActiveStyle(style tcell.Style) ButtonStyle {
	bs.activeStyle = &style
	return bs
}

func (bs ButtonStyle) ActiveStyle() tcell.Style {
	if bs.activeStyle != nil {
		return *bs.activeStyle
	}
	return GetThemeValue[ButtonTheme](string(bs.button)).Active
}

func (bs ButtonStyle) WithText(text string) ButtonStyle {
	bs.text = text
	return bs
}

func (bs ButtonStyle) WithAllignment(a TextAlignment) ButtonStyle {
	bs.allignment = &a
	return bs
}

func (bs ButtonStyle) Allignment() TextAlignment {
	if bs.style != nil {
		return *bs.allignment
	}
	return GetThemeValue[ButtonTheme](string(bs.button)).Alignment
}

func (bs ButtonStyle) WithRectangle(rect twin.Rectangle) ButtonStyle {
	bs.rect = rect
	return bs
}

func (bs ButtonStyle) WithOnEnter(f func(b *Button)) ButtonStyle {
	bs.onEnter = f
	return bs
}

func NewButton(owner twin.Component, bs ButtonStyle) (*Button, error) {
	if bs.button == "" {
		bs.button = NormalButtonType
	}
	b := &Button{bs: bs}
	err := b.Box.Init(owner, b)
	if err != nil {
		return nil, err
	}
	b.Box.SetBounds(bs.rect)
	b.recalc(bs.rect)
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

func (b *Button) SetBounds(rect twin.Rectangle) {
	b.lock.Lock()
	defer b.lock.Unlock()
	bnds := b.Bounds()
	if bnds == rect {
		return
	}
	b.Box.SetBounds(rect)
	b.recalc(rect)
}

func (b *Button) OnDraw(cc *twin.CanvasContext) {
	r := b.Bounds().Normalized()
	var style tcell.Style
	if twin.IsActive(b) {
		style = b.bs.ActiveStyle()
	} else {
		style = b.bs.Style()
	}

	cc.FilledRectangle(r, style)
	cc.Print(twin.Point{X: int(b.txtOffs.Load()), Y: 0}, b.bs.text, style)
}

func (b *Button) recalc(r twin.Rectangle) {
	w := runewidth.StringWidth(b.bs.text)
	offs := int32(0)
	switch b.bs.Allignment() {
	case AllignLeft:
		offs = 0
	case AllignRight:
		offs = int32(r.Width - w)
	case AllignCenter:
		offs = int32((r.Width - w) / 2)
	}
	b.bs.rect = r // not necessary, but just in case
	b.txtOffs.Store(offs)
}

func (b *Button) onEnter() bool {
	if b.bs.onEnter != nil {
		b.bs.onEnter(b)
		return true
	}
	return false
}
