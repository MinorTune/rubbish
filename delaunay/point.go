package delaunay

type Point struct {
	X, Y int
}

func (p *Point) Add(a Point) Point {
	return Point{X: p.X + a.X, Y: p.Y + a.Y}
}
func (p *Point) Sub(a Point) Point {
	return Point{X: p.X - a.X, Y: p.Y - a.Y}
}
func (p *Point) Eq(a Point) bool {
	if p.X != a.X || p.Y != a.Y {
		return false
	}
	return true
}

/*
return 0 online
return <0 line right
return >0 line left
*/
func PointPAtLineAB(p, a, b Point) int {
	return (a.Y-b.Y)*p.X + (b.X-a.X)*p.Y + a.X*b.Y - b.X*a.Y
}
