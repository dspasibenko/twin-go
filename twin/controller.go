package twin

import (
	"context"
	"github.com/gdamore/tcell/v2"
	"sync"
	"sync/atomic"
)

type controller struct {
	s        tcell.Screen
	root     *rootContainer
	done     chan struct{}
	runs     atomic.Bool
	lock     sync.Mutex
	dirtySet map[Component]bool
}

type resizeEvent struct {
	tcell.EventTime
	comp Component
}

type activateEvent struct {
	tcell.EventTime
	comp Component
}

var c *controller

func init() {
	c = new(controller)
	var err error
	c.s, err = tcell.NewScreen()
	if err != nil {
		panic(err)
	}
	if err := c.s.Init(); err != nil {
		panic(err)
	}
	c.s.EnableMouse()
	//	c.s.EnablePaste()
	//	c.s.EnableFocus()
	c.dirtySet = make(map[Component]bool)
	c.done = make(chan struct{})
	c.root = newRootContainer()
}

func (c *controller) run() (context.Context, context.CancelFunc) {
	if !c.runs.CompareAndSwap(false, true) {
		panic("twin controller can be run once")
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer close(c.done)
		defer c.s.Fini()
		defer cancel()
		defer c.closeAll()
		c.s.Clear()
		go func() {
			<-ctx.Done() // if someone called cancel() here or there...
			c.s.PostEvent(tcell.NewEventInterrupt(nil))
		}()
		c.onScreenResize()
		mousePressed := false
		for {
			c.onLoop()
			e := c.s.PollEvent()
			if e == nil {
				// we're done
				break
			}
			switch ev := e.(type) {
			case *tcell.EventKey:
				if ev.Key() == tcell.KeyCtrlC {
					return
				}
				c.onKeyPressed(c.root, ev)
			case *tcell.EventInterrupt:
				return
			case *tcell.EventResize:
				c.onScreenResize()
			case *resizeEvent:
				c.onResize(ev.comp)
			case *activateEvent:
				c.setActive(ev.comp)
			case *tcell.EventMouse:
				btns := ev.Buttons()
				clicks := btns & 255
				x, y := ev.Position()
				if !mousePressed && clicks != 0 {
					mousePressed = true
				}
				if mousePressed && clicks == 0 {
					c.onMouse(Point{x, y}, func(comp Component, p Point) {
						comp.OnMousePressed(p)
					})
					mousePressed = false
				}
				if btns&0xF00 != 0 {
					// translate the Up and Down to Left/Right if the modifiers are pressed
					if ev.Modifiers() != 0 {
						if btns == tcell.WheelUp {
							btns = tcell.WheelLeft
						} else if btns == tcell.WheelDown {
							btns = tcell.WheelRight
						}
					}
					c.onMouse(Point{x, y}, func(comp Component, p Point) {
						comp.OnMouseWheel(p, MouseWheel(btns))
					})
				}
			}
		}
	}()
	return ctx, cancel
}

func (c *controller) onKeyPressed(comp Component, ke *tcell.EventKey) bool {
	_, chld := comp.box().getActiveChild()
	if chld != nil {
		if c.onKeyPressed(chld, ke) {
			return true
		}
	}
	return comp.OnKeyPressed(ke)
}

type mouseF func(comp Component, p Point)

func (c *controller) onMouse(p Point, mf mouseF) {
	cc := newCanvas(c.root.Bounds().Size())
	c.onMouseComp(cc, c.root, p, mf)
}

func (c *controller) onMouseComp(cc *CanvasContext, comp Component, p Point, mf mouseF) bool {
	if !comp.IsVisible() {
		return false
	}
	b := comp.Bounds()
	tl := cc.physicalPointXY(b.TopLeft())
	b = b.Move(tl)
	if !b.Contains(p) {
		return false
	}
	b = comp.ChildrenCanvasBounds()
	b = b.Move(cc.physicalPointXY(b.TopLeft()))
	if b.Contains(p) {
		// p is in the chldrn bounds
		cc.pushRelativeRegion(comp.VirtualOffset(), b)
		defer cc.pop()
		children := comp.box().children()
		_, active := comp.box().getActiveChild()
		if active != nil && c.onMouseComp(cc, active, p, mf) {
			return true
		}
		for i := len(children) - 1; i >= 0; i-- {
			child := children[i]
			if child == active {
				continue
			}
			if c.onMouseComp(cc, child, p, mf) {
				return true
			}
		}
	}
	mf(comp, Point{X: p.X - tl.X, Y: p.Y - tl.Y})
	return c.setActive(comp)
}

