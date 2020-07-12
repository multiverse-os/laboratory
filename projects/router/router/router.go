// Copyright (C) 2014 Chris Hinsley.

//package name
package router

//package imports
import (
	"../layer"
	"../mymath"
	"fmt"
	"math/rand"
	"sort"
	"time"
)

/////////////////////////
//public structures/types
/////////////////////////

//dimensions of pcb board in grid points/layers
type Dims struct {
	Width  int
	Height int
	Depth  int
}

//grid point and collections
type Point struct {
	X int
	Y int
	Z int
}
type Vectors []*Point
type Vectorss []Vectors

//netlist structures
type Tpoint struct {
	X float32
	Y float32
	Z float32
}
type Path []*Tpoint
type Paths []Path

type Cord struct {
	X float32
	Y float32
}
type Cords []*Cord

type Terminal struct {
	Radius float32
	Gap    float32
	Term   Tpoint
	Shape  Cords
}
type Terminals []*Terminal

type Track struct {
	Radius float32
	Via    float32
	Gap    float32
	Terms  Terminals
}

type Output struct {
	Radius float32
	Via    float32
	Gap    float32
	Terms  Terminals
	Paths  Paths
}

//////////////////////////
//private structures/types
//////////////////////////

//sortable point
type sort_point struct {
	mark float32
	node *Point
}
type sort_points []*sort_point

type nets []*net

type aabb struct {
	minx int
	miny int
	maxx int
	maxy int
}

//for sorting nets
type by_group nets

func (s by_group) Len() int {
	return len(s)
}
func (s by_group) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s by_group) Less(i, j int) bool {
	if s[i].area == s[j].area {
		return s[i].radius > s[j].radius
	}
	return s[i].area < s[j].area
}

///////////////////////////
//private utility functions
///////////////////////////

//insert sort_point in ascending order
func insert_sort_point(nodes *sort_points, node *Point, mark float32) *sort_points {
	n := *nodes
	mn := sort_point{mark, node}
	for i := 0; i < len(n); i++ {
		if n[i].mark >= mark {
			n = append(n, nil)
			copy(n[i+1:], n[i:])
			n[i] = &mn
			return &n
		}
	}
	n = append(n, &mn)
	return &n
}

//pcb object
type Pcb struct {
	width                 int
	height                int
	depth                 int
	stride                int
	routing_flood_vectors *Vectorss
	routing_path_vectors  *Vectorss
	dfunc                 func(*mymath.Point, *mymath.Point) float32
	resolution            int
	verbosity             int
	quantization          int
	viascost              int
	layers                *layer.Layers
	netlist               nets
	nodes                 []int
	deform                map[Point]*mymath.Point
}

//pcb methods

////////////////
//public methods
////////////////

func NewPcb(dims *Dims, rfvs, rpvs *Vectorss, dfunc func(*mymath.Point, *mymath.Point) float32,
	res, verb, quant, viascost int) *Pcb {
	p := Pcb{}
	p.Init(dims, rfvs, rpvs, dfunc, res, verb, quant, viascost)
	return &p
}

func (self *Pcb) Init(dims *Dims, rfvs, rpvs *Vectorss,
	dfunc func(*mymath.Point, *mymath.Point) float32, res, verb, quant, viascost int) {
	self.width = dims.Width
	self.height = dims.Height
	self.depth = dims.Depth
	self.routing_flood_vectors = rfvs
	self.routing_path_vectors = rpvs
	self.dfunc = dfunc
	self.resolution = res
	self.verbosity = verb
	self.quantization = quant * res
	self.viascost = viascost
	self.layers = layer.NewLayers(layer.Dims{self.width * 3, self.height * 3, self.depth}, 3.0/float32(res))
	self.netlist = nil
	self.width *= res
	self.height *= res
	self.stride = self.width * self.height
	self.nodes = make([]int, self.stride*self.depth, self.stride*self.depth)
	self.deform = map[Point]*mymath.Point{}
}

