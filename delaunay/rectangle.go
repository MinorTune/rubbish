package delaunay

import "image"

type Rectangle struct {
	image.Rectangle
}

func (r *Rectangle) ContainPoint(p Point) bool {
	if p.X < r.Min.X || p.X > r.Max.X {
		return false
	}
	if p.Y < r.Min.Y || p.Y > r.Max.Y {
		return false
	}
	return true
}

func CircleRect(c Circle) Rectangle {
	return Rectangle{image.Rect(c.P.X-int(c.R), c.P.Y-int(c.R), c.P.X+int(c.R), c.P.Y+int(c.R))}
}

func Rect(p1, p2 Point) Rectangle {
	return Rectangle{image.Rect(p1.X, p1.Y, p2.X, p2.Y)}
}
