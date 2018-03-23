package delaunay

import "container/list"
import "log"

//import "fmt"

type delaunay struct {
	//所有顶点
	Points []Point
	//所有阻挡线段
	Lines []Line2D
}

func (d *delaunay) Init(polygons []Polygon) {
	for _, v := range polygons {
		length := len(v.Data)
		for i := 0; i < length-1; i++ {
			d.Points = append(d.Points, v.Data[i])
			d.Lines = append(d.Lines, Line2D{A: v.Data[i], B: v.Data[i+1]})
		}
		d.Points = append(d.Points, v.Data[length-1])
		d.Lines = append(d.Lines, Line2D{A: v.Data[length-1], B: v.Data[0]})
	}
	//log.Println("Points:", d.Points)
	log.Println("Line2Ds:", d.Lines)
}

func (d *delaunay) isVisiblePointOfLine(p Point, l Line2D) bool {
	if p.Eq(l.A) || p.Eq(l.B) {
		return false
	}
	if PointPAtLineAB(p, l.A, l.B) < 0 {
		//log.Println("not line right", p, l)
		//不在线的右边
		return false
	}
	if !d.isVisibleIn2Point(l.A, p) {
		return false
	}

	if !d.isVisibleIn2Point(l.B, p) {
		return false
	}
	return true
}

func (d *delaunay) isVisibleIn2Point(p1, p2 Point) bool {
	linep1p2 := Line2D{A: p1, B: p2}
	for _, v := range d.Lines {
		if p, ok := linep1p2.InterSection(v); ok {
			if !p.Eq(p1) && !p.Eq(p2) {
				return false
			}
		}
	}
	return true
}

func (d *delaunay) FindDT(dtline Line2D) (p3 Point, ok bool) {
	p1 := dtline.A
	p2 := dtline.B

	allVPoint := make([]Point, 0)

	for _, v := range d.Points {
		if d.isVisiblePointOfLine(v, dtline) {
			allVPoint = append(allVPoint, v)
		}
	}
	//log.Println(dtline.A, dtline.B, "allVPoint:", allVPoint)
	if len(allVPoint) == 0 {
		ok = false
		return
	}
	p3 = allVPoint[0]
	loop := true
	for loop {
		loop = false
		circle, err := CircumCircle(p1, p2, p3)
		if err != nil {
			log.Fatal(err)
		}
		rect := CircleRect(circle)
		angle132 := (LineAngle(p1, p3, p2))
		for _, p := range allVPoint {
			if p.Eq(p1) || p.Eq(p2) || p.Eq(p3) {
				continue
			}
			if !rect.ContainPoint(p) {
				continue
			}
			lineangle := (LineAngle(p1, p, p2))
			if lineangle > angle132 {
				//log.Println(p1, p3, p2, angle132, "<<<<<<<<", lineangle, p1, p, p2)
				p3 = p
				loop = true
				break
			} else {
				//log.Println(p1, p, p2, lineangle, "<<<<<<<<", angle132, p1, p3, p2)
			}
		}

	}
	//log.Fatal(p1, p2, p3)
	ok = true
	return
}

func (d *delaunay) ListRemoveOrPush(l *list.List, v Line2D) {
	for _, l := range d.Lines {
		if l.Eq(v) {
			return
		} else if l.Eq(Line2D{A: v.B, B: v.A}) {
			return
		}
	}
	for e := l.Front(); e != nil; e = e.Next() {
		p := e.Value.(Line2D)
		if v.Eq(p) {
			//			log.Println("popLine:", v)
			l.Remove(e)
			return
		} else if v.Eq(Line2D{A: p.B, B: p.A}) {
			l.Remove(e)
			return
		}
	}
	//	log.Println("pushLine:", v)
	l.PushBack(v)
}

func CreateDealnay(polygons []Polygon) (triangles []Triangle) {
	d := new(delaunay)
	d.Init(polygons)

	lineV := list.New()
	lineV.PushBack(d.Lines[0])

	for lineV.Len() != 0 {
		dtline := lineV.Remove(lineV.Front()).(Line2D)
		p3, ok := d.FindDT(dtline)
		if !ok {
			continue
		}

		log.Println("get dt point:", p3, ok, dtline, lineV.Len())
		line13 := Line2D{A: dtline.A, B: p3}
		line32 := Line2D{A: p3, B: dtline.B}

		//		var wt string
		//		fmt.Scan(&wt)
		triangles = append(triangles, Triangle{A: dtline.A, B: dtline.B, C: p3})
		d.ListRemoveOrPush(lineV, line13)
		d.ListRemoveOrPush(lineV, line32)
	}

	return
}