func (self *Pcb) Copy() *Pcb {
	new_pcb := Pcb{}
	new_pcb.width = self.width
	new_pcb.height = self.height
	new_pcb.depth = self.depth
	new_pcb.stride = self.stride
	new_pcb.routing_flood_vectors = self.routing_flood_vectors
	new_pcb.routing_path_vectors = self.routing_path_vectors
	new_pcb.dfunc = self.dfunc
	new_pcb.resolution = self.resolution
	new_pcb.verbosity = self.verbosity
	new_pcb.quantization = self.quantization
	new_pcb.viascost = self.viascost
	new_pcb.layers = layer.NewLayers(layer.Dims{self.width * 3, self.height * 3, self.depth}, 3.0/float32(self.resolution))
	new_pcb.netlist = nil
	for _, net := range self.netlist {
		new_pcb.netlist = append(new_pcb.netlist, net.copy())
	}
	new_pcb.nodes = make([]int, self.stride*self.depth, self.stride*self.depth)
	new_pcb.deform = self.deform
	return &new_pcb
}

//add net
func (self *Pcb) Add_track(t *Track) {
	self.netlist = append(self.netlist, newnet(t.Terms, t.Radius, t.Via, t.Gap, *self))
}

//attempt to route board within time
func (self *Pcb) Route(timeout float64) bool {
	self.remove_netlist()
	self.unmark_distances()
	self.reset_areas()
	self.shuffle_netlist()
	sort.Sort(by_group(self.netlist))
	hoisted_nets := map[*net]bool{}
	index := 0
	start_time := time.Now()
	for index < len(self.netlist) {
		if self.netlist[index].route() {
			index += 1
		} else {
			if index == 0 {
				self.reset_areas()
				self.shuffle_netlist()
				sort.Sort(by_group(self.netlist))
				hoisted_nets = map[*net]bool{}
			} else {
				pos := 0
				self.netlist, pos = hoist_net(self.netlist, index)
				if (pos == index) || hoisted_nets[self.netlist[pos]] {
					if pos != 0 {
						self.netlist[pos].area = self.netlist[pos-1].area
						self.netlist, pos = hoist_net(self.netlist, pos)
					}
					delete(hoisted_nets, self.netlist[pos])
				} else {
					hoisted_nets[self.netlist[pos]] = true
				}
				for index > pos {
					self.netlist[index].remove()
					self.netlist[index].shuffle_topology()
					index -= 1
				}
			}
		}
		if time.Since(start_time).Seconds() > timeout {
			return false
		}
		if self.verbosity >= 1 {
			self.Print_netlist()
		}
	}
	return true
}

//cost of board in complexity terms
func (self *Pcb) Cost() int {
	sum := 0
	for _, net := range self.netlist {
		for _, path := range net.paths {
			sum += len(path)
		}
	}
	return sum
}

//increase area quantization
func (self *Pcb) Increase_quantization() {
	self.quantization++
}

//output dimensions of board for viewer app
func (self *Pcb) Print_pcb() {
	scale := 1.0 / float32(self.resolution)
	fmt.Print("[")
	fmt.Print(float32(self.width)*scale, ",")
	fmt.Print(float32(self.height)*scale, ",")
	fmt.Print(self.depth)
	fmt.Println("]")
}

//output netlist and paths of board for viewer app
func (self *Pcb) Print_netlist() {
	for _, net := range self.netlist {
		net.print_net()
	}
	fmt.Println("[]")
}

//output stats to screen
func (self *Pcb) Print_stats() {
	num_vias := 0
	num_terminals := 0
	num_nets := len(self.netlist)
	for _, net := range self.netlist {
		num_terminals += len(net.terminals)
		for _, path := range net.paths {
			p1 := *path[0]
			for _, node := range path[1:] {
				p0 := p1
				p1 = *node
				if p0.Z != p1.Z {
					num_vias++
				}
			}
		}
	}
	println("Number of Terminals:", num_terminals)
	println("Number of Nets:", num_nets)
	println("Number of Vias:", num_vias)
}

