package twin

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
)

// Component is the interface which all the twin objects should implement. So as the
// Component returns box() all widgets should inherit Box struct
//
// All the functions may be overwritten by the wiget and return required value. All of the
// interface funcitons should be thread-safe and return desired values. The Box provides
// full implementation of the Component, so the widget may override only what is needed.
//
// The OnXXX() functions family are the notification methods and they will be always called
// by the twin from one go-routine only. They are not intended to be called not by the twin core
// so the behavior will be undefined.
type Component interface {
	// IsVisible returns whether the component is visible or not
	IsVisible() bool

	// Bounds returns the position and size of the component. The position is defined relative to the region
	// of the parent component.
	Bounds() Rectangle

	// VirtualOffset returns the offset of the drawing area. Normally returns (0,0), but is used for
	// scrollable areads
	VirtualOffset() Point

	// ChildrenCanvasBounds returns the area for drawing chldrn. Can be useful if the component has some
	// borders or scroll bars
	ChildrenCanvasBounds() Rectangle // maybe

	// CanBeFocused returns whether the component may be focused or not
	CanBeFocused() bool

	// Draw renders the component within the specified physical region, as defined by the cc parameter.
	// The implementation utilizes Raylib functions, such as rl.Rectangle(), to draw the component on
	// the display. The cc parameter specifies the position of the component on the physical display.
	//
	// By default, Draw() is invoked for the physical region where the component is defined. Raywin
	// uses scissors to constrain the drawing area. The implementation can adjust the drawing area
	// by calling rl.BeginScissorMode() if the region need to be changed.
	//
	// Raywin invokes Draw() for all visible components in each frame. A component is considered visible
	// if IsVisible() returns true and its Bounds() intersect with the visible region defined by its
	// parent Component (see Container).
	//
	// For a Container, the function will be called before its chldrn Draw().
	OnDraw(cc *CanvasContext)

	// OnKeyPressed is called when the component is notified about the key pressed.
	// If the processing should stop on the component, it must return true, if the key is not
	// handled and can be try by other component, it returns false
	OnKeyPressed(ke *tcell.EventKey) bool

	// OnOwnerResized called if the owner is resized
	OnOwnerResized()

	// OnMousePressed notifies about pressed mouse on the point p, disregarding virtual offset, but
	// on the component basis (0, 0) is the top left the component corner
	// The result is whether the component handled the action or skips it
	OnMousePressed(p Point) bool

	OnMouseWheel(p Point, wheel MouseWheel) bool

	// OnFocus is called when the component receives or loses the focus
	OnFocus(focused bool)

	// OnClosed is the final call for the object if closed
	OnClosed()

	// OnChildClosed is called when the component child is closed.
	OnChildClosed(comp Component)

	// box is the private function to make the interface be implemented by the base Box
	// defined in the package.
	box() *Box
}

type MouseWheel int

const (
	MouseWheelUp    = MouseWheel(tcell.WheelUp)
	MouseWheelDown  = MouseWheel(tcell.WheelDown)
	MouseWheelLeft  = MouseWheel(tcell.WheelLeft)
	MouseWheelRight = MouseWheel(tcell.WheelRight)
)

type Rectangle struct {
	X      int
	Y      int
	Width  int
	Height int
}

type Point struct {
	X int
	Y int
}

type Size struct {
	Width  int
	Height int
}

func (r Rectangle) Mend() Rectangle {
	return Rectangle{r.X, r.Y, max(0, r.Width), max(0, r.Height)}
}

func (r Rectangle) Move(p Point) Rectangle {
	r.X = p.X
	r.Y = p.Y
	return r
}

func (r Rectangle) Contains(p Point) bool {
	return p.X >= r.X && p.Y >= r.Y && p.X < r.X+r.Width && p.Y < r.Y+r.Height
}

func (r Rectangle) TopLeft() Point {
	return Point{X: r.X, Y: r.Y}
}
func (r Rectangle) TopRight() Point {
	return Point{X: r.X + r.Width - 1, Y: r.Y}
}

func (r Rectangle) BottomLeft() Point {
	return Point{X: r.X, Y: r.Y + r.Height - 1}
}

func (r Rectangle) BottomRight() Point {
	return Point{X: r.X + r.Width - 1, Y: r.Y + r.Height - 1}
}

func (r Rectangle) Size() Size {
	return Size{Width: r.Width, Height: r.Height}
}

func (r Rectangle) Normalized() Rectangle {
	return Rectangle{X: 0, Y: 0, Width: r.Width, Height: r.Height}
}

func (r Rectangle) String() string {
	return fmt.Sprintf("Rectangle{X:%d, Y:%d, Width:%d, Height:%d}", r.X, r.Y, r.Width, r.Height)
}

func (sz Size) Add(w, h int) Size {
	return Size{Width: max(0, sz.Width+w), Height: max(0, sz.Height+h)}
}

func (p Point) Add(x, y int) Point {
	return Point{X: p.X + x, Y: p.Y + y}
}

// Run runs the main cycle of twin and returns the context and the cancel() function
// twin will work until the context is closed. By default CTRL+C combination will close
// twin automatically
func Run() (context.Context, context.CancelFunc) {
	return c.run()
}

// Done returns the channel indicating that the all resources are completely released.
// It is worth to use when the main the context is closed, to wait the channel is
// closed before exisitng from the main process
func Done() <-chan struct{} {
	return c.done
}

// Root returns the root container for the all elements
func Root() Component {
	return c.root
}

// Redraw calls the comp redrawing forcedly
func Redraw(comp Component) {
	c.reDrawNeeded(comp, &tcell.EventTime{})
}

// Close allows to close the comp
func Close(comp Component) {
	comp.box().close()
}

// IsActive returns whether the comp is active or not
func IsActive(comp Component) bool {
	return comp.box().isActive()
}

// This returns the Component which was provided to Box.Init(). It can be different
// than the comp instance , if the comp type is an intermediate struct and This component was inherited
// from it.
func This(comp Component) Component {
	return comp.box().this
}

// Owner returns the owner of the comp
func Owner(comp Component) Component {
	return comp.box().owner
}

// NewModalPad creates a transparent component as an owner for a modal component. As soon as
// a component put on the modal pad, call SetActive() for it to make the component behavior as modal one
// The modal pad is closed by ESC button.
func NewModalPad() Component {
	m := &modalPad{}
	_ = m.Init(c.root, m)
	m.SetBounds(c.root.Bounds())
	return m
}

func SetActive(comp Component) {
	c.s.PostEvent(&activateEvent{comp: comp})
}
