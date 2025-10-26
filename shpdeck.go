package shpdeck

import (
	"fmt"
	"io"
	"strings"

	"github.com/jonas-p/go-shp"
)

type Geometry struct {
	Xmin, Xmax, Ymin, Ymax, Latmin, Latmax, Longmin, Longmax float64
}

type Config struct {
	maptype   string
	color     string
	shapesize float64
}

// types used from go-shp
type Point shp.Point
type Polygon shp.Polygon
type PolyLine shp.PolyLine
type MultiPoint shp.MultiPoint

const (
	linefmt = "<line xp1=\"%.7f\" yp1=\"%.7f\" xp2=\"%.7f\" yp2=\"%.7f\" color=%q opacity=%q sp=\"%.3f\"/>\n"
	dotfmt  = "<ellipse xp=\"%.7f\" yp=\"%.7f\" hr=\"100\"color=%q opacity=%q wp=\"%.3f\"/>\n"
)

// vmap maps one interval to another
func vmap(value float64, low1 float64, high1 float64, low2 float64, high2 float64) float64 {
	return low2 + (high2-low2)*(value-low1)/(high1-low1)
}

// colorop makes a color and optional opacity in the form of name:op
func colorop(color string) (string, string) {
	ci := strings.Index(color, ":")
	op := "100"
	if ci > 0 && ci < len(color) {
		op = color[ci+1:]
		color = color[0:ci]
	}
	return color, op
}

// deckpolygon makes deck markup for a polygon given x, y coordinates slices
func deckpolygon(w io.Writer, x, y []float64, color string) {
	nc := len(x)
	//fmt.Fprintf(os.Stderr, "xlen=%03d\n\n", nc)
	if nc < 3 || nc != len(y) {
		return
	}
	fill, op := colorop(color)
	end := nc - 1
	fmt.Fprintf(w, "<polygon color=%q opacity=%q xc=\"%.5f", fill, op, x[0])
	for i := 1; i < nc; i++ {
		fmt.Fprintf(w, " %.5f", x[i])
	}
	fmt.Fprintf(w, " %.5f\" ", x[end])
	fmt.Fprintf(w, "yc=\"%.5f", y[0])
	for i := 1; i < nc; i++ {
		fmt.Fprintf(w, " %.5f", y[i])
	}
	fmt.Fprintf(w, " %.5f\"/>\n", y[end])
}

// deckdot makes a series of circles in deck markup from a set of (x,y) coordinates
func deckdot(w io.Writer, x, y []float64, color string, size float64) {
	fill, op := colorop(color)
	for i := range len(x) {
		fmt.Fprintf(w, dotfmt, x[i], y[i], fill, op, size)
	}
}

// deckpolyline makes a series of lines in deck markup from a set of (x,y) coordinates
func deckpolyline(w io.Writer, x, y []float64, color string, size float64) {
	fill, op := colorop(color)
	lx := len(x)
	for i := 0; i < lx-1; i++ {
		fmt.Fprintf(w, linefmt, x[i], y[i], x[i+1], y[i+1], fill, op, size)
	}
	fmt.Fprintf(w, linefmt, x[0], y[0], x[lx-1], y[lx-1], fill, op, size)
}

// mapshape writes markup to the destination according to the specified shape
func mapshape(w io.Writer, x, y []float64, shape string, color string, size float64) {
	switch shape {
	case "p", "poly", "region", "polygon":
		deckpolygon(w, x, y, color)
	case "l", "line", "border":
		deckpolyline(w, x, y, color, size)
	case "d", "dot", "circle":
		deckdot(w, x, y, color, size)
	}
}

// Open is a wrapper of shp.Open
func Open(s string) (*shp.Reader, error) {
	return shp.Open(s)
}

