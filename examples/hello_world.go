package main

import (
	"github.com/dspasibenko/twin-go/twin"
	"github.com/dspasibenko/twin-go/twin/components"
	"github.com/gdamore/tcell/v2"
	"os"
	"syscall"
)

type CBox struct {
	components.ScrollableBox
}

func newCBox(owner twin.Container, style tcell.Style) *CBox {
	cb := new(CBox)
	_ = cb.Init(owner, cb, style,
		components.ScrollableBoxHasBorderBM|components.ScrollableBoxHasBothScrollsBM)
	return cb
}

func (cb *CBox) OnOwnerResized() {
	b := cb.Owner().ChildrenCanvasBounds()
	bb := cb.Bounds()
	if b.Height < bb.Height+bb.Y {
		bb.Y = max(0, b.Height-bb.Height)
		cb.SetBounds(bb)
	}
}

func main() {
	f, err := os.Create("output.log")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	syscall.Dup2(int(f.Fd()), int(os.Stdout.Fd()))
	ctx, _ := twin.Run()
	redB := newCBox(twin.Root(), tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
	redB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	redB.SetBounds(twin.Rectangle{X: 10, Y: 10, Width: 50, Height: 20})

	yb := newCBox(twin.Root(), tcell.StyleDefault.Background(tcell.ColorAquaMarine).Foreground(tcell.ColorBlack))
	yb.SetBounds(twin.Rectangle{X: 62, Y: 10, Width: 50, Height: 20})

	yb = newCBox(twin.Root(), tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorBlack))
	yb.SetBounds(twin.Rectangle{X: 10, Y: 31, Width: 50, Height: 20})

	blueB := newCBox(redB, tcell.StyleDefault.Background(tcell.ColorBlue))
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 10, Y: 10, Width: 10, Height: 5})

	blueB = newCBox(redB, tcell.StyleDefault.Background(tcell.ColorBlue))
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 21, Y: 5, Width: 10, Height: 5})

	blueB = newCBox(redB, tcell.StyleDefault.Background(tcell.ColorBlue))
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 8, Y: 12, Width: 10, Height: 5})
	<-ctx.Done()
	<-twin.Done()
}
