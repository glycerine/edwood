package frame

import (
	"image"
	"unicode/utf8"
)

func (f *Frame) ptofcharptb(p int, pt image.Point, bn int) image.Point {
	var w int
	var r rune

	for _, b := range f.box[bn:] {
		pt = f.cklinewrap(pt, b)
		l := nrune(b)
		if p < l {
			if b.Nrune > 0 {
				for s := 0; s < len(b.Ptr) && p > 0; s += w {
					p--
					r, w = utf8.DecodeRune(b.Ptr[s:])
					pt.X += f.Font.BytesWidth(b.Ptr[s : s+w])
					if r == utf8.RuneError || pt.X > f.Rect.Max.X {
						panic("frptofchar")
					}
				}
			}
			break
		}
		p -= l
		pt = f.advance(pt, b)
	}

	return pt
}

func (f *Frame) Ptofchar(p int) image.Point {
	return f.ptofcharptb(p, f.Rect.Min, 0)
}

func (f *Frame) ptofcharnb(p int, nb int) image.Point {
	pt := image.Point{}
	pt = f.ptofcharptb(p, f.Rect.Min, 0)
	return pt
}

func (f *Frame) grid(p image.Point) image.Point {
	p.Y -= f.Rect.Min.Y
	p.Y -= p.Y % f.Font.DefaultHeight()
	p.Y += f.Rect.Min.Y
	if p.X > f.Rect.Max.X {
		p.X = f.Rect.Max.X
	}
	return p
}

func (f *Frame) Charofpt(pt image.Point) int {
	var w, bn int
	var p int
	var r rune

	pt = f.grid(pt)
	qt := f.Rect.Min

	for bn = 0; bn < len(f.box) && qt.Y < pt.Y; bn++ {
		b := f.box[bn]
		qt = f.cklinewrap(qt, b)
		if qt.Y >= pt.Y {
			break
		}
		qt = f.advance(qt, b)
		p += nrune(b)
	}

	for _, b := range f.box[bn:] {
		if qt.X > pt.X {
			break
		}
		qt = f.cklinewrap(qt, b)
		if qt.Y > pt.Y {
			break
		}
		if qt.X+b.Wid > pt.X {
			if b.Nrune < 0 {
				qt = f.advance(qt, b)
			} else {
				s := 0
				for ; s < len(b.Ptr); s += w {
					r, w = utf8.DecodeRune(b.Ptr[s:])
					if r == utf8.RuneError {
						panic("end of string in frcharofpt")
					}
					qt.X += f.Font.StringWidth(string(b.Ptr[s : s+1]))
					if qt.X > pt.X {
						break
					}
					p++
				}
			}
		} else {
			p += nrune(b)
			qt = f.advance(qt, b)
		}
	}
	return p
}
