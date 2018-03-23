package delaunay

import "log"
import "math"
import "errors"

type Circle struct {
	P Point
	R float64
}

func ThreePointsOnOneLine(p1, p2, p3 Point) bool {
	if p1.X == p2.X {
		if p2.X == p3.X {
			return true
		}
		return false
	}

	k1 := float64(p2.Y-p1.Y) / float64(p2.Y-p1.Y)
	k2 := float64(p3.Y-p2.Y) / float64(p3.X-p2.X)
	DIFF := 0.00000001
	if math.Abs(k1-k2) < DIFF {
		return true
	}
	return false
}

func CircumCircle(p1, p2, p3 Point) (cir Circle, err error) {
	if ThreePointsOnOneLine(p1, p2, p3) {
		err = errors.New("p1 p2 p3 online")
		log.Println(p1, p2, p3)
		return
	}

	defer func() {
		if e := recover(); e != nil {
			log.Println(p1, p2, p3)
			panic(e)
		}
	}()
	x1 := p1.X
	x2 := p2.X
	x3 := p3.X
	y1 := p1.Y
	y2 := p2.Y
	y3 := p3.Y

	a := math.Sqrt(float64((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2)))
	b := math.Sqrt(float64((x1-x3)*(x1-x3) + (y1-y3)*(y1-y3)))
	c := math.Sqrt(float64((x2-x3)*(x2-x3) + (y2-y3)*(y2-y3)))
	p := (a + b + c) / 2
	S := math.Sqrt(p * (p - a) * (p - b) * (p - c))
	cir.R = a * b * c / (4 * S)
	t1 := x1*x1 + y1*y1
	t2 := x2*x2 + y2*y2
	t3 := x3*x3 + y3*y3
	temp := x1*y2 + x2*y3 + x3*y1 - x1*y3 - x2*y1 - x3*y2
	x := (t2*y3 + t1*y2 + t3*y1 - t2*y1 - t3*y2 - t1*y3) / temp / 2
	y := (t3*x2 + t2*x1 + t1*x3 - t1*x2 - t2*x3 - t3*x1) / temp / 2
	cir.P.X = x
	cir.P.Y = y
	//log.Println(p1, p2, p3, "---", cir, err)
	return
}