/////////////////
//private methods
/////////////////

//convert grid point to space point
func (self *Pcb) grid_to_space_point(p *Point) *mymath.Point {
	sp := self.deform[*p]
	if sp != nil {
		return sp
	}
	return &mymath.Point{float32(p.X), float32(p.Y), float32(p.Z)}
}

//set grid point to value
func (self *Pcb) set_node(node *Point, value int) {
	self.nodes[(self.stride*node.Z)+(node.Y*self.width)+node.X] = value
}

//get grid point value
func (self *Pcb) get_node(node *Point) int {
	return self.nodes[(self.stride*node.Z)+(node.Y*self.width)+node.X]
}

//generate all grid points surrounding point, that are not value 0
func (self *Pcb) all_marked(vectors *Vectorss, node *Point) *sort_points {
	vec := *vectors
	x, y, z := node.X, node.Y, node.Z
	yield := make(sort_points, 0, len(vec[z%2]))
	for _, v := range vec[z%2] {
		nx := x + v.X
		ny := y + v.Y
		nz := z + v.Z
		if (0 <= nx) && (nx < self.width) && (0 <= ny) && (ny < self.height) && (0 <= nz) && (nz < self.depth) {
			n := Point{nx, ny, nz}
			mark := self.get_node(&n)
			if mark != 0 {
				yield = append(yield, &sort_point{float32(mark), &n})
			}
		}
	}
	return &yield
}

//generate all grid points surrounding point, that are value 0
func (self *Pcb) all_not_marked(vectors *Vectorss, node *Point) *Vectors {
	vec := *vectors
	x, y, z := node.X, node.Y, node.Z
	yield := make(Vectors, 0, len(vec[z%2]))
	for _, v := range vec[z%2] {
		nx := x + v.X
		ny := y + v.Y
		nz := z + v.Z
		if (0 <= nx) && (nx < self.width) && (0 <= ny) && (ny < self.height) && (0 <= nz) && (nz < self.depth) {
			n := Point{nx, ny, nz}
			if self.get_node(&n) == 0 {
				yield = append(yield, &n)
			}
		}
	}
	return &yield
}

//generate all grid points surrounding point, that are nearer the goal point, sorted
func (self *Pcb) all_nearer_sorted(vectors *Vectorss, node, goal *Point,
	dfunc func(*mymath.Point, *mymath.Point) float32) *Vectors {
	gp := self.grid_to_space_point(goal)
	distance := float32(self.get_node(node))
	nodes := &sort_points{}
	for _, mn := range *self.all_marked(vectors, node) {
		if (distance - mn.mark) > 0 {
			mnp := self.grid_to_space_point(mn.node)
			nodes = insert_sort_point(nodes, mn.node, dfunc(mnp, gp))
		}
	}
	yield := make(Vectors, len(*nodes), len(*nodes))
	for i, node := range *nodes {
		yield[i] = node.node
	}
	return &yield
}

//generate all grid points surrounding point that are not shorting with an existing track
func (self *Pcb) all_not_shorting(gather *Vectors, node *Point, radius, gap float32) *Vectors {
	yield := make(Vectors, 0, 16)
	np := self.grid_to_space_point(node)
	for _, new_node := range *gather {
		nnp := self.grid_to_space_point(new_node)
		if !self.layers.Hit_line(np, nnp, radius, gap) {
			yield = append(yield, new_node)
		}
	}
	return &yield
}

