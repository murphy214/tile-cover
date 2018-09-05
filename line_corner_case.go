package tilecover

import (
	m "github.com/murphy214/mercantile"
	"math"
)

func deg2rad(d float64) float64 {
	return d * math.Pi / 180.0
}

var EarthRadius = 3963.190592 // miles

// in miles
func Distance(pt1, pt2 []float64) float64 {
	dLat := deg2rad(pt2[1] - pt1[1])
	dLng := math.Abs(deg2rad(pt2[0] - pt1[0]))

	// yes trig functions
	dLat2Sin := math.Sin(dLat / 2)
	dLng2Sin := math.Sin(dLng / 2)
	a := dLat2Sin*dLat2Sin + math.Cos(deg2rad(pt1[1]))*math.Cos(deg2rad(pt2[1]))*dLng2Sin*dLng2Sin

	return 2.0 * EarthRadius * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
}

// fills values
func FillVals(val1, val2 int) []int {
	if val1 > val2 {
		dummy := val1
		val1 = val2
		val2 = dummy
	}

	if val1 == val2 || val2-1 == val1 {
		return []int{}
	}
	vals := make([]int, val2-val1-1)
	current := val1
	i := 0
	for current < val2-1 {
		current++
		vals[i] = current
		i++
	}
	return vals
}

// a semi normal iterpolation function
func Interpolate(pt1, pt2 []float64, val float64, ybool bool) []float64 {
	if !ybool {
		slope := (pt2[1] - pt1[1]) / (pt2[0] - pt1[0])
		return []float64{val, pt1[1] + (val-pt1[0])*slope}
	} else {
		pt1, pt2 = []float64{pt1[1], pt1[0]}, []float64{pt2[1], pt2[0]}
		slope := (pt2[1] - pt1[1]) / (pt2[0] - pt1[0])
		pt := []float64{val, pt1[1] + (val-pt1[0])*slope}
		return []float64{pt[1], pt[0]}
	}
	return []float64{}
}

// produces a dimension
func ProduceVals(vals []int, constant int, xbool bool, size int) []float64 {
	floatmap := map[float64]string{}
	if xbool {
		for _, val := range vals {
			tileid := m.TileID{int64(val), int64(constant), uint64(size)}
			bds := m.Bounds(tileid)
			floatmap[bds.W] = ""
			floatmap[bds.E] = ""
		}

	} else {
		for _, val := range vals {
			tileid := m.TileID{int64(constant), int64(val), uint64(size)}
			bds := m.Bounds(tileid)
			floatmap[bds.S] = ""
			floatmap[bds.N] = ""
		}
	}

	// creating float values
	floatvals := make([]float64, len(floatmap)*2)
	i := 0
	for k := range floatmap {
		floatvals[i*2] = k - .0000001
		floatvals[i*2+1] = k + .0000001
		i++
	}
	return floatvals
}

// fills the tiles between two given points
func BetweenTiles(pt1, pt2 []float64, size int) []m.TileID {
	tile1, tile2 := m.Tile(pt1[0], pt1[1], size), m.Tile(pt2[0], pt2[1], size)
	tiles := map[m.TileID]string{}
	tiles[tile1] = ""
	tiles[tile2] = ""
	// inequality thing
	if !((tile1.X == tile2.X) && (tile1.Y == tile2.Y)) {
		// getting xs
		xs := FillVals(int(tile1.X), int(tile2.X))
		xfloats := ProduceVals(xs, int(tile1.Y), true, size)

		// iterting through each x
		for _, xval := range xfloats {
			pt := Interpolate(pt1, pt2, xval, false)
			tiles[m.Tile(pt[0], pt[1], size)] = ""
		}

		// getting ys
		ys := FillVals(int(tile1.Y), int(tile2.Y))
		yfloats := ProduceVals(ys, int(tile1.X), false, size)

		// iterting through each x
		for _, yval := range yfloats {
			pt := Interpolate(pt1, pt2, yval, true)
			tiles[m.Tile(pt[0], pt[1], size)] = ""
		}

		// total tiles
		totaltiles := make([]m.TileID, len(tiles))
		i := 0
		for k := range tiles {
			totaltiles[i] = k
			i++
		}
		return totaltiles
	} else {
		return []m.TileID{tile1, tile2}
	}
	return []m.TileID{}
}
