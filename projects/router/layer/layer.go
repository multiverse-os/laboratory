//package name
package layer

//package imports
import (
	"../mymath"
	"math"
)

/////////////////////////
//private structure/types
/////////////////////////

type dims struct {
	width  int
	height int
}

type point struct {
	x float32
	y float32
}

type line struct {
	p1     point
	p2     point
	radius float32
	gap    float32
}

type record struct {
	id   int
	line line
}

type aabb struct {
	minx int
	miny int
	maxx int
	maxy int
}

type bucket []*record
type buckets []bucket

//layer object
type layer struct {
	width   int
	height  int
	scale   float32
	buckets buckets
	test    int
}

//layer methods

/////////////////
//private methods
/////////////////

func newlayer(dims dims, s float32) *layer {
	l := layer{}
	l.init(dims, s)
	return &l
}

func (self *layer) init(dims dims, s float32) {
	self.width = dims.width
	self.height = dims.height
	self.scale = s
	self.buckets = make(buckets, (self.width * self.height), (self.width * self.height))
	for i := 0; i < (self.width * self.height); i++ {
		self.buckets[i] = bucket{}
	}
	self.test = 0
	return
}

func (self *layer) aabb(l *line) *aabb {
	x1, y1, x2, y2 := l.p1.x, l.p1.y, l.p2.x, l.p2.y
	if x1 > x2 {
		x1, x2 = x2, x1
	}
	if y1 > y2 {
		y1, y2 = y2, y1
	}
	r := l.radius + l.gap
	minx := int(math.Floor(float64((x1 - r) * self.scale)))
	miny := int(math.Floor(float64((y1 - r) * self.scale)))
	maxx := int(math.Ceil(float64((x2 + r) * self.scale)))
	maxy := int(math.Ceil(float64((y2 + r) * self.scale)))
	if minx < 0 {
		minx = 0
	}
	if miny < 0 {
		miny = 0
	}
	if maxx > self.width {
		maxx = self.width
	}
	if maxy > self.height {
		maxy = self.height
	}
	return &aabb{minx, miny, maxx, maxy}
}

func lines_equal(l1, l2 *line) bool {
	if l1.p1.x != l2.p1.x {
		return false
	}
	if l1.p1.y != l2.p1.y {
		return false
	}
	if l1.p2.x != l2.p2.x {
		return false
	}
	if l1.p2.y != l2.p2.y {
		return false
	}
	if l1.radius != l2.radius {
		return false
	}
	if l1.gap != l2.gap {
		return false
	}
	return true
}

func (self *layer) add_line(l *line) {
	new_record := record{0, *l}
	bb := self.aabb(l)
	for y := bb.miny; y < bb.maxy; y++ {
		for x := bb.minx; x < bb.maxx; x++ {
			b := y*self.width + x
			found := false
			for _, record := range self.buckets[b] {
				if lines_equal(&record.line, l) {
					found = true
					break
				}
			}
			if !found {
				self.buckets[b] = append(self.buckets[b], &new_record)
			}
		}
	}
}

func (self *layer) sub_line(l *line) {
	bb := self.aabb(l)
	for y := bb.miny; y < bb.maxy; y++ {
		for x := bb.minx; x < bb.maxx; x++ {
			b := y*self.width + x
			for i := len(self.buckets[b]) - 1; i >= 0; i-- {
				if lines_equal(&self.buckets[b][i].line, l) {
					self.buckets[b] = append(self.buckets[b][:i], self.buckets[b][i+1:]...)
				}
			}
		}
	}
}

func (self *layer) hit_line(l *line) bool {
	self.test += 1
	bb := self.aabb(l)
	l1_p1 := mymath.Point{l.p1.x, l.p1.y}
	l1_p2 := mymath.Point{l.p2.x, l.p2.y}
	l2_p1 := mymath.Point{0.0, 0.0}
	l2_p2 := mymath.Point{0.0, 0.0}
	for y := bb.miny; y < bb.maxy; y++ {
		for x := bb.minx; x < bb.maxx; x++ {
			for _, record := range self.buckets[y*self.width+x] {
				if record.id != self.test {
					record.id = self.test
					r := l.radius + record.line.radius
					if l.gap >= record.line.gap {
						r += l.gap
					} else {
						r += record.line.gap
					}
					l2_p1[0], l2_p1[1] = record.line.p1.x, record.line.p1.y
					l2_p2[0], l2_p2[1] = record.line.p2.x, record.line.p2.y
					if mymath.Collide_thick_lines_2d(&l1_p1, &l1_p2, &l2_p1, &l2_p2, r) {
						return true
					}
				}
			}
		}
	}
	return false
}

/////////////////////////
//public structures/types
/////////////////////////

type Dims struct {
	Width  int
	Height int
	Depth  int
}

//layers object
type Layers struct {
	depth  int
	layers []*layer
}

//layers methods

////////////////
//public methods
////////////////

func NewLayers(dims Dims, s float32) *Layers {
	l := Layers{}
	l.Init(dims, s)
	return &l
}

func (self *Layers) Init(dm Dims, s float32) *Layers {
	width := dm.Width
	height := dm.Height
	self.depth = dm.Depth
	self.layers = make([]*layer, self.depth, self.depth)
	for z := 0; z < self.depth; z++ {
		self.layers[z] = newlayer(dims{width, height}, s)
	}
	return self
}

func (self *Layers) Add_line(pp1, pp2 *mymath.Point, r, g float32) {
	p1 := *pp1
	p2 := *pp2
	z1 := int(p1[2])
	z2 := int(p2[2])
	if z1 > z2 {
		z1, z2 = z2, z1
	}
	line := line{point{p1[0], p1[1]}, point{p2[0], p2[1]}, r, g}
	for z := z1; z <= z2; z++ {
		self.layers[z].add_line(&line)
	}
}

func (self *Layers) Sub_line(pp1, pp2 *mymath.Point, r, g float32) {
	p1 := *pp1
	p2 := *pp2
	z1 := int(p1[2])
	z2 := int(p2[2])
	if z1 > z2 {
		z1, z2 = z2, z1
	}
	line := line{point{p1[0], p1[1]}, point{p2[0], p2[1]}, r, g}
	for z := z1; z <= z2; z++ {
		self.layers[z].sub_line(&line)
	}
}

func (self *Layers) Hit_line(pp1, pp2 *mymath.Point, r, g float32) bool {
	p1 := *pp1
	p2 := *pp2
	z1 := int(p1[2])
	z2 := int(p2[2])
	if z1 > z2 {
		z1, z2 = z2, z1
	}
	line := line{point{p1[0], p1[1]}, point{p2[0], p2[1]}, r, g}
	for z := z1; z <= z2; z++ {
		if self.layers[z].hit_line(&line) {
			return true
		}
	}
	return false
}
