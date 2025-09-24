package twin

import (
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"strings"
)

// CanvasContext struct allows to track the stack of regions, so that one includes another.
type CanvasContext struct {
	stack []ctxStackElem
}

type ctxStackElem struct {
	p Point
	r Rectangle
}

// isVisible returns whether the region r is visible on the stack of regions.
// It returns true, if at least one cell is visible
func (cc *CanvasContext) isVisible(r Rectangle) bool {
	p := cc.physicalPointXY(r.TopLeft())
	pr := cc.physicalRegion()
	return !(p.X+r.Width < pr.X || pr.X+pr.Width < p.X ||
		p.Y+r.Height < pr.Y || pr.Y+pr.Height < p.Y)
}

// Print prints the text using the style provided. x and y are coordinates of the text
// on the owner's rectangle
func (cc *CanvasContext) Print(p Point, str string, style tcell.Style) {
	pp := cc.physicalPointXY(p)
	pr := cc.physicalRegion()
	prBR := pr.BottomRight()
	if pp.Y < pr.Y || prBR.Y < pp.Y {
		return // is not visible vertically
	}
	for _, chr := range str {
		w := runewidth.RuneWidth(chr)
		if w == 0 {
			chr = ' '
			w = 1
		}
		if pp.X > prBR.X || prBR.X < pp.X+w-1 {
			return // moving right of the rightest border
		}
		if pp.X >= pr.X {
			// only put the chr on the screen if it's completely visible on the rectangle (pr)
			c.s.SetContent(int(pp.X), int(pp.Y), chr, nil, style)
		}
		pp.X += w
	}
}

func (cc *CanvasContext) FilledRectangle(r Rectangle, style tcell.Style) {
	str := strings.Repeat(" ", int(r.Width))
	for i := 0; i < r.Height; i++ {
		cc.Print(Point{X: r.X, Y: r.Y + i}, str, style)
	}
}

func (cc *CanvasContext) Rectangle(r Rectangle, doubleLines bool, style tcell.Style) {
	if r.Width <= 0 || r.Height <= 0 {
		return
	}

	var h, v, tl, tr, bl, br rune
	if doubleLines {
		h, v = '═', '║'
		tl, tr, bl, br = '╔', '╗', '╚', '╝'
	} else {
		h, v = '─', '│'
		tl, tr, bl, br = '┌', '┐', '└', '┘'
	}

	if r.Width == 1 && r.Height == 1 {
		cc.Print(r.TopLeft(), "+", style)
		return
	}

	if r.Height == 1 {
		cc.Print(r.TopLeft(), strings.Repeat(string(h), int(r.Width)), style)
		return
	}

	if r.Width == 1 {
		for y := 0; y < r.Height; y++ {
			cc.Print(Point{X: r.X, Y: r.Y + y}, string(v), style)
		}
		return
	}

	cc.Print(r.TopLeft(), string(tl), style)
	cc.Print(Point{X: r.X + r.Width - 1, Y: r.Y}, string(tr), style)
	cc.Print(Point{X: r.X, Y: r.Y + r.Height - 1}, string(bl), style)
	cc.Print(r.BottomRight(), string(br), style)

	if r.Width > 2 {
		cc.Print(Point{X: r.X + 1, Y: r.Y}, strings.Repeat(string(h), int(r.Width-2)), style)
		cc.Print(Point{X: r.X + 1, Y: r.Y + r.Height - 1}, strings.Repeat(string(h), int(r.Width-2)), style)
	}

	for y := 1; y < r.Height-1; y++ {
		cc.Print(Point{X: r.X, Y: r.Y + y}, string(v), style)
		cc.Print(Point{X: r.X + r.Width - 1, Y: r.Y + y}, string(v), style)
	}
}

// DrawVScrollBar draws the vertical scroll bar at the position pos:
// sbSize - the size of scroll bar in cells
// virtSize - the size of the field
// wSize - the visible window size (wSize <= virtSize)
// vOffset - the virtual offset >= 0
func (cc *CanvasContext) DrawVScrollBar(pos Point, sbSize, virtSize, wSize, vOffset int, style tcell.Style) {
	maxOffset := max(0, virtSize-wSize)
	vOffset = max(0, min(vOffset, maxOffset))

	bar := " "
	bg := "░"
	switch {
	case sbSize == 1:
		cc.Print(pos, bar, style)
		return

	case sbSize == 2:
		cc.Print(pos, bar, style)
		pos.Y++
		cc.Print(pos, bar, style)
		return

	default:
		cc.Print(pos, "▲", style)
		cc.Print(Point{X: pos.X, Y: pos.Y + sbSize - 1}, "▼", style)
		pos.Y++

		trackLen := sbSize - 2

		thumbLen := min(trackLen, max(1, wSize*trackLen/virtSize))
		offs := 0
		if maxOffset > 0 {
			offs = vOffset * (trackLen - thumbLen) / maxOffset
		}

		for i := 0; i < offs; i++ {
			cc.Print(pos, bg, style)
			pos.Y++
		}
		fg, _, _ := style.Decompose()
		for i := 0; i < thumbLen; i++ {
			cc.Print(pos, bar, tcell.StyleDefault.Background(fg))
			pos.Y++
		}
		for i := 0; i < trackLen-offs-thumbLen; i++ {
			cc.Print(pos, bg, style)
			pos.Y++
		}
	}
}

