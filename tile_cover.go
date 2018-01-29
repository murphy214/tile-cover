package tile_cover

import (
	"github.com/paulmach/go.geojson"
	f "github.com/murphy214/feature-map"
	m "github.com/murphy214/mercantile"
	"math"
)


// pip function
type Poly [][][]float64
func (cont Poly) Pip(p []float64) bool {
	// Cast ray from p.x towards the right
	intersections := 0
	for _,c := range cont {
		for i := range c {
			curr := c[i]
			ii := i + 1
			if ii == len(c) {
				ii = 0
			}
			next := c[ii]

			// Is the point out of the edge's bounding box?
			// bottom vertex is inclusive (belongs to edge), top vertex is
			// exclusive (not part of edge) -- i.e. p lies "slightly above
			// the ray"
			bottom, top := curr, next
			if bottom[1] > top[1] {
				bottom, top = top, bottom
			}
			if p[1] < bottom[1] || p[1] >= top[1] {
				continue
			}
			// Edge is from curr to next.

			if p[0] >= math.Max(curr[0], next[0]) ||
				next[1] == curr[1] {
				continue
			}

			// Find where the line intersects...
			xint := (p[1]-curr[1])*(next[0]-curr[0])/(next[1]-curr[1]) + curr[0]
			if curr[0] != next[0] && p[0] > xint {
				continue
			}

			intersections++
		}
	}

	return intersections%2 != 0
}

func Get_Corners(bds m.Extrema) ([]float64,[]float64,[]float64,[]float64) {
	wn := []float64{bds.W,bds.N}
	ws := []float64{bds.W,bds.S}
	en := []float64{bds.E,bds.N}
	es := []float64{bds.E,bds.S}
	return wn,ws,en,es
}

func Get_Min_Max(bds m.Extrema,zoom int) (int,int,int,int) {
	// getting corners
	wn,_,_,es := Get_Corners(bds)

	wnt := m.Tile(wn[0],wn[1],zoom)
	est := m.Tile(es[0],es[1],zoom)
	minx,maxx,miny,maxy := int(wnt.X),int(est.X),int(wnt.Y),int(est.Y)
	return minx,maxx,miny,maxy
}

// assembles a set of ranges
func Between(minval,maxval int) []int {
	current := minval 
	newlist := []int{current}
	for current < maxval {
		current += 1
		newlist = append(newlist,current)
	}
	return newlist
}

// getting tiles that cover the bounding box
func Get_BB_Tiles(bds m.Extrema,zoom int) []m.TileID {
	// the getting min and max
	minx,maxx, miny,maxy := Get_Min_Max(bds,zoom)
	
	// getting xs and ys
	xs,ys := Between(minx,maxx),Between(miny,maxy)

	// gertting all tiles
	newlist := []m.TileID{}
	for _,x := range xs {
		for _,y := range ys {
			newlist = append(newlist,m.TileID{int64(x),int64(y),uint64(zoom)})
		}
	}
	return newlist
}

// regular interpolation
func Interpolate(pt1,pt2 []float64,x float64) []float64 {
	slope := (pt2[1] - pt1[1]) / (pt2[0] - pt1[0])
	if math.Abs(slope) == math.Inf(1) {
		return []float64{pt1[0],(pt1[1]+pt2[1])/2.0}
	}
	return []float64{x,(x - pt1[0]) * slope + pt1[1]}

}

