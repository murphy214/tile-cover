# tile-cover

This package is a barebones implementation of a tile covering algorithm for simple line and polygon features. It could be much more complex but really I just needed something to work with tile-reduce and figured this is standalone enough to warant its own repo. 

# Usage 
```go
package main

import (
	"github.com/paulmach/go.geojson"
	"io/ioutil"
	"fmt"
	"github.com/murphy214/tile-cover"
	m "github.com/murphy214/mercantile"

)

func main() {
	bytevals,_ := ioutil.ReadFile("states.geojson")
	fc, _ := geojson.UnmarshalFeatureCollection(bytevals)
	feat := fc.Features[20]

	tileids := tile_cover.Tile_Cover(feat)
	for _,i := range tileids {
		fmt.Println(m.Tilestr(i))
	}
}
```

# Output 
![](https://user-images.githubusercontent.com/10904982/35519140-6876e2a0-04e1-11e8-9348-d87eb4614dcd.png)
