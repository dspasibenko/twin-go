package twin

import (
	"fmt"
	"github.com/dspasibenko/twin-go/pkg/golibs/errors"
	"github.com/gdamore/tcell/v2"
	"reflect"
	"sync"
	"sync/atomic"
)

type Box struct {
	visible atomic.Bool
	active  atomic.Bool
	bounds  atomic.Value
	lock    sync.Mutex
	// owner contains a reference to the owner of the component
	owner Container
	// this is the reference to the BaseComponent holder. This is because
	// a component may "extend" BaseComponent, we need to store the reference
	// to the holder. See Init()
	this   Component
	tpName atomic.Value
	closed atomic.Bool
}

// Init initializes BaseComponent. owner should be non-nil the owner of the Component,
// `this` contains the final struct, which implements Component, but which embed the bc
//
// Init must be called for any instance as first thing after its creation.
func (b *Box) Init(owner Container, this Component) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	return b.init(owner, this)
}

func (b *Box) init(owner Container, this Component) error {
	if this.box() != b {
		return fmt.Errorf("this %s must embed %s: %w", this, b, errors.ErrInvalid)
	}
	if b.owner != nil {
		return fmt.Errorf("this %s already has owner %s: %w", this, b.owner, errors.ErrInvalid)
	}
	o := owner.baseContainer()
	if owner.(Component).box() == b {
		return fmt.Errorf("this %s cannot be added to itself %s: %w", this, owner, errors.ErrInvalid)
	}
	b.visible.Store(true)
	b.active.Store(false)
	if b.bounds.Load() == nil {
		b.bounds.Store(Rectangle{})
	}
	b.tpName.Store(reflect.TypeOf(this).String()) // to be sure that AssertInitialized is nil
	b.this = this
	err := o.addChild(this)
	if err != nil {
		b.this = nil
	} else {
		b.owner = owner
	}
	return err
}

// SetBounds allows to assing the comonent position and dimensions by the `r`
// This call is always trigger re-drawing
func (b *Box) SetBounds(r Rectangle) {
	b.bounds.Store(r.Mend())
	c.reDrawNeeded(b.this)
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
	b.visible.Store(visible)
	c.reDrawNeeded(b.this)
}

// Draw is the BaseComponent drawing procedure which does nothing. It is here to support
// the Component interface, should be re-defined in the derived structure
func (b *Box) Draw(cc *CanvasContext) {
}

// Close allows to close the BaseComponent
func (b *Box) Close() {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.close()
}

// AssertInitialized returns an error if the component is not initialized
func (b *Box) AssertInitialized() error {
	if b.tpName.Load() == nil {
		return fmt.Errorf("Init() is not called %s: %w", b.String(), errors.ErrInvalid)
	}
	if b.isClosed() {
		return fmt.Errorf("%s is closed: %w", b.String(), errors.ErrClosed)
	}
	return nil
}

func (b *Box) isClosed() bool {
	return b.closed.Load()
}

func (b *Box) Owner() Container {
	return b.owner
}

func (b *Box) CanBeFocused() bool { return false }

func (b *Box) OnKeyPressed(ke *tcell.EventKey) bool {
	if b.owner == nil {
		return false
	}
	if ke.Key() == tcell.KeyTab || ke.Key() == tcell.KeyDown {
		return b.owner.baseContainer().nextActive()
	}
	if ke.Key() == tcell.KeyBacktab || ke.Key() == tcell.KeyUp {
		return b.owner.baseContainer().prevActive()
	}
	return false
}

func (b *Box) OnFocus(focused bool) {}

func (b *Box) OnOwnerResized() {}

func (b *Box) OnMousePressed(p Point) bool { return false }

// String returns the `bc` description
func (b *Box) String() string {
	v := b.tpName.Load()
	tp := "N/A"
	if v != nil {
		tp = v.(string)
	}
	return fmt.Sprintf("{Type:%s, Bounds:%s, visible:%t, closed:%t}", tp, b.Bounds(),
		b.IsVisible(), b.closed.Load())
}

func (b *Box) This() Component { return b.this }

func (b *Box) lockIfAlive() bool {
	if b.closed.Load() {
		return false
	}
	b.lock.Lock()
	if b.closed.Load() {
		b.lock.Unlock()
		return false
	}
	return true
}

func (b *Box) box() *Box {
	return b
}

// IsActive returns whether the Box is active or not. Not recommended to override the function.
func (b *Box) IsActive() bool {
	return b.active.Load()
}

func (b *Box) setActive(active bool) {
	b.lock.Lock()
	if !b.IsVisible() {
		b.lock.Unlock()
		panic("box is not visible")
	}
	if b.active.Load() == active {
		return
	}
	b.active.Store(active)
	b.lock.Unlock()
	b.this.OnFocus(active)
	c.reDrawNeeded(b.this)
}

func (b *Box) close() {
	if b.closed.Load() {
		return
	}
	b.closed.Store(true)
	if b.owner != nil {
		b.owner.baseContainer().removeChild(b.this)
	}
	b.owner = nil
	b.this = nil
}