// interpolates points
func Interpolate_Pts(pt1,pt2 []float64,zoom int) []m.TileID {
	newtiles := []m.TileID{m.Tile(pt1[0],pt1[1],zoom),m.Tile(pt2[0],pt2[1],zoom)}
	var bds m.Extrema
	if pt1[0] >= pt2[0] {
		bds.E = pt1[0]
		bds.W = pt2[0]
	} else {
		bds.W = pt1[0]
		bds.E = pt2[0]
	}
	if pt1[1] >= pt2[1] {
		bds.N = pt1[1]
		bds.S = pt2[1]
	} else {
		bds.S = pt1[1]
		bds.N = pt2[1]
	}

	minx,maxx,miny,maxy := Get_Min_Max(bds,zoom)
	for _,x := range Between(minx,maxx) {
		tmpbds := m.Bounds(m.TileID{int64(x),int64(miny),uint64(zoom)})
		x1,x2 := tmpbds.W + .00000001,tmpbds.E - .00000001
		//x1,x2 := tmpbds.W ,tmpbds.E

		if bds.W <= x1 && bds.E >= x1 {
			pt := Interpolate(pt1,pt2,x1)
			newtiles = append(newtiles,m.Tile(pt[0],pt[1],zoom))
		}
		if bds.W <= x2 && bds.E >= x2 {
			pt := Interpolate(pt1,pt2,x2)
			newtiles = append(newtiles,m.Tile(pt[0],pt[1],zoom))
		}
	}
	pt1b := []float64{pt1[1],pt1[0]}
	pt2b := []float64{pt2[1],pt2[0]}

	for _,y := range Between(miny,maxy) {
		tmpbds := m.Bounds(m.TileID{int64(minx),int64(y),uint64(zoom)})
		y1,y2 := tmpbds.S + .00000001,tmpbds.N - .00000001
		//y1,y2 := tmpbds.S ,tmpbds.N

		if bds.S <= y1 && bds.N >= y1 {
			pt := Interpolate(pt1b,pt2b,y1)
			pt = []float64{pt[1],pt[0]}
			newtiles = append(newtiles,m.Tile(pt[0],pt[1],zoom))
		}
		if bds.S <= y2 && bds.N >= y2 {
			pt := Interpolate(pt1b,pt2b,y2)
			pt = []float64{pt[1],pt[0]}
			newtiles = append(newtiles,m.Tile(pt[0],pt[1],zoom))
		}
	}


	return newtiles
}

func Interpolate_Line(coords [][]float64,zoom int) []m.TileID {
	newtiles := []m.TileID{}
	oldpt := coords[0]
	for _,pt := range coords {
		newtiles = append(newtiles,Interpolate_Pts(oldpt,pt,zoom)...)
		oldpt = pt
	}

	newtiles2 := []m.TileID{}
	mymap := map[m.TileID]string{}
	for _,i := range newtiles {	
		_,ok := mymap[i]
		if ok == false {
			mymap[i] = ""
			newtiles2 = append(newtiles2,i)
		}
	}
	return newtiles2
}


func Fix_Polygon(coords [][][]float64) [][][]float64 {
	for i,cont := range coords {
		if (cont[0][0] != cont[len(cont) - 1][0]) && (cont[0][1] != cont[len(cont) - 1][1]) {
			cont = append(cont,cont[0])
			coords[i] = cont
		}
	}
	return coords
}


// covers a tile
func Tile_Cover(feat *geojson.Feature) []m.TileID {
	innertiles := []m.TileID{}
	zoom := 0

	if feat.Geometry.Type == "Polygon" {
		poly := Poly(feat.Geometry.Polygon)

		// getting bds
		bds := f.Get_Bds(feat.Geometry)
		boolval := true
		for boolval {
			tiles := Get_BB_Tiles(bds,zoom)
			for _,tile := range tiles {
				tmpbds := m.Bounds(tile)
				wn,ws,en,es := Get_Corners(tmpbds)
				wnpip,wspip,enpip,espip := poly.Pip(wn),poly.Pip(ws),poly.Pip(en),poly.Pip(es)
				mybool := wnpip && wspip && enpip && espip
				if mybool {
					boolval = false
					innertiles = append(innertiles,tile)
				}
			} 
			zoom += 1
		}
		zoom = zoom - 1
		for _,cont := range feat.Geometry.Polygon {

			innertiles = append(innertiles,Interpolate_Line(cont,zoom)...)
		}
		newtiles2 := []m.TileID{}
		mymap := map[m.TileID]string{}
		for _,i := range innertiles {	
			_,ok := mymap[i]
			if ok == false {
				mymap[i] = ""
				newtiles2 = append(newtiles2,i)
			}
		}
		innertiles = newtiles2
	} else if feat.Geometry.Type == "LineString" {
		boolval := true 
		for boolval {
			tempinnertiles := Interpolate_Line(feat.Geometry.LineString,zoom)
			if len(tempinnertiles) >= 5 || zoom >= 20 {
				boolval = false
				innertiles = tempinnertiles

			}
			zoom += 1
		}
	}
	return innertiles
}