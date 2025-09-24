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

func newCBox(owner twin.Container, color tcell.Color) *CBox {
	cb := new(CBox)
	_ = cb.Init(owner, cb, tcell.StyleDefault.Background(color).Foreground(tcell.ColorWhite),
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
	redB := newCBox(twin.Root(), tcell.ColorRed)
	redB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	redB.SetBounds(twin.Rectangle{X: 10, Y: 10, Width: 50, Height: 20})

	blueB := newCBox(redB, tcell.ColorBlue)
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 10, Y: 10, Width: 50, Height: 20})
	<-ctx.Done()
	<-twin.Done()
}
