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

func newCBox(owner twin.Component, style tcell.Style) *CBox {
	cb := new(CBox)
	_ = cb.Init(owner, cb, components.ScrollableBoxStyle{})
	l, _ := components.NewLabel(cb, components.LabelStyle{}.WithPureText("Label1\nHa\nkjlajsdfl"))
	l.SetBounds(twin.Rectangle{X: 1, Y: 1, Width: 7, Height: 2})
	return cb
}

func (cb *CBox) OnOwnerResized() {
	b := twin.Owner(cb).ChildrenCanvasBounds()
	bb := cb.Bounds()
	if b.Height < bb.Height+bb.Y {
		bb.Y = max(0, b.Height-bb.Height)
	}
	if b.Width < bb.Width+bb.X {
		bb.X = max(0, b.Width-bb.Width)
	}
	bb.Height = min(bb.Height, b.Height)
	bb.Width = min(bb.Width, b.Width)
	cb.SetBounds(bb)
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
	yb.SetVirtualSize(twin.Size{Width: 60, Height: 1000})
	yb.SetBounds(twin.Rectangle{X: 62, Y: 10, Width: 50, Height: 50})

	yb = newCBox(twin.Root(), tcell.StyleDefault.Background(tcell.ColorYellow).Foreground(tcell.ColorGrey))
	yb.SetBounds(twin.Rectangle{X: 10, Y: 31, Width: 50, Height: 20})

	blueB := newCBox(redB, tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack))
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 10, Y: 10, Width: 10, Height: 5})

	blueB = newCBox(redB, tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack))
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 21, Y: 5, Width: 10, Height: 5})

	blueB = newCBox(redB, tcell.StyleDefault.Background(tcell.ColorBlue).Foreground(tcell.ColorBlack))
	blueB.SetVirtualSize(twin.Size{Width: 100, Height: 100})
	blueB.SetBounds(twin.Rectangle{X: 8, Y: 12, Width: 10, Height: 5})

	mBox := newCBox(twin.NewModalPad(), tcell.StyleDefault.Background(tcell.ColorRed).Foreground(tcell.ColorWhite))
	mBox.SetVirtualSize(twin.Size{Width: 99, Height: 30})
	mBox.SetBounds(twin.Rectangle{X: 0, Y: 0, Width: 100, Height: 30})

	lb := &components.ListBox{}
	lb.Init(mBox, lb, components.ListBoxStyle{})
	lb.SetBounds(twin.Rectangle{X: 2, Y: 2, Width: 20, Height: 10})
	lb.SetVirtualSize(twin.Size{Width: 20, Height: 11})

	components.NewButton(mBox, components.ButtonStyle{}.WithText("[ Ok ]").
		WithRectangle(twin.Rectangle{X: 2, Y: 3, Width: 10, Height: 1}).
		WithOnEnter(func(b *components.Button) { twin.Close(mBox) }))

	components.NewButton(mBox, components.ButtonStyle{}.WithText("[ Cancel ]").
		WithRectangle(twin.Rectangle{X: 15, Y: 3, Width: 10, Height: 1}))

	twin.SetActive(mBox)
	<-ctx.Done()
	<-twin.Done()
}
