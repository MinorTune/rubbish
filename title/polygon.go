package title

import "delaunay"
import "strings"
import "strconv"

type Polygon struct {
	Points_str string `xml:"points,attr"`
	Points     []delaunay.Point
	Walkable   []delaunay.Triangle
}

func (p *Polygon) Init() error {
	tmpstr := strings.Split(p.Points_str, " ")
	for i := 0; i < len(tmpstr); i++ {
		temp := strings.Split(tmpstr[i], ",")
		if len(temp) == 2 {
			x, err := strconv.ParseFloat(temp[0], 64)
			if err != nil {
				return err
			}
			y, err := strconv.ParseFloat(temp[1], 64)
			if err != nil {
				return err
			}
			p.Points = append(p.Points, delaunay.Point{X: int(x), Y: int(y)})
		}
	}
	return nil
}
func (p *Polygon) PointsAdd(p1 delaunay.Point) {
	for i := 0; i < len(p.Points); i++ {
		p.Points[i] = p.Points[i].Add(p1)
	}
}
