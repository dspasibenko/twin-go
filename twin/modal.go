package twin

import (
	"github.com/gdamore/tcell/v2"
)

type modalPad struct {
	Box
}

func (m *modalPad) CanBeFocused() bool { return true }

func (m *modalPad) OnMousePressed(p Point) bool { return true }

func (m *modalPad) OnChildClosed(child Component) {
	m.close()
}

func (m *modalPad) OnKeyPressed(ke *tcell.EventKey) bool {
	if ke.Key() == tcell.KeyEscape {
		m.close()
		return true
	}
	_, chld := m.getActiveChild()
	if chld == nil {
		chldrn := m.children()
		if len(chldrn) == 0 {
			return true
		}
		chld = chldrn[0]
		chld.box().setActive(true)
	}
	chld.OnKeyPressed(ke)
	return true
}

func (m *modalPad) OnOwnerResized() {
	m.SetBounds(c.root.Bounds())
}
