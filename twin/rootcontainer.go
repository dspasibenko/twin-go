package twin

import "github.com/gdamore/tcell/v2"

type rootContainer struct {
	Box
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
	r.chldrn.Store([]Component(nil))
	r.SetVisible(true)
}

// IsVisible for rootContainer is always true
func (r *rootContainer) IsVisible() bool {
	return true
}

// Draw for the display - either the background color or a wallpaper picture
func (r *rootContainer) OnDraw(cc *CanvasContext) {
	cc.FilledRectangle(r.Bounds(), r.style)
}

func (r *rootContainer) CanBeFocused() bool {
	return true
}
