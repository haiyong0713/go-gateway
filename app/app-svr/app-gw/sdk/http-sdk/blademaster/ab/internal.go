package ab

import (
	"strconv"
	"strings"

	"go-common/library/log"
)

type meta struct {
	id int64
	//nolint:structcheck
	parent  *layer
	cond    Condition
	version string
	buckets []uint16
}

type domain struct {
	meta
	layers []*layer
}

func (d *domain) Version() string {
	return d.version
}

func buildDomain(cd *Domain) (d *domain) {
	d = &domain{
		meta: meta{
			id:      cd.ID,
			cond:    parseCondition(cd.Condition),
			version: cd.Version,
			buckets: cd.Buckets,
		},
	}
	for _, cl := range cd.Layers {
		if l := buildLayer(cl); l != nil {
			l.parent = d
			d.layers = append(d.layers, l)
		}
	}
	return
}

type layer struct {
	id            int64
	conflictExpID int64
	index         uint
	diversion     uint
	parent        *domain
	isLaunched    bool
	bucketMapping map[uint16]*item
	whitelist     map[Condition]*item
}

func buildLayer(cl *Layer) (l *layer) {
	diversion, ok := Registry.loadEnv()[cl.Diversion]
	if !ok {
		log.Error("ab: fail to find diversion env(%+v)\n", cl.Diversion)
		return nil
	}

	l = &layer{
		id:            cl.ID,
		conflictExpID: -cl.ID,
		diversion:     diversion.index,
		isLaunched:    cl.Launched,
		bucketMapping: make(map[uint16]*item),
		whitelist:     make(map[Condition]*item),
	}
	for _, cd := range cl.Domains {
		d := buildDomain(cd)
		d.parent = l
		i := &item{
			Type:   typeDomain,
			Domain: d,
		}
		for _, b := range d.buckets {
			l.bucketMapping[b] = i
		}
	}
	for _, ce := range cl.Exps {
		e := buildExp(ce)
		e.parent = l
		i := &item{
			Type: typeExp,
			Exp:  e,
		}
		for _, b := range e.buckets {
			l.bucketMapping[b] = i
		}
	}
	return
}

type exp struct {
	meta
	seed          string
	bucketToGroup map[uint16]*item
}

func (e *exp) Condition() Condition {
	return e.cond
}

func buildExp(ce *Exp) (e *exp) {
	e = &exp{
		meta: meta{
			id:      ce.ID,
			cond:    parseCondition(ce.Condition),
			version: ce.Version,
			buckets: ce.Buckets,
		},
		seed:          strconv.FormatInt(ce.ID, 10),
		bucketToGroup: make(map[uint16]*item),
	}
	for _, cg := range ce.Groups {
		g := buildGroup(cg)
		g.parent = e
		i := &item{
			Type:  typeGroup,
			Group: g,
		}
		for _, b := range g.buckets {
			e.bucketToGroup[b] = i
		}
	}
	return
}

type group struct {
	id        int64
	parent    *exp
	buckets   []uint16
	flags     map[uint]KV
	whitelist []Condition
}

func buildGroup(cg *Group) (g *group) {
	g = &group{
		id:      cg.ID,
		buckets: cg.Buckets,
		flags:   make(map[uint]KV),
	}
	for k, v := range cg.Whitelist {
		g.whitelist = append(g.whitelist, newInCondition(k, v...))
	}
	for k, v := range cg.Flags {
		if f, ok := Registry.flagMap[k]; ok {
			kv, err := parseKV(f.val, strings.Trim(v, "\""))
			if err != nil {
				log.Error("ab: fail to parse flag value(%s=%s)", k, v)
				continue
			}
			g.flags[f.index] = kv
		}
	}
	return
}

// inverseTree is a tree of all layers containing a flag
type inverseTree struct {
	leaves map[int64]*layer // layer id -> *layer, leavers are layers containing groups directly
	root   *layer           // common ancestor of all layers containing this flag
	launch *layer           // launch layer containing this flag
}

type tree struct {
	flagLayers []*inverseTree // slice of inverse trees, can use flag index to find related inverse tree

	layerIDMap map[int64]*layer // layer id -> *layer
	groupIDMap map[int64]*group // group id -> *group
	// the previous two maps are for injecting states from upstream systems
	layers []*layer // layer index -> *layer
	root   *domain
}

func (t *tree) init(d *domain) {
	t.root = d
	registerDomain(d, t)
}

// LayerNum returns total count of layers.
func (t *tree) LayerNum() int {
	return len(t.layers)
}

// Layer returns the i-th layer.
func (t *tree) Layer(i int) *layer {
	return t.layers[i]
}

// Root returns top-level domain of current tree.
func (t *tree) Root() *domain {
	return t.root
}

func newTree() (t *tree) {
	t = &tree{
		flagLayers: make([]*inverseTree, len(Registry.flags)),
		layerIDMap: make(map[int64]*layer),
		groupIDMap: make(map[int64]*group),
	}
	return
}

func buildTree(cd *Domain) (t *tree) {
	d := buildDomain(cd)
	t = newTree()
	t.init(d)
	return
}

func registerDomain(d *domain, t *tree) {
	for _, l := range d.layers {
		registerLayer(l, t)
	}
}

func registerLayer(l *layer, t *tree) {
	if _, ok := t.layerIDMap[l.id]; ok {
		return
	}
	l.index = uint(len(t.layers))
	t.layers = append(t.layers, l)
	t.layerIDMap[l.id] = l

	for _, u := range l.bucketMapping {
		switch u.Type {
		case typeDomain:
			registerDomain(u.Domain, t)
		case typeExp:
			for _, i := range u.Exp.bucketToGroup {
				g := i.Group
				if _, ok := t.groupIDMap[g.id]; ok {
					continue
				}
				t.groupIDMap[g.id] = g
				rootLayer := l
				if !rootLayer.isLaunched {
					for rootLayer.parent != t.root {
						rootLayer = rootLayer.parent.parent
					}
				}
				for _, cond := range g.whitelist {
					rootLayer.whitelist[cond] = i
				}
				for k := range g.flags {
					it := t.flagLayers[k]
					if it == nil {
						it = &inverseTree{
							leaves: make(map[int64]*layer),
						}
						t.flagLayers[k] = it
					}
					if l.isLaunched {
						it.launch = l
					} else {
						it.leaves[l.id] = l
						it.root = rootLayer
					}
				}
			}
		default:
		}
	}
}
