package components

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"strings"
	"sync"
)

type Label struct {
	twin.Box
	allignment int
	style      tcell.Style
	pureText   string
	lines      []string
	offsets    []twin.Point
	lock       sync.Mutex
}

func NewLabel(owner twin.Component, text string, allignment int, style tcell.Style) (*Label, error) {
	l := &Label{}
	l.style = style
	l.allignment = allignment
	err := l.Init(owner, l)
	if err != nil {
		return nil, err
	}
	l.pureText = text
	l.SetText(text)
	return l, nil
}

func (l *Label) CanBeFocused() bool {
	return false
}

func (l *Label) SetBounds(r twin.Rectangle) {
	l.setText(l.pureText, r)
	l.Box.SetBounds(r)
}

func (l *Label) SetText(text string) {
	l.setText(text, l.Bounds())
}

func (l *Label) OnDraw(cc *twin.CanvasContext) {
	b := l.Bounds().Normalized()
	cc.FilledRectangle(b, l.style)
	for i, line := range l.lines {
		if i > b.Height {
			break
		}
		cc.Print(l.offsets[i], line, l.style)
	}
}

func (l *Label) setText(text string, b twin.Rectangle) {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.lines = []string{}
	l.offsets = []twin.Point{}
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		ln := strings.TrimSpace(line)
		ln = runewidth.Truncate(ln, b.Width, "")
		w := runewidth.StringWidth(ln)
		l.lines = append(l.lines, ln)
		switch l.allignment {
		case AllignLeft:
			l.offsets = append(l.offsets, twin.Point{X: 0, Y: i})
		case AllignRight:
			l.offsets = append(l.offsets, twin.Point{X: b.Width - w, Y: i})
		case AllignCenter:
			l.offsets = append(l.offsets, twin.Point{X: max(0, (b.Width-w)/2), Y: i})
		}
	}
}