//flood fill distances from starts till ends covered
func (self *Pcb) mark_distances(vectors *Vectorss, radius, via, gap float32, starts *map[Point]bool, ends *Vectors) {
	via_vectors := &Vectorss{
		Vectors{&Point{0, 0, -1}, &Point{0, 0, 1}},
		Vectors{&Point{0, 0, -1}, &Point{0, 0, 1}}}
	distance := 1
	nodes := *starts
	vias_nodes := map[int]*map[Point]bool{}
	for (len(nodes) > 0) || (len(vias_nodes) > 0) {
		for node := range nodes {
			self.set_node(&node, distance)
		}
		flag := true
		for _, node := range *ends {
			if self.get_node(node) == 0 {
				flag = false
				break
			}
		}
		if flag {
			break
		}
		new_nodes := map[Point]bool{}
		for node := range nodes {
			for _, new_node := range *self.all_not_shorting(self.all_not_marked(vectors, &node), &node, radius, gap) {
				new_nodes[*new_node] = true
			}
		}
		new_vias_nodes := map[Point]bool{}
		for node := range nodes {
			for _, new_vias_node := range *self.all_not_shorting(self.all_not_marked(via_vectors, &node), &node, via, gap) {
				new_vias_nodes[*new_vias_node] = true
			}
		}
		if len(new_vias_nodes) > 0 {
			vias_nodes[distance+self.viascost] = &new_vias_nodes
		}
		delay_nodes := vias_nodes[distance]
		if delay_nodes != nil {
			for vias_node := range *delay_nodes {
				if self.get_node(&vias_node) == 0 {
					new_nodes[vias_node] = true
				}
			}
			delete(vias_nodes, distance)
		}
		nodes = new_nodes
		distance++
	}
}

//set all grid values back to 0
func (self *Pcb) unmark_distances() {
	for i := 0; i < len(self.nodes); i++ {
		self.nodes[i] = 0
	}
}

//reset areas
func (self *Pcb) reset_areas() {
	for _, net := range self.netlist {
		net.area, net.bbox = aabb_terminals(net.terminals, self.quantization)
	}
}

//shuffle order of netlist
func (self *Pcb) shuffle_netlist() {
	for _, net := range self.netlist {
		net.shuffle_topology()
	}
	new_nets := make(nets, len(self.netlist), len(self.netlist))
	for i, r := range rand.Perm(len(self.netlist)) {
		new_nets[i] = self.netlist[r]
	}
	self.netlist = new_nets
}

//move net to top of area group
func hoist_net(ns nets, n int) (nets, int) {
	i := 0
	if n != 0 {
		for i = n; i >= 0; i-- {
			if ns[i].area < ns[n].area {
				break
			}
		}
		i++
		if n != i {
			net := ns[n]
			copy(ns[i+1:], ns[i:n])
			ns[i] = net
		}
	}
	return ns, i
}

func (self *Pcb) remove_netlist() {
	for _, net := range self.netlist {
		net.remove()
	}
}

//net object
type net struct {
	pcb       Pcb
	terminals Terminals
	radius    float32
	via       float32
	gap       float32
	area      int
	bbox      aabb
	paths     Vectorss
}

//net methods

/////////////////
//private methods
/////////////////

func newnet(terms Terminals, radius, via, gap float32, pcb Pcb) *net {
	n := net{}
	n.init(terms, radius, via, gap, pcb)
	return &n
}

//scale terminals for resolution of grid
func scale_terminals(terms Terminals, res int) Terminals {
	for i := 0; i < len(terms); i++ {
		terms[i].Radius *= float32(res)
		terms[i].Gap *= float32(res)
		terms[i].Term.X *= float32(res)
		terms[i].Term.Y *= float32(res)
		terms[i].Term.Z *= float32(res)
		for _, cord := range terms[i].Shape {
			cord.X *= float32(res)
			cord.Y *= float32(res)
		}
	}
	return terms
}

