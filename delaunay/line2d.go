package delaunay

import "math"

type Line2D struct {
	A, B Point
}

func (l *Line2D) Eq(a Line2D) bool {
	return l.A.Eq(a.A) && l.B.Eq(a.B)
}

func (l *Line2D) Add(a Point) Line2D {
	return Line2D{A: l.A.Add(a), B: l.B.Add(a)}
}
func (l *Line2D) InterSection(l2 Line2D) (p Point, ok bool) {
	ok = false
	a, b, c, d := l.A, l.B, l2.A, l2.B
	rect1 := Rect(a, b)
	rect2 := Rect(c, d)
	if !rect1.Overlaps(rect2.Rectangle) {
		return
	}

	area_abc := (a.X-c.X)*(b.Y-c.Y) - (a.Y-c.Y)*(b.X-c.X)
	area_abd := (a.X-d.X)*(b.Y-d.Y) - (a.Y-d.Y)*(b.X-d.X)
	if area_abc*area_abd >= 0 {
		return
	}

	area_cda := (c.X-a.X)*(d.Y-a.Y) - (c.Y-a.Y)*(d.X-a.X)
	area_cdb := area_cda + area_abc - area_abd
	if area_cda*area_cdb >= 0 {
		return
	}

	t := area_cda / (area_abd - area_abc)
	dx := t * (b.X - a.X)
	dy := t * (b.Y - a.Y)

	p.X = a.X + dx
	p.Y = a.Y + dy
	ok = true
	return

}
func LineAngle(p1, p2, p3 Point) float64 {
	line1 := p1.Sub(p2)
	line2 := p3.Sub(p2)
	radian1 := math.Atan2(float64(line1.Y), float64(line1.X))
	radian2 := math.Atan2(float64(line2.Y), float64(line2.X))
	var r float64
	if radian1 > radian2 {
		r = radian1 - radian2
	} else {
		r = radian2 - radian1
		//return math.Pi*2 - (radian2 - radian1)
	}

	if r > math.Pi {
		return math.Pi*2 - r
	}
	return r
}
