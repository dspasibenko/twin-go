package components

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"strings"
	"sync"
	"sync/atomic"
)

type Label struct {
	twin.Box
	ls   LabelStyle
	ll   atomic.Value // labelLines
	lock sync.Mutex
}

type labelLines struct {
	lines   []string
	offsets []twin.Point
}

type LabelStyle struct {
	allignment *TextAlignment
	style      *tcell.Style
	pureText   string
	rect       twin.Rectangle
}

func (ls LabelStyle) WithAlignment(allignment TextAlignment) LabelStyle {
	ls.allignment = &allignment
	return ls
}

func (ls LabelStyle) WithStyle(style tcell.Style) LabelStyle {
	ls.style = &style
	return ls
}

func (ls LabelStyle) WithPureText(pureText string) LabelStyle {
	ls.pureText = pureText
	return ls
}

func (ls LabelStyle) WithRectangle(r twin.Rectangle) LabelStyle {
	ls.rect = r
	return ls
}

func (ls LabelStyle) Style() tcell.Style {
	if ls.style != nil {
		return *ls.style
	}
	lbs := GetThemeValue[LabelTheme]("label")
	return lbs.Style
}

func (ls LabelStyle) Allignment() TextAlignment {
	if ls.allignment != nil {
		return *ls.allignment
	}
	lbs := GetThemeValue[LabelTheme]("label")
	return lbs.Alignment
}

func NewLabel(owner twin.Component, ls LabelStyle) (*Label, error) {
	l := &Label{}
	l.ls = ls
	err := l.Init(owner, l)
	if err != nil {
		return nil, err
	}
	l.Box.SetBounds(ls.rect)
	l.SetText(ls.pureText)
	return l, nil
}

func (l *Label) CanBeFocused() bool {
	return false
}

func (l *Label) SetBounds(r twin.Rectangle) {
	l.setText(l.ls.pureText, r)
	l.Box.SetBounds(r)
}

func (l *Label) SetText(text string) {
	l.setText(text, l.Bounds())
}

func (l *Label) OnDraw(cc *twin.CanvasContext) {
	b := l.Bounds().Normalized()
	stl := l.ls.Style()
	cc.FilledRectangle(b, stl)
	ll := l.ll.Load().(labelLines)
	for i, line := range ll.lines {
		if i > b.Height {
			break
		}
		cc.Print(ll.offsets[i], line, stl)
	}
}

func (l *Label) setText(text string, b twin.Rectangle) {
	l.lock.Lock()
	defer l.lock.Unlock()
	ll := labelLines{}
	allignment := l.ls.Allignment()
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		ln := strings.TrimSpace(line)
		ln = runewidth.Truncate(ln, b.Width, "")
		w := runewidth.StringWidth(ln)
		ll.lines = append(ll.lines, ln)
		switch allignment {
		case AllignLeft:
			ll.offsets = append(ll.offsets, twin.Point{X: 0, Y: i})
		case AllignRight:
			ll.offsets = append(ll.offsets, twin.Point{X: b.Width - w, Y: i})
		case AllignCenter:
			ll.offsets = append(ll.offsets, twin.Point{X: max(0, (b.Width-w)/2), Y: i})
		}
	}
	l.ll.Store(ll)
}