//aabb of terminals
func aabb_terminals(terms Terminals, quantization int) (int, aabb) {
	minx := (int(terms[0].Term.X) / quantization) * quantization
	miny := (int(terms[0].Term.Y) / quantization) * quantization
	maxx := ((int(terms[0].Term.X) + (quantization - 1)) / quantization) * quantization
	maxy := ((int(terms[0].Term.Y) + (quantization - 1)) / quantization) * quantization
	for i := 1; i < len(terms); i++ {
		tminx := (int(terms[i].Term.X) / quantization) * quantization
		tminy := (int(terms[i].Term.Y) / quantization) * quantization
		tmaxx := ((int(terms[i].Term.X) + (quantization - 1)) / quantization) * quantization
		tmaxy := ((int(terms[i].Term.Y) + (quantization - 1)) / quantization) * quantization
		if tminx < minx {
			minx = tminx
		}
		if tminy < miny {
			miny = tminy
		}
		if tmaxx > maxx {
			maxx = tmaxx
		}
		if tmaxy > maxy {
			maxy = tmaxy
		}
	}
	rec := aabb{minx, miny, maxx, maxy}
	return (maxx - minx) * (maxy - miny), rec
}

func (self *net) init(terms Terminals, radius, via, gap float32, pcb Pcb) {
	self.pcb = pcb
	self.radius = radius * float32(pcb.resolution)
	self.via = via * float32(pcb.resolution)
	self.gap = gap * float32(pcb.resolution)
	self.paths = make(Vectorss, 0)
	self.terminals = scale_terminals(terms, pcb.resolution)
	self.area, self.bbox = aabb_terminals(terms, pcb.quantization)
	self.remove()
	for _, term := range self.terminals {
		for z := 0; z < pcb.depth; z++ {
			p := Point{int(term.Term.X + 0.5), int(term.Term.Y + 0.5), z}
			sp := mymath.Point{term.Term.X, term.Term.Y, float32(z)}
			pcb.deform[p] = &sp
		}
	}
}

//copy terminals
func copy_terminals(terms Terminals) Terminals {
	new_terms := make(Terminals, len(terms), cap(terms))
	copy(new_terms, terms)
	return new_terms
}

func (self *net) copy() *net {
	new_net := net{}
	new_net.pcb = self.pcb
	new_net.radius = self.radius
	new_net.via = self.via
	new_net.gap = self.gap
	new_net.area = self.area
	new_net.terminals = copy_terminals(self.terminals)
	new_net.paths = self.optimise_paths(self.paths[:])
	return &new_net
}

//randomize order of terminals
func shuffle_terminals(terms Terminals) Terminals {
	new_terms := make(Terminals, len(terms), len(terms))
	for i, r := range rand.Perm(len(terms)) {
		new_terms[i] = terms[r]
	}
	return new_terms
}

func (self *net) shuffle_topology() {
	self.terminals = shuffle_terminals(self.terminals)
}

//add terminal entries to spacial cache
func (self *net) add_terminal_collision_lines() {
	for _, node := range self.terminals {
		r, g, x, y, shape := node.Radius, node.Gap, node.Term.X, node.Term.Y, node.Shape
		if len(shape) == 0 {
			self.pcb.layers.Add_line(&mymath.Point{x, y, 0}, &mymath.Point{x, y, float32(self.pcb.depth - 1)}, r, g)
		} else {
			for z := 0; z < self.pcb.depth; z++ {
				p1 := mymath.Point{x + shape[0].X, y + shape[0].Y, float32(z)}
				for i := 1; i < len(shape); i++ {
					p0 := p1
					p1 = mymath.Point{x + shape[i].X, y + shape[i].Y, float32(z)}
					self.pcb.layers.Add_line(&p0, &p1, r, g)
				}
			}
		}
	}
}

