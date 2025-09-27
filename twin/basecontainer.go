package twin

import (
	"fmt"
	"github.com/dspasibenko/twin-go/pkg/golibs/errors"
	"github.com/gdamore/tcell/v2"
	"sync/atomic"
)

type BaseContainer struct {
	Box

	children atomic.Value
}

// Init initializes BaseContainer, returns nil if the component initialized successfully
func (bc *BaseContainer) Init(owner Container, this Component) error {
	bc.lock.Lock()
	defer bc.lock.Unlock()
	if err := bc.init(owner, this); err != nil {
		return err
	}
	bc.children.Store([]Component(nil))
	return nil
}

func (bc *BaseContainer) SetBounds(r Rectangle) {
	bc.Box.SetBounds(r)
	bc.callForChildren(func(c Component) { c.OnOwnerResized() })
}

func (bc *BaseContainer) ChildrenBounds() Rectangle {
	chldrn := bc.Children()
	var r Rectangle
	if len(chldrn) == 0 {
		return r
	}
	r = chldrn[0].Bounds()
	for _, chld := range chldrn[1:] {
		b := chld.Bounds()
		r.X = min(r.X, b.X)
		r.Y = min(r.Y, b.Y)
		mx := b.X + b.Width
		if mx > r.X+r.Width {
			r.Width = mx - r.X
		}
		my := b.Y + b.Height
		if my > r.Y+r.Height {
			r.Height = my - r.Y
		}
	}
	return r
}

func (bc *BaseContainer) OnKeyPressed(ke *tcell.EventKey) bool {
	_, active := bc.getActive()
	if active != nil && active.OnKeyPressed(ke) {
		return true
	}
	if ke.Key() == tcell.KeyTab || ke.Key() == tcell.KeyDown {
		return bc.nextActive()
	}
	if ke.Key() == tcell.KeyBacktab || ke.Key() == tcell.KeyUp {
		return bc.prevActive()
	}

	return bc.box().OnKeyPressed(ke)
}

func (bc *BaseContainer) callForChildren(f func(Component)) {
	for _, child := range bc.Children() {
		f(child)
	}
}

// OnAddChild is the default implementation, please see Container interface
func (bc *BaseContainer) onAddChild(c Component, children []Component) ([]Component, error) {
	idx := childIndex(children, c)
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
	return append(nv, c), nil
}

// addChild adds the new comopnent c to the container
func (bc *BaseContainer) addChild(c Component) error {
	if !bc.lockIfAlive() {
		return fmt.Errorf("AddChild: failed to add %s to the %s container, which is not initialized: %w", c, bc, errors.ErrClosed)
	}
	defer bc.lock.Unlock()

	cb := c.box()
	if err := cb.AssertInitialized(); err != nil {
		return err
	}
	if cb.owner != nil && cb.owner != bc.this.(Container).baseContainer() {
		return fmt.Errorf("the component %s, already has an owner: %w", cb, errors.ErrInvalid)
	}

	v := bc.children.Load().([]Component)
	nv, err := bc.this.(Container).baseContainer().onAddChild(c, v)
	if err != nil {
		return err
	}
	bc.children.Store(nv)
	return nil
}

func (bc *BaseContainer) baseContainer() *BaseContainer {
	return bc
}

// removeChild removes the component comp from the container
func (bc *BaseContainer) removeChild(comp Component) bool {
	if !bc.lockIfAlive() {
		return false
	}
	defer bc.lock.Unlock()

	v := bc.children.Load().([]Component)
	idx := childIndex(v, comp)
	if idx < len(v) {
		c.reDrawNeeded(bc.this)
		nv := make([]Component, 0, len(v)-1)
		nv = append(nv, v[:idx]...)
		nv = append(nv, v[idx+1:]...)
		bc.children.Store(nv)
		return true
	}
	return false
}

// Children returns list of owned components
func (bc *BaseContainer) Children() []Component {
	return bc.children.Load().([]Component)
}

// Close terminates the Container and all its children
func (bc *BaseContainer) Close() {
	if !bc.lockIfAlive() {
		return
	}
	children := bc.children.Load().([]Component)
	bc.children.Store([]Component(nil))
	bc.close()
	bc.lock.Unlock()
	for _, c := range children {
		c.Close()
	}
}

func (bc *BaseContainer) CanBeFocused() bool { return true }

// String returns the BaseContainer string representation
func (bc *BaseContainer) String() string {
	return fmt.Sprintf("{BC: %s, children: %d}", bc.box(), len(bc.Children()))
}

func childIndex(children []Component, c Component) int {
	for idx, c1 := range children {
		if c1 == c {
			return idx
		}
	}
	return len(children)
}

func (bc *BaseContainer) getActive() (int, Component) {
	children := bc.Children()
	for i, child := range children {
		if child.box().IsActive() {
			return i, child
		}
	}
	return -1, nil
}

func (bc *BaseContainer) nextActive() bool {
	i, comp := bc.getActive()
	i++
	setActiveFalse(comp)
	children := bc.Children()
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

func (bc *BaseContainer) prevActive() bool {
	i, comp := bc.getActive()
	setActiveFalse(comp)
	children := bc.Children()
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
	if cont, ok := comp.(Container); ok {
		_, child := cont.baseContainer().getActive()
		setActiveFalse(child)
	}
}
