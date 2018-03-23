package title

import "delaunay"

type Object struct {
	ID      int     `xml:"id,attr"`
	X       float64 `xml:"x,attr"`
	Y       float64 `xml:"y,attr"`
	Layer   int
	Polygon Polygon `xml:"polygon"`
}

func (o *Object) InitPoints() error {
	err := o.Polygon.Init()
	if err != nil {
		return err
	}

	o.Polygon.PointsAdd(delaunay.Point{X: int(o.X), Y: int(o.Y)})
	return nil
}