func (cc *CanvasContext) DrawHScrollBar(pos Point, sbSize, virtSize, wSize, hOffset int, style tcell.Style) {
	if sbSize <= 0 || virtSize <= 0 || wSize <= 0 {
		return
	}

	maxOffset := max(0, virtSize-wSize)
	hOffset = max(0, min(hOffset, maxOffset))

	bar := "█"
	bg := "░"

	switch {
	case sbSize == 1:
		cc.Print(pos, bar, style)

	case sbSize == 2:
		cc.Print(pos, bar+bar, style)

	default:
		// стрелки
		cc.Print(pos, "◀", style)
		cc.Print(Point{X: pos.X + sbSize - 1, Y: pos.Y}, "▶", style)

		trackLen := sbSize - 2
		thumbLen := min(trackLen, max(1, wSize*trackLen/virtSize))

		offs := 0
		if maxOffset > 0 {
			offs = hOffset * (trackLen - thumbLen) / maxOffset
		}

		cc.Print(Point{X: pos.X + 1, Y: pos.Y}, strings.Repeat(bg, trackLen), style)
		cc.Print(Point{X: pos.X + 1 + offs, Y: pos.Y}, strings.Repeat(bar, thumbLen), style)
	}
}

// physicalPointXY returns the coordinates for a Component's point (x,y) on the
// physical display.
func (cc *CanvasContext) physicalPointXY(vp Point) Point {
	cse := cc.stack[len(cc.stack)-1]
	return Point{vp.X - cse.p.X + cse.r.X, vp.Y - cse.p.Y + cse.r.Y}
}

// physicalRegion returns the region for the physical screen
func (cc *CanvasContext) physicalRegion() Rectangle {
	return cc.stack[len(cc.stack)-1].r
}

// newCanvas constructs the new instance of CanvasContext with the physical dimensions
func newCanvas(s Size) *CanvasContext {
	cc := &CanvasContext{}
	disp := ctxStackElem{r: Rectangle{X: 0, Y: 0, Width: s.Width, Height: s.Height}}
	cc.stack = append(cc.stack, disp) // the cc.stack[0] is always the display resolution
	return cc
}

// pushRelativeRegion adds the physical region (the display coordinates) for the region r, which
// is defined in its parent coordinates stored on top of the stack. vp defines the virtual offset
// in the r. After the call the top of the stack will contain the physical region for r and its
// virtual point for calculation of the region r children, if any...
func (cc *CanvasContext) pushRelativeRegion(vp Point, r Rectangle) {
	cse := cc.stack[len(cc.stack)-1]
	r.X += cse.r.X // physical X
	r.Y += cse.r.Y // physical Y
	r.X -= cse.p.X // make a correction to the virtual offset for X
	r.Y -= cse.p.Y // and Y
	if r.X < cse.r.X {
		r.Width = max(0, r.Width-(cse.r.X-r.X))
		vp.X += cse.r.X - r.X
		r.X = cse.r.X
	}
	if r.Y < cse.r.Y {
		r.Height = max(0, r.Height-(cse.r.Y-r.Y))
		vp.Y += cse.r.Y - r.Y
		r.Y = cse.r.Y
	}
	r.Width = max(0, min(r.Width, cse.r.Width-(r.X-cse.r.X)))
	r.Height = max(0, min(r.Height, cse.r.Height-(r.Y-cse.r.Y)))
	cc.stack = append(cc.stack, ctxStackElem{p: vp, r: r})
}

func (cc *CanvasContext) pop() {
	if len(cc.stack) < 2 {
		panic("pop() for empty stack called")
	}
	cc.stack = cc.stack[:len(cc.stack)-1]
}

// relativePoingXY gets the physical point p and turns it to the canvas relative point {x, y}
func (cc *CanvasContext) relativePointXY(p Point) Point {
	cse := cc.stack[len(cc.stack)-1]
	return Point{p.X - cse.r.X + cse.p.X, p.Y - cse.r.Y + cse.p.Y}
}

func (cc *CanvasContext) isEmpty() bool {
	return len(cc.stack) == 1
}
