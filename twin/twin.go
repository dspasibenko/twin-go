package twin

import (
	"context"
	"fmt"
	"github.com/gdamore/tcell/v2"
)

type Component interface {
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
	// For a Container, the function will be called before its children Draw().
	Draw(cc *CanvasContext)

	// IsVisible returns whether the component is visible or not
	IsVisible() bool

	// SetVisible sets the component visibility
	SetVisible(b bool)

	// Close closes the component and frees all resources
	Close()

	// SetBounds defines the Component position on the parent's component region and its size.
	SetBounds(r Rectangle)

	// Bounds returns the position and size of the component. The position is defined relative to the region
	// of the parent component.
	Bounds() Rectangle

	// VirtualOffset returns the offset of the drawing area. Normally returns (0,0), but is used for
	// scrollable areads
	VirtualOffset() Point

	// ChildrenCanvasBounds returns the area for drawing children. Can be useful if the component has some
	// borders or scroll bars
	ChildrenCanvasBounds() Rectangle

	// CanBeFocused returns whether the component may be focused or not
	CanBeFocused() bool

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

	// OnFocus is called when the component receives or loses the focus
	OnFocus(focused bool)

	// box is the private function to make the interface be implemented by the base Box
	// defined in the package.
	box() *Box
}

type Container interface {
	Component

	Children() []Component

	baseContainer() *BaseContainer
}

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
func Root() Container {
	return c.root
}

func Redraw(comp Component) {
	c.reDrawNeeded(comp)
}