func (c *controller) setActive(comp Component) bool {
	if !comp.CanBeFocused() || !comp.IsVisible() {
		return false
	}
	if comp.box().isActive() {
		return true
	}
	o := comp.box().owner
	for o != nil {
		if !o.IsVisible() || !o.CanBeFocused() {
			return false
		}
		o = o.box().owner
	}
	setActiveFalse(c.root)
	comp.box().setActive(true)
	o = comp.box().owner
	for o != nil {
		o.box().setActive(true)
		o = o.box().owner
	}
	return true
}

func (c *controller) onLoop() {
	c.lock.Lock()
	dirtySet := c.dirtySet
	if len(dirtySet) > 0 {
		c.dirtySet = make(map[Component]bool)
	}
	var deleted []Component
	c.deleteComponent(c.root, &deleted) // handle deleted comps
	c.lock.Unlock()

	if len(deleted) > 0 {
		for _, d := range deleted {
			d.box().closeActually()
		}
		dirtySet[c.root] = true
	}
	if len(dirtySet) == 0 {
		return
	}
	cc := newCanvas(c.root.Bounds().Size())
	c.draw(cc, c.root, false, dirtySet)
	c.s.Show()
}

func (c *controller) deleteComponent(comp Component, deleted *[]Component) bool {
	// check first if there is deleted childs
	closed := comp.box().isClosed()
	updateChildren := closed
	for _, child := range comp.box().children() {
		if closed {
			child.box().close()
		}
		if c.deleteComponent(child, deleted) {
			updateChildren = true
		}
	}
	if updateChildren {
		comp.box().removeClosedChildren()
	}
	if closed {
		*deleted = append(*deleted, comp)
	}
	return closed
}

func (c *controller) closeAll() {
	c.lock.Lock()
	c.root.close()
	var deleted []Component
	c.deleteComponent(c.root, &deleted)
	c.lock.Unlock()
	for _, d := range deleted {
		d.box().closeActually()
	}
}

func (c *controller) draw(cc *CanvasContext, comp Component, force bool, ds map[Component]bool) {
	if !comp.IsVisible() {
		return
	}
	if force || ds[comp] {
		cc.pushRelativeRegion(Point{0, 0}, comp.Bounds())
		comp.OnDraw(cc)
		cc.pop()
		force = true // redraw all chldrn then automatically
	}
	var active Component
	cc.pushRelativeRegion(comp.VirtualOffset(), comp.ChildrenCanvasBounds())
	chldrn := comp.box().children()
	for _, chld := range chldrn {
		if chld.box().isActive() {
			active = chld
		} else {
			c.draw(cc, chld, force, ds)
		}
	}
	if active != nil {
		c.draw(cc, active, force, ds)
	}
	cc.pop()
}

func (c *controller) reDrawNeeded(comp Component, event tcell.Event) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if comp.box().owner != nil {
		c.dirtySet[comp.box().owner] = true
	} else {
		c.dirtySet[comp] = true
	}

	if len(c.dirtySet) > 1 {
		return
	}
	c.s.PostEvent(event) // kick the loop to wake up
}

func (c *controller) onScreenResize() {
	w, h := c.s.Size()
	sz := Size{Width: w, Height: h}
	curSz := c.root.Bounds().Size()
	if sz == curSz {
		return
	}
	c.root.bounds.Store(Rectangle{X: 0, Y: 0, Width: w, Height: h})
	c.onResize(c.root)
	c.reDrawNeeded(c.root, &tcell.EventTime{})
}

func (c *controller) onResize(comp Component) {
	for _, chld := range comp.box().children() {
		chld.OnOwnerResized()
	}
}

// Call the event that a component is resized
func (c *controller) resize(comp Component) {
	c.reDrawNeeded(comp, &resizeEvent{comp: comp})
}
