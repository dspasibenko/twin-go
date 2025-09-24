package twin

import "github.com/gdamore/tcell/v2"

type rootContainer struct {
	BaseContainer
	style tcell.Style
}

func newRootContainer() *rootContainer {
	r := &rootContainer{}
	r.style = tcell.StyleDefault.Background(tcell.ColorBlack)
	r.init()
	return r
}

func (r *rootContainer) init() {
	r.this = r
	r.children.Store([]Component(nil))
	r.SetVisible(true)
}

// IsVisible for rootContainer is always true
func (r *rootContainer) IsVisible() bool {
	return true
}

// Close is overwritten for rootContainer due to no owner notification about the close
// (it doesn't call close() for itself comparing to BaseContainer)
func (r *rootContainer) Close() {
	if !r.lockIfAlive() {
		return
	}
	children := r.children.Load().([]Component)
	r.children.Store([]Component(nil))
	r.closed.Store(true)
	r.lock.Unlock()
	for _, c := range children {
		c.Close()
	}
}

// Draw for the display - either the background color or a wallpaper picture
func (r *rootContainer) Draw(cc *CanvasContext) {
	b := r.Bounds()
	cc.FilledRectangle(r.Bounds(), r.style)
	cc.Print(Point{b.Width / 2, b.Height / 2}, "X", tcell.StyleDefault.Foreground(tcell.ColorWhite))
}