// polygonCoords converts a set of coordinates and makes polygons
// the polygons are mapped from geographical coordinates to screen bounding box
// the coordinates are processed in the order specified by a vector that contains
// the coordinate indicies.
func PolygonCoords(dest io.Writer, poly *shp.Polygon, g Geometry, c Config) {
	// for every part...
	last := poly.NumParts - 1
	for i := range last {
		// index into each part, reading coordinates, and map to map geometries
		x := []float64{}
		y := []float64{}
		for j := poly.Parts[i]; j < poly.Parts[i+1]; j++ {
			x = append(x, vmap(poly.Points[j].X, g.Longmin, g.Longmax, g.Xmin, g.Xmax))
			y = append(y, vmap(poly.Points[j].Y, g.Latmin, g.Latmax, g.Ymin, g.Ymax))
		}
		mapshape(dest, x, y, c.maptype, c.color, c.shapesize)
	}
	// process the last part
	x := []float64{}
	y := []float64{}
	for k := poly.Parts[last]; k < poly.NumPoints; k++ {
		x = append(x, vmap(poly.Points[k].X, g.Longmin, g.Longmax, g.Xmin, g.Xmax))
		y = append(y, vmap(poly.Points[k].Y, g.Latmin, g.Latmax, g.Ymin, g.Ymax))
	}
	mapshape(dest, x, y, c.maptype, c.color, c.shapesize)
}

// polygonCoords converts a set of coordinates and makes polylines
// the polylines are mapped from geographical coordinates to screen bounding box
// the coordinates are processed in the order specified by a vector that contains
// the coordinate indicies.
func PolylineCoords(dest io.Writer, poly *shp.PolyLine, g Geometry, c Config) {
	// for every part...
	last := poly.NumParts - 1
	for i := range last {
		// index into each part, reading coordinates, and map to map geometries
		x := []float64{}
		y := []float64{}
		for j := poly.Parts[i]; j < poly.Parts[i+1]; j++ {
			x = append(x, vmap(poly.Points[j].X, g.Longmin, g.Longmax, g.Xmin, g.Xmax))
			y = append(y, vmap(poly.Points[j].Y, g.Latmin, g.Latmax, g.Ymin, g.Ymax))
		}
		mapshape(dest, x, y, c.maptype, c.color, c.shapesize)
	}
	// process the last part
	x := []float64{}
	y := []float64{}
	for k := poly.Parts[last]; k < poly.NumPoints; k++ {
		x = append(x, vmap(poly.Points[k].X, g.Longmin, g.Longmax, g.Xmin, g.Xmax))
		y = append(y, vmap(poly.Points[k].Y, g.Latmin, g.Latmax, g.Ymin, g.Ymax))
	}
	mapshape(dest, x, y, c.maptype, c.color, c.shapesize)
}

// multipointCoords converts a set of coordinates and makes circles for each coordinate.
// the coordinates are mapped from geographical coordinates to screen bounding box
func MultipointCoords(dest io.Writer, mp *shp.MultiPoint, g Geometry, c Config) {
	x := []float64{}
	y := []float64{}
	for i := int32(0); i < mp.NumPoints; i++ {
		x = append(x, vmap(mp.Points[i].X, g.Longmin, g.Longmax, g.Xmin, g.Xmax))
		y = append(y, vmap(mp.Points[i].Y, g.Latmin, g.Latmax, g.Ymin, g.Ymax))
	}
	mapshape(dest, x, y, "dot", c.color, c.shapesize)
}

// pointCoords places a circle at a coordinate.
// the coordinates are mapped from geographical coordinates to screen bounding box.
func PointCoords(dest io.Writer, p *shp.Point, g Geometry, c Config) {
	x := vmap(p.X, g.Longmin, g.Longmax, g.Xmin, g.Xmax)
	y := vmap(p.Y, g.Latmin, g.Latmax, g.Ymin, g.Ymax)
	fill, op := colorop(c.color)
	fmt.Fprintf(dest, dotfmt, x, y, fill, op, c.shapesize)
}
