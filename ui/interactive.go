package ui

import (
	"errors"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var (
	berr         = errors.New("Couldn't Get Base Selection in Interactive View")
	defaultfg    = tcell.ColorGreen
	defaultbg    = tcell.ColorDefault
	Defaultstyle = tcell.StyleDefault.
			Foreground(defaultfg).
			Background(defaultbg)
)

type Content struct {
	Content string
	Style   tcell.Style
}

type _range struct {
	Start int
	End   int
}

type interactiveView struct {
	visual  bool
	vrange  *_range
	baseSel int
	Table   *tview.Table
}

func NewInteractiveView() *interactiveView {
	view := tview.NewTable()
	view.SetSelectable(true, false)
	view.SetBackgroundColor(tcell.ColorDefault)

	i := &interactiveView{
		Table:  view,
		vrange: &_range{},
		visual: false,
	}

	view.SetDrawFunc(func(s tcell.Screen,
		x, y, width, height int) (int, int, int, int) {
		i.update()
		return i.Table.GetInnerRect()
	})
	view.SetInputCapture(i.capture)
	return i
}

func (i *interactiveView) SelectionHandler(selrow int) {
	if i.visual {
		i.toggleVisualMode()
	}
	if GetCurrentView().ContextHandler() != nil {
		GetCurrentView().ContextHandler()(i.vrange.Start, i.vrange.End, selrow)
	}
}

func (i *interactiveView) exitVisualMode() {
	if i.vrange.Start < i.baseSel {
		i.Table.Select(i.vrange.Start, -1)
	} else if i.vrange.End > i.baseSel {
		i.Table.Select(i.vrange.End, -1)
	}
	i.baseSel = -1
}

func (i *interactiveView) enterVisualMode() {
	row, _ := i.Table.GetSelection()
	i.baseSel = row
	i.vrange.Start, i.vrange.End = row, row
}

func (i *interactiveView) toggleVisualMode() {
	if i.visual {
		i.exitVisualMode()
	} else if !i.visual {
		i.enterVisualMode()
	}
	i.visual = !i.visual
}

func (i *interactiveView) getHandler(s string) func(e *tcell.EventKey) *tcell.EventKey {
	vr := i.vrange
	check := func() {
		if vr.Start <= -1 {
			vr.Start = 0
		}
		if vr.End <= -1 {
			vr.End = 0
		}
		if vr.End >= i.Table.GetRowCount() {
			vr.End = i.Table.GetRowCount() - 1
		}
		if vr.Start >= i.Table.GetRowCount() {
			vr.Start = i.Table.GetRowCount() - 1
		}
	}
	funcMap := map[string]func(e *tcell.EventKey) *tcell.EventKey{
		"up": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				check()
				if vr.End > i.baseSel {
					vr.End--
				} else if vr.Start <= i.baseSel {
					vr.Start--
				}
				if i.baseSel == -1 {
					panic(berr)
				}
				return nil
			}
			return e
		},
		"down": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				check()
				if vr.Start < i.baseSel {
					vr.Start++
				} else if vr.Start == i.baseSel {
					vr.End++
				}
				if i.baseSel == -1 {
					panic(berr)
				}
				return nil
			}
			return e
		},
		"exitvisual": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				i.exitVisualMode()
				i.visual = false
				return nil
			}
			return e
		},
		"top": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				i.vrange.Start = 0
				i.vrange.End = i.baseSel
				i.Table.ScrollToBeginning()
				return nil
			}
			return e
		},
		"bottom": func(e *tcell.EventKey) *tcell.EventKey {
			if i.visual {
				i.vrange.Start = i.baseSel
				i.vrange.End = i.Table.GetRowCount() - 1
				i.Table.ScrollToEnd()
				return nil
			}
			return e
		},
		"openCtx": func(e *tcell.EventKey) *tcell.EventKey {
			if GetCurrentView().ContextOpener() != nil {
				GetCurrentView().ContextOpener()(Ui.Root, Ui.Main.SelectionHandler)
				return nil
			}
			return e
		},
	}
	if val, ok := funcMap[s]; ok {
		return val
	} else {
		return nil
	}
}

func (i *interactiveView) capture(e *tcell.EventKey) *tcell.EventKey {
	switch e.Rune() {
	case 'j':
		{
			return i.getHandler("down")(e)
		}
	case 'k':
		{
			return i.getHandler("up")(e)
		}
	case 'v':
		{
			i.toggleVisualMode()
			return nil
		}
	case 'g':
		{
			return i.getHandler("top")(e)
		}
	case 'G':
		{
			return i.getHandler("bottom")(e)
		}
	case GetCurrentView().ContextKey():
		{
			return i.getHandler("openCtx")(e)
		}
	default:
		{
			if e.Key() == tcell.KeyEscape {
				return i.getHandler("exitvisual")(e)
			} else if GetCurrentView().ExternalInputCapture() != nil {
				return GetCurrentView().ExternalInputCapture()(e)
			}
			return e
		}
	}
}

func GetCell(text string, st tcell.Style) *tview.TableCell {
	return tview.NewTableCell(text).
		SetAlign(tview.AlignLeft).
		SetStyle(st)
}

func (i *interactiveView) update() {
	i.Table.Clear()
	if GetCurrentView().Content() != nil {
		s := GetCurrentView().Content()()
		_, _, w, _ := i.Table.GetInnerRect()
		for x := range s {
			b := ""
			fg := tcell.ColorDefault
			if i.visual && (x >= i.vrange.Start && x <= i.vrange.End) {
				b = "[blue::]█[::]"
				fg = tcell.ColorBlue
			}
			n := len(s[x])
			i.Table.SetCell(x, 0,
				GetCell(b, tcell.StyleDefault.Foreground(fg).Background(fg)))
			for y := range s[x] {
				i.Table.SetCell(x, y+1,
					GetCell(s[x][y].Content, s[x][y].Style).SetMaxWidth(w/n).SetExpansion(1))
			}
		}
	}
}