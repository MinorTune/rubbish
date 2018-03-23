package title

import "delaunay"
import "log"

type ObjectGroup struct {
	Name    string   `xml:"name,attr"`
	Objects []Object `xml:"object"`
}

func (o *ObjectGroup) InitWalkable() error {
	//Init Object.Polygon.Points
	objects_len := len(o.Objects)
	for i := 0; i < objects_len; i++ {
		if err := o.Objects[i].InitPoints(); err != nil {
			return err
		}
	}
	//Init Object.Layer
	for i := 0; i < objects_len; i++ {
		for j := 0; j < objects_len; j++ {
			if i == j {
				continue
			}
			c := 0
			if delaunay.RayCasting(o.Objects[i].Polygon.Points[c], o.Objects[j].Polygon.Points) != -1 {
				o.Objects[i].Layer = o.Objects[i].Layer + 1
				log.Println(o.Objects[i].ID, "in", o.Objects[j].ID)
			} else {
				log.Println(o.Objects[i].ID, "out", o.Objects[j].ID)
			}
		}
	}
	//Init Object.Polygon.Walkable

	for i := 0; i < objects_len; i++ {
		layer := o.Objects[i].Layer
		if layer&1 == 1 {
			continue
		}
		polygons := make([]delaunay.Polygon, 0)
		polygons = append(polygons, delaunay.Polygon{Data: o.Objects[i].Polygon.Points})
		for j := 0; j < objects_len; j++ {
			if i == j {
				continue
			}
			if o.Objects[j].Layer != layer+1 {
				continue
			}

			if delaunay.RayCasting(o.Objects[j].Polygon.Points[0], o.Objects[i].Polygon.Points) != -1 {
				polygons = append(polygons, delaunay.Polygon{Data: o.Objects[j].Polygon.Points})
			}
		}
		//初始化寻路网格

		log.Println("Walkable begin:", polygons)
		o.Objects[i].Polygon.Walkable = delaunay.CreateDealnay(polygons)
		log.Println("Walkable end:", o.Objects[i].Polygon.Walkable)

	}
	return nil
}