//remove terminal entries from spacial cache
func (self *net) sub_terminal_collision_lines() {
	for _, node := range self.terminals {
		r, g, x, y, shape := node.Radius, node.Gap, node.Term.X, node.Term.Y, node.Shape
		if len(shape) == 0 {
			self.pcb.layers.Sub_line(&mymath.Point{x, y, 0}, &mymath.Point{x, y, float32(self.pcb.depth - 1)}, r, g)
		} else {
			for z := 0; z < self.pcb.depth; z++ {
				p1 := mymath.Point{x + shape[0].X, y + shape[0].Y, float32(z)}
				for i := 1; i < len(shape); i++ {
					p0 := p1
					p1 = mymath.Point{x + shape[i].X, y + shape[i].Y, float32(z)}
					self.pcb.layers.Sub_line(&p0, &p1, r, g)
				}
			}
		}
	}
}

//add paths entries to spacial cache
func (self *net) add_paths_collision_lines() {
	for _, path := range self.paths {
		p1 := self.pcb.grid_to_space_point(path[0])
		for i := 1; i < len(path); i++ {
			p0 := p1
			p1 = self.pcb.grid_to_space_point(path[i])
			if path[i-1].Z != path[i].Z {
				//via direction
				self.pcb.layers.Add_line(p0, p1, self.via, self.gap)
			} else {
				//not via direction
				self.pcb.layers.Add_line(p0, p1, self.radius, self.gap)
			}
		}
	}
}

//remove paths entries from spacial cache
func (self *net) sub_paths_collision_lines() {
	for _, path := range self.paths {
		p1 := self.pcb.grid_to_space_point(path[0])
		for i := 1; i < len(path); i++ {
			p0 := p1
			p1 = self.pcb.grid_to_space_point(path[i])
			if path[i-1].Z != path[i].Z {
				//via direction
				self.pcb.layers.Sub_line(p0, p1, self.via, self.gap)
			} else {
				//not via direction
				self.pcb.layers.Sub_line(p0, p1, self.radius, self.gap)
			}
		}
	}
}

//remove net entries from spacial grid
func (self *net) remove() {
	self.sub_paths_collision_lines()
	self.sub_terminal_collision_lines()
	self.paths = nil
	self.add_terminal_collision_lines()
}

//remove redundant points from paths
func (self *net) optimise_paths(paths Vectorss) Vectorss {
	opt_paths := make(Vectorss, 0)
	for _, path := range paths {
		opt_path := make(Vectors, 0)
		d := &mymath.Point{0, 0, 0}
		p1 := self.pcb.grid_to_space_point(path[0])
		for i := 1; i < len(path); i++ {
			p0 := p1
			p1 = self.pcb.grid_to_space_point(path[i])
			d1 := mymath.Norm_3d(mymath.Sub_3d(p1, p0))
			if !mymath.Equal_3d(d1, d) {
				opt_path = append(opt_path, path[i-1])
				d = d1
			}
		}
		opt_path = append(opt_path, path[len(path)-1])
		opt_paths = append(opt_paths, opt_path)
	}
	return opt_paths
}

//backtrack path from ends to starts
func (self *net) backtrack_path(vis *map[Point]bool, end *Point, radius, via, gap float32) (Vectors, bool) {
	via_vectors := &Vectorss{
		Vectors{&Point{0, 0, -1}, &Point{0, 0, 1}},
		Vectors{&Point{0, 0, -1}, &Point{0, 0, 1}}}
	visited := *vis
	path := Vectors{end}
	dv := &mymath.Point{0, 0, 0}
	for {
		path_node := path[len(path)-1]
		if visited[*path_node] {
			//found existing track
			return path, true
		}
		nearer_nodes := make(Vectors, 0)
		for _, node := range *self.pcb.all_not_shorting(
			self.pcb.all_nearer_sorted(self.pcb.routing_path_vectors, path_node, end, self.pcb.dfunc),
			path_node, radius, gap) {
			nearer_nodes = append(nearer_nodes, node)
		}
		for _, node := range *self.pcb.all_not_shorting(
			self.pcb.all_nearer_sorted(via_vectors, path_node, end, self.pcb.dfunc),
			path_node, via, gap) {
			nearer_nodes = append(nearer_nodes, node)
		}
		if len(nearer_nodes) == 0 {
			//no nearer nodes
			return path, false
		}
		next_node := nearer_nodes[0]
		dv2 := mymath.Norm_3d(self.pcb.grid_to_space_point(path_node))
		if !visited[*next_node] {
			for i := 1; i < len(nearer_nodes); i++ {
				node := nearer_nodes[i]
				dv1 := mymath.Norm_3d(self.pcb.grid_to_space_point(node))
				if mymath.Equal_3d(dv, mymath.Sub_3d(dv1, dv2)) {
					next_node = node
				}
			}
		}
		dv1 := mymath.Norm_3d(self.pcb.grid_to_space_point(next_node))
		dv = mymath.Norm_3d(mymath.Sub_3d(dv1, dv2))
		path = append(path, next_node)
	}
}

