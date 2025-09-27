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
	c.s.EnablePaste()
	c.s.EnableFocus()
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
		c.s.Clear()
		go func() {
			<-ctx.Done() // if someone called cancel() here or there...
			c.s.PostEvent(tcell.NewEventInterrupt(nil))
		}()
		c.resize()
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
				c.root.OnKeyPressed(ev)
			case *tcell.EventInterrupt:
				return
			case *tcell.EventResize:
				c.resize()
			case *tcell.EventMouse:
				btns := ev.Buttons()
				x, y := ev.Position()
				if !mousePressed && btns != tcell.ButtonNone {
					mousePressed = true
				}
				if mousePressed && btns == tcell.ButtonNone {
					c.onMousePressed(Point{x, y})
					mousePressed = false
				}
			}
		}
	}()
	return ctx, cancel
}

func (c *controller) onMousePressed(p Point) {
	cc := newCanvas(c.root.Bounds().Size())
	c.onMousePressedComp(cc, c.root, p)
}

func (c *controller) onMousePressedComp(cc *CanvasContext, comp Component, p Point) bool {
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
		// p is in the children bounds
		cc.pushRelativeRegion(comp.VirtualOffset(), b)
		defer cc.pop()
		if cont, ok := comp.(Container); ok {
			for _, child := range cont.Children() {
				if c.onMousePressedComp(cc, child, p) {
					return true
				}
			}
		}
	}
	return comp.OnMousePressed(Point{X: p.X - tl.X, Y: p.Y - tl.Y})
}

func (c *controller) onLoop() {
	c.lock.Lock()
	dirtySet := c.dirtySet
	c.dirtySet = make(map[Component]bool)
	c.lock.Unlock()
	if len(dirtySet) == 0 {
		return
	}
	cc := newCanvas(c.root.Bounds().Size())
	c.draw(cc, c.root, false, dirtySet)
	c.s.Show()
}

func (c *controller) draw(cc *CanvasContext, comp Component, force bool, ds map[Component]bool) {
	if force || ds[comp] {
		cc.pushRelativeRegion(Point{0, 0}, comp.Bounds())
		comp.Draw(cc)
		cc.pop()
		force = true // redraw all children then automatically
	}
	if bc, ok := comp.(Container); ok {
		cc.pushRelativeRegion(bc.VirtualOffset(), bc.ChildrenCanvasBounds())
		chldrn := bc.Children()
		for _, chld := range chldrn {
			c.draw(cc, chld, force, ds)
		}
		cc.pop()
	}
}

func (c *controller) reDrawNeeded(comp Component) {
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
	c.s.PostEvent(&tcell.EventTime{}) // kick the loop to wake up
}

func (c *controller) resize() {
	w, h := c.s.Size()
	sz := Size{Width: w, Height: h}
	curSz := c.root.Bounds().Size()
	if sz == curSz {
		return
	}
	c.root.SetBounds(Rectangle{X: 0, Y: 0, Width: w, Height: h})
}
