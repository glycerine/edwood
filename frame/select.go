package frame

import (
	"9fans.net/go/draw"
	"image"
)

// SetSelectionExtent sets the rune offsets of the selection maintained
// by the Frame. p0 and p1 must be values that could be returned by Charofpt.
// TODO(rjk): It is conceivable that we don't need this. It seems like an egregious
// abstraction violation that it exists.
func (f *Frame) SetSelectionExtent(p0, p1 int) {
	f.sp0, f.sp1 = p0, p1
}

// GetSelectionExtent returns the rune offsets of the selection maintained by
// the Frame.
func (f *Frame) GetSelectionExtent() (int, int) {
	return f.sp0, f.sp1
}

func region(a, b int) int {
	if a < b {
		return -1
	}
	if a == b {
		return 0
	}
	return 1
}

// Select takes ownership of the mouse channel to update the selection
// so long as a button is down in downevent. Selection stops when the
// staring point buttondown is altered. getmorelines is a callback provided
// by the caller to provide n additional lines on demand to the specified frame.
func (f *Frame) Select(mc *draw.Mousectl, downevent *draw.Mouse, getmorelines func(*Frame, int)) (int, int) {
	// log.Println("--- Select Start ---")
	// defer log.Println("--- Select End ---")

	omp := downevent.Point
	omb := downevent.Buttons

	// TODO(rjk): Figure out what Modified is really for.
	// Hypothesis: track if we have had inserts and removals during the selection loop.
	f.modified = false

	p0 := f.Charofpt(omp)
	p1 := p0
	f.DrawSel(f.Ptofchar(p0), p0, p1, true)

	reg := 0
	pin := 0

	for {
		me := <-mc.C
		mp := me.Point
		mb := me.Buttons

		scrled := false
		if mp.Y < f.Rect.Min.Y {
			getmorelines(f, -(f.Rect.Min.Y-mp.Y)/f.Font.DefaultHeight()-1)
			// As a result of scrolling, we will have called Insert. Insert will
			// remove the selection. But not put it back. But it will correct
			// P1 and P0 to reflect the insertion.
			// TODO(rjk): Add a unittest to prove this statement.
			p0 = f.sp1
			p1 = f.sp0
			scrled = true
		} else if mp.Y > f.Rect.Max.Y {
			getmorelines(f, (mp.Y-f.Rect.Max.Y)/f.Font.DefaultHeight()+1)
			p0 = f.sp1
			p1 = f.sp0
			scrled = true
		}
		if scrled {
			if reg != region(p1, p0) {
				tmp := p0
				p0 = p1
				p1 = tmp
			}
			reg = region(p1, p0)
		}

		q := f.Charofpt(mp)

		// log.Printf("select, before state table p0=%d p1=%d q=%d pin=%d", p0, p1, q, pin)
		switch {
		case p0 == p1 && q == p0:
			pin = 0
		case pin == 0 && q > p0:
			pin = 1
			p1 = q
		case pin == 0 && q < p0:
			pin = -1
			p0 = q
		case pin == -1 && q < p1:
			p0 = q
		case pin == -1 && q > p1: // We skipped equality.
			p0 = p1
			p1 = q
			pin = 1
		case pin == -1 && q == p1:
			p0 = q
			p1 = q
			pin = 0
		case pin == 1 && q > p0:
			p1 = q
		case pin == 1 && q == p0:
			pin = 0
			p0 = q
			p1 = q
		case pin == 1 && q < p0: // We skipped equality.
			pin = -1
			p1 = p0
			p0 = q
		}
		// log.Printf("select, after state table p0=%d p1=%d q=%d pin=%d", p0, p1, q, pin)

		f.DrawSel(f.Ptofchar(p0), p0, p1, true)

		if scrled {
			// TODO(rjk): Document why we need this call and what it's for.
			getmorelines(f, 0)
		}
		if err := f.Display.Flush(); err != nil {
			panic(err)
		}
		if omb != mb {
			break
		}
	}
	return f.sp0, f.sp1
}

func (f *Frame) SelectPaint(p0, p1 image.Point, col *draw.Image) {
	q0 := p0
	q1 := p1

	q0.Y += f.Font.DefaultHeight()
	q1.Y += f.Font.DefaultHeight()

	n := (p1.Y - p0.Y) / f.Font.DefaultHeight()
	if f.Background == nil {
		panic("Frame.SelectPaint B == nil")
	}
	if p0.Y == f.Rect.Max.Y {
		return
	}
	if n == 0 {
		f.Background.Draw(Rpt(p0, q1), col, nil, image.ZP)
	} else {
		if p0.X >= f.Rect.Max.X {
			p0.X = f.Rect.Max.X - 1
		}
		f.Background.Draw(image.Rect(p0.X, p0.Y, f.Rect.Max.X, q0.Y), col, nil, image.ZP)
		if n > 1 {
			f.Background.Draw(image.Rect(f.Rect.Min.X, q0.Y, f.Rect.Max.X, p1.Y), col, nil, image.ZP)
		}
		f.Background.Draw(image.Rect(f.Rect.Min.X, p1.Y, q1.X, q1.Y), col, nil, image.ZP)
	}
}
