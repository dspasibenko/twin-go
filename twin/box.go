package twin

import (
	"fmt"
	"github.com/dspasibenko/twin-go/pkg/golibs/errors"
	"github.com/gdamore/tcell/v2"
	"reflect"
	"sync/atomic"
)

type Box struct {
	visible atomic.Bool
	active  atomic.Bool
	bounds  atomic.Value
	// owner contains a reference to the owner of the component
	owner Component
	// this is the reference to the BaseComponent holder. This is because
	// a component may "extend" BaseComponent, we need to store the reference
	// to the holder. See Init()
	this   Component
	tpName atomic.Value
	closed atomic.Bool
	chldrn atomic.Value
}

// Init initializes Box. owner should be non-nil the owner of the Component,
// `this` contains the final struct, which implements Component, but which embed the b
//
// Init must be called for any instance as first thing after its creation. No redraw is
// called, because the Bounds are not set yet
func (b *Box) Init(owner Component, this Component) error {
	if this.box() != b {
		return fmt.Errorf("this %s must embed %s: %w", this, b, errors.ErrInvalid)
	}
	if b.owner != nil {
		return fmt.Errorf("this %s already has owner %s: %w", this, b.owner, errors.ErrInvalid)
	}
	if owner.box() == b {
		return fmt.Errorf("this %s cannot be added to itself %s: %w", this, owner, errors.ErrInvalid)
	}
	b.visible.Store(true)
	b.active.Store(false)
	b.chldrn.Store([]Component{})
	if b.bounds.Load() == nil {
		b.bounds.Store(Rectangle{})
	}
	b.tpName.Store(reflect.TypeOf(this).String()) // to be sure that AssertInitialized is nil
	b.this = this
	err := owner.box().addChild(this)
	if err != nil {
		b.this = nil
	} else {
		b.owner = owner
	}
	return err
}

// Bounds returns the component position on its owner coordinates, and its size as rl.RectangleInt32
func (b *Box) Bounds() Rectangle {
	v := b.bounds.Load()
	if v == nil {
		return Rectangle{}
	}
	return v.(Rectangle)
}

func (b *Box) VirtualOffset() Point {
	return Point{}
}

// ChildrenCanvasBounds is same as bounds by default
func (b *Box) ChildrenCanvasBounds() Rectangle {
	return b.Bounds()
}

// IsVisible returns whether the component is visible or not
func (b *Box) IsVisible() bool {
	return b.visible.Load()
}

// SetVisible allows to specify the component visibility, it is always trigger re-drawing
func (b *Box) SetVisible(visible bool) {
	before := b.visible.Swap(visible)
	if before != visible {
		c.reDrawNeeded(b.this, &tcell.EventTime{})
	}
}

// children returns list of owned components
func (b *Box) children() []Component {
	return b.chldrn.Load().([]Component)
}

// Close allows to close the BaseComponent
func (b *Box) close() {
	b.closed.Store(true)
}

func (b *Box) isClosed() bool {
	return b.closed.Load()
}

func (b *Box) CanBeFocused() bool { return false }

// OnDraw is the BaseComponent drawing procedure which does nothing. It is here to support
// the Component interface, should be re-defined in the derived structure
func (b *Box) OnDraw(cc *CanvasContext) {
}

func (b *Box) OnKeyPressed(ke *tcell.EventKey) bool {
	_, chld := b.getActiveChild()
	if chld != nil {
		if chld.OnKeyPressed(ke) {
			return true
		}
	}
	if ke.Key() == tcell.KeyTab || ke.Key() == tcell.KeyDown {
		return b.nextActive()
	}
	if ke.Key() == tcell.KeyBacktab || ke.Key() == tcell.KeyUp {
		return b.prevActive()
	}
	return false
}

func (b *Box) OnFocus(focused bool) {}

func (b *Box) OnOwnerResized() {}

func (b *Box) OnMousePressed(p Point) bool { return false }

func (b *Box) OnClosed() {}

func (b *Box) OnChildClosed(child Component) {}

// String returns the `bc` description
func (b *Box) String() string {
	v := b.tpName.Load()
	tp := "N/A"
	if v != nil {
		tp = v.(string)
	}
	return fmt.Sprintf("{Type:%s, Bounds:%s, visible:%t, closed:%t, active:%t, children:%d}", tp, b.Bounds(),
		b.IsVisible(), b.closed.Load(), b.isActive(), len(b.children()))
}

func (b *Box) box() *Box {
	return b
}

// isActive returns whether the Box is active or not. Not recommended to override the function.
func (b *Box) isActive() bool {
	return b.active.Load()
}

// SetBounds allows to assign the component position and dimensions by the `r`
// This call is always trigger re-drawing
func (b *Box) SetBounds(r Rectangle) {
	b.bounds.Store(r.Mend())
	c.resize(b.this)
}

func (b *Box) addChild(comp Component) error {
	c.lock.Lock()
	defer c.lock.Unlock()

	if b.isClosed() {
		return errors.ErrClosed
	}
	cb := comp.box()
	if cb.owner != nil && cb.owner != b.this {
		return fmt.Errorf("the component %s, already has an other owner: %w", cb, errors.ErrInvalid)
	}

	children := b.chldrn.Load().([]Component)
	idx := childIndex(children, comp)
	var nv []Component
	if idx < len(children) {
		// c is in the list, change its position then, briniging on top
		nv = make([]Component, 0, len(children))
		nv = append(nv, children[:idx]...)
		nv = append(nv, children[idx+1:]...)
	} else {
		nv = make([]Component, 0, len(children)+1)
		nv = append(nv, children...)
	}
	nv = append(nv, comp)
	b.chldrn.Store(nv)
	return nil
}

func (b *Box) updateChildren(children []Component) {
	b.chldrn.Store(children)
}

func (b *Box) setActive(active bool) {
	if !b.IsVisible() {
		panic("box is not visible")
	}
	if b.active.Load() == active {
		return
	}
	b.active.Store(active)
	b.this.OnFocus(active)
	c.reDrawNeeded(b.this, &tcell.EventTime{})
}

func (b *Box) removeClosedChildren() {
	if b.isClosed() {
		b.chldrn.Store([]Component(nil))
		return
	}
	children := []Component{}
	for _, child := range b.children() {
		if child.box().isClosed() {
			continue
		}
		children = append(children, child)
	}
	b.chldrn.Store(children)
}

func (b *Box) closeActually() {
	b.this.OnClosed()
	if b.owner != nil {
		b.owner.OnChildClosed(b.this)
	}
	b.owner = nil
	b.this = nil
}

func childIndex(children []Component, c Component) int {
	for idx, c1 := range children {
		if c1 == c {
			return idx
		}
	}
	return len(children)
}

func (b *Box) getActiveChild() (int, Component) {
	children := b.children()
	for i, child := range children {
		if child.box().isActive() {
			return i, child
		}
	}
	return -1, nil
}

func (b *Box) nextActive() bool {
	i, comp := b.getActiveChild()
	i++
	setActiveFalse(comp)
	children := b.children()
	for i < len(children) {
		child := children[i]
		if child.IsVisible() && child.CanBeFocused() {
			child.box().setActive(true)
			return true
		}
		i++
	}
	return false
}

func (b *Box) prevActive() bool {
	i, comp := b.getActiveChild()
	setActiveFalse(comp)
	children := b.children()
	if i < 0 {
		i = len(children)
	}
	i--
	for i >= 0 && i < len(children) {
		child := children[i]
		if child.IsVisible() && child.CanBeFocused() {
			child.box().setActive(true)
			return true
		}
		i--
	}
	return false
}

func setActiveFalse(comp Component) {
	if comp == nil {
		return
	}
	comp.box().setActive(false)
	_, chld := comp.box().getActiveChild()
	setActiveFalse(chld)
}
