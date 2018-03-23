package delaunay

type Polygon struct {
	Data []Point
}

/*
return value 0 on
return value 1 in
return value -1 out
*/
func RayCasting(p Point, poly []Point) int {
	px := p.X
	py := p.Y
	flag := false

	i := 0
	l := len(poly)
	j := l - 1
	for i < l {
		sx := poly[i].X
		sy := poly[i].Y
		tx := poly[j].X
		ty := poly[j].Y

		// 点与多边形顶点重合
		if (sx == px && sy == py) || (tx == px && ty == py) {
			return 0
		}

		// 判断线段两端点是否在射线两侧
		if (sy < py && ty >= py) || (sy >= py && ty < py) {
			// 线段上与射线 Y 坐标相同的点的 X 坐标
			x := sx + (py-sy)*(tx-sx)/(ty-sy)
			// 点在多边形的边上
			if x == px {
				return 0
			}
			// 射线穿过多边形的边界
			if x > px {
				flag = !flag
			}
		}
		j = i
		i++
	}

	// 射线穿过多边形边界的次数为奇数时点在多边形内
	if flag {
		return 1
	}
	return -1
}