//attempt to route this net on the current boards state
func (self *net) route() bool {
	if self.radius == 0.0 {
		//unused terminals track !
		return true
	}
	self.paths = make(Vectorss, 0)
	self.sub_terminal_collision_lines()
	visited := map[Point]bool{}
	for index := 1; index < len(self.terminals); index++ {
		for z := 0; z < self.pcb.depth; z++ {
			x, y := int(self.terminals[index-1].Term.X+0.5), int(self.terminals[index-1].Term.Y+0.5)
			visited[Point{x, y, z}] = true
		}
		ends := make(Vectors, self.pcb.depth, self.pcb.depth)
		for z := 0; z < self.pcb.depth; z++ {
			x, y := int(self.terminals[index].Term.X+0.5), int(self.terminals[index].Term.Y+0.5)
			ends[z] = &Point{x, y, z}
		}
		self.pcb.mark_distances(self.pcb.routing_flood_vectors, self.radius, self.via, self.gap, &visited, &ends)
		e := make(sort_points, 0, len(ends))
		end_nodes := &e
		for _, node := range ends {
			end_nodes = insert_sort_point(end_nodes, node, float32(self.pcb.get_node(node)))
		}
		e = *end_nodes
		path, success := self.backtrack_path(&visited, e[0].node, self.radius, self.via, self.gap)
		self.pcb.unmark_distances()
		if !success {
			self.remove()
			return false
		}
		for _, node := range path {
			visited[*node] = true
		}
		self.paths = append(self.paths, path)
	}
	self.paths = self.optimise_paths(self.paths[:])
	self.add_paths_collision_lines()
	self.add_terminal_collision_lines()
	return true
}

//output net, terminals and paths, for viewer app
func (self *net) print_net() {
	scale := 1.0 / float32(self.pcb.resolution)
	fmt.Print("[", self.radius*scale, ",", self.via*scale, ",", self.gap*scale, ",[")
	for i, t := range self.terminals {
		fmt.Print("(", t.Radius*scale, ",", t.Gap*scale, ",(", t.Term.X*scale, ",", t.Term.Y*scale, ",", t.Term.Z, "),[")
		for j, c := range t.Shape {
			fmt.Print("(", c.X*scale, ",", c.Y*scale, ")")
			if j != (len(t.Shape) - 1) {
				fmt.Print(",")
			}
		}
		fmt.Print("])")
		if i != (len(self.terminals) - 1) {
			fmt.Print(",")
		}
	}
	fmt.Print("],[")
	for i, path := range self.paths {
		fmt.Print("[")
		for j, p := range path {
			psp := self.pcb.grid_to_space_point(p)
			sp := *psp
			fmt.Print("(", sp[0]*scale, ",", sp[1]*scale, ",", sp[2], ")")
			if j != (len(path) - 1) {
				fmt.Print(",")
			}
		}
		fmt.Print("]")
		if i != (len(self.paths) - 1) {
			fmt.Print(",")
		}
	}
	fmt.Println("]]")
	return
}
