package ab

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/willf/bitset"
)

type contextKey struct{}

const (
	EnvKeyVersion = "version"
	EnvKeyCost    = "cost"
	EnvKeyUA      = "ua"
	EnvKeyReqId   = "request_id"
)

var _contextKey = contextKey{}

// T contains all ab test context
type T struct {
	// cache of frozen flag values, array of kv
	flagValues        []*KV
	flagDefaultValues []*KV

	// array of domain/group
	frozenUnits []*item

	root         *tree
	env          []*KV
	didGroups    map[int64]struct{}
	conflictMask *bitset.BitSet
	usedFlags    *bitset.BitSet
	// creation timestamp of T
	start time.Time
}

// Add registers biz variables for condition evaluation. Caution: must add kv first, then retrieve flag values.
func (t *T) Add(kvs ...KV) {
	envMap := Registry.loadEnv()
	for i := range kvs {
		kv := kvs[i]
		if v, ok := envMap[kv.Key]; ok {
			t.env[v.index] = &kv
		}
	}
}

// Log reports experiment result with environments.
func (t *T) Log(kv ...KV) {
	if len(Registry.flags) == 0 {
		return
	}
	var flags []string
	for i, ok := t.usedFlags.NextSet(0); ok; i, ok = t.usedFlags.NextSet(i + 1) {
		flags = append(flags, Registry.flags[i].name)
	}
	if t.root != nil && t.root.Root() != nil {
		kv = append(kv, KVString(EnvKeyVersion, t.root.Root().Version()))
	}
	kv = append(kv, KVInt(EnvKeyCost, time.Since(t.start).Nanoseconds()/int64(time.Millisecond)))
	Registry.reporter.Log(t.Snapshot(), flags, kv)
}

func (t *T) did(id int64) (b bool) {
	_, b = t.didGroups[id]
	return
}

func (t *T) isConflict(d *domain, l *layer) bool {
	u := t.frozenUnits[l.index]
	if u == nil {
		return false
	}
	if u.Type == typeNil && d == nil {
		return false
	} else if d != nil && u.Type == typeDomain && u.Domain == d {
		return false
	}
	return true
}

// Snapshot returns internal states of experiment.
func (t *T) Snapshot() (states []State) {
	for i, it := range t.frozenUnits {
		if it == nil {
			continue
		}
		if t.conflictMask.Test(uint(i)) {
			states = append(states, State{
				Type:  LayerConflict,
				Value: t.root.layers[i].id,
			})
			continue
		}
		if it.Type == typeDomain {
			continue
		}
		st := State{}
		if it.Type == typeNil {
			st.Type = LayerNoHit
			st.Value = t.root.layers[i].id
		} else if it.Type == typeGroup {
			st.Type = ExpHit
			st.Value = it.Group.id
		}
		states = append(states, st)
	}
	return
}

func (t *T) divert(l *layer, frozen []*item) (i *item) {
	var (
		curLayer, parentLayer *layer
		curDomain             *domain
		parentUnit, unit      *item
	)
	if i = frozen[l.index]; i != nil {
		return i
	}

	path := []*layer{l}
	for len(path) > 0 {
		curLayer = path[0]
		if frozen[curLayer.index] != nil {
			path = path[1:]
			continue
		}
		curDomain = curLayer.parent
		parentLayer = curDomain.parent
		if parentLayer != nil {
			parentUnit = frozen[parentLayer.index]
			if parentUnit == nil {
				path = append(path, parentLayer)
				continue
			} else if parentUnit.Type == typeNil || parentUnit.Domain != curDomain {
				return emptyItem
			}
		}
		unit = curLayer.divert(t)
		if unit.Type == typeGroup {
			return unit
		}
		frozen[curLayer.index] = unit
	}
	return frozen[l.index]
}

func (t *T) value(f *flag) (kv *KV) {
	var (
		l         *layer
		layerTree *inverseTree
		unit      *item
	)
	index := f.index
	t.usedFlags.Set(index)
	if kv = t.flagValues[index]; kv != nil {
		return
	}
	if t.root != nil {
		if layerTree = t.root.flagLayers[index]; layerTree != nil {
			if layerTree.launch != nil {
				unit = t.divert(layerTree.launch, t.frozenUnits)
				if unit.Type == typeGroup {
					t.applyUnit(unit, layerTree.launch, false)
				}
			}
			if layerTree.root != nil && t.frozenUnits[layerTree.root.index] == nil {
				unit = t.divert(layerTree.root, t.frozenUnits)
				if unit.Type == typeGroup {
					t.applyUnit(unit, layerTree.root, true)
				}
			}
			for _, l = range layerTree.leaves {
				unit = t.divert(l, t.frozenUnits)
				if unit.Type == typeGroup {
					t.applyUnit(unit, l, true)
				}
			}
		}
		if kv = t.flagValues[index]; kv != nil {
			return
		}
		if kv = t.flagDefaultValues[index]; kv != nil {
			return
		}
	}
	kv = &f.val
	t.flagValues[index] = kv
	return
}

func (t *T) applyUnit(i *item, l *layer, backtrace bool) {
	if oldUnit := t.frozenUnits[l.index]; oldUnit != nil {
		if oldUnit.Domain != i.Domain && oldUnit.Group != i.Group {
			t.conflictMask.Set(l.index)
		}
		return
	}
	if backtrace {
		path := make(map[uint]*domain)
		curDomain := l.parent
		for {
			if curDomain.parent == nil {
				break
			}
			if t.isConflict(curDomain, curDomain.parent) {
				t.conflictMask.Set(curDomain.parent.index)
				return
			}
			path[curDomain.parent.index] = curDomain
			curDomain = curDomain.parent.parent
		}
		for index, unit := range path {
			t.frozenUnits[index] = &item{
				Type:   typeDomain,
				Domain: unit,
			}
		}
	}

	t.frozenUnits[l.index] = i
	if i.Type != typeGroup {
		return
	}
	g := i.Group
	t.didGroups[g.id] = struct{}{}
	for k := range g.flags {
		//nolint:gosimple
		v, _ := g.flags[k]
		if l.isLaunched {
			t.flagDefaultValues[k] = &v
		} else {
			t.flagValues[k] = &v
		}
	}
}

func (t *T) restore(states ...State) {
	if t.root == nil {
		return
	}
	var (
		g  *group
		l  *layer
		tp itemType
	)
	for _, s := range states {
		g = nil
		//nolint:ineffassign
		l = nil
		switch s.Type {
		case LayerConflict:
			t.conflictMask.Set(t.root.layerIDMap[s.Value].index)
			continue
		case LayerNoHit:
			tp = typeNil
			l = t.root.layerIDMap[s.Value]
		case ExpHit:
			tp = typeGroup
			g = t.root.groupIDMap[s.Value]
			l = g.parent.parent
		default:
			continue
		}
		t.applyUnit(&item{
			Type:  tp,
			Group: g,
		}, l, true)
	}
}

// New creates T with given env variables.
func New(env ...KV) (t *T) {
	var (
		nflag, nlayer, nenv uint
		tree                *tree
		envMap              map[string]envVar
	)
	nflag = uint(len(Registry.flagMap))
	tree = Registry.Tree()
	if tree == nil {
		nlayer = 0
	} else {
		nlayer = uint(tree.LayerNum())
	}
	envMap = Registry.loadEnv()
	nenv = uint(len(envMap))
	t = &T{
		flagValues:        make([]*KV, nflag),
		flagDefaultValues: make([]*KV, nflag),
		frozenUnits:       make([]*item, nlayer),
		root:              tree,
		env:               make([]*KV, nenv),
		didGroups:         make(map[int64]struct{}),
		conflictMask:      bitset.New(nlayer),
		usedFlags:         bitset.New(nflag),
		start:             time.Now(),
	}
	t.Add(env...)
	return t
}

// NewContext adds T as a part of given context and returns a new one.
func NewContext(ctx context.Context, t *T) context.Context {
	return context.WithValue(ctx, _contextKey, t)
}

// FromContext returns T from current context, nil if not exists.
func FromContext(ctx context.Context) (t *T, ok bool) {
	t, ok = ctx.Value(_contextKey).(*T)
	return
}

// Extract restores experiment states to T from carrier.
func Extract(t *T, carrier Carrier) *T {
	states, kvs := carrier.Get()
	if states != nil {
		t.restore(states...)
	}
	if kvs != nil {
		t.Add(kvs...)
	}
	return t
}

// Inject sets experiment states of T into carrier.
func Inject(t *T, carrier Carrier) {
	var kvs []KV
	for _, kv := range t.env {
		if kv != nil {
			kvs = append(kvs, *kv)
		}
	}
	carrier.Set(t.Snapshot(), kvs)
}

// Refresh updates root domain.
func Refresh(cd *Domain) error {
	t := buildTree(cd)
	if t == nil {
		return errors.Errorf("ab: unable to build tree from domain(%+v)\n", cd)
	}
	Registry.storeTree(t)
	return nil
}

func divertKey(t *T, l *layer) (s string) {
	kv := t.env[l.diversion]
	if kv == nil {
		s = ""
	} else if kv.Type == typeInt64 {
		s = strconv.FormatInt(kv.Int64, 10)
	} else {
		s = kv.String
	}
	return
}

func (l *layer) divert(t *T) *item {
	for w, g := range l.whitelist {
		if w.Matches(t) {
			return g
		}
	}
	//nolint:staticcheck
	if l.diversion < 0 || t.env[l.diversion] == nil {
		return emptyItem
	}

	key := fmt.Sprintf("%s_%d", divertKey(t, l), l.id)
	index := Registry.policy(key) % 100
	child := l.bucketMapping[uint16(index)]
	if child == nil {
		return emptyItem
	}
	if cond := child.Cond(); cond != nil && !cond.Matches(t) {
		return emptyItem
	}
	if child.Type == typeExp {
		return child.Exp.divert(t)
	}
	return child
}

func (e *exp) divert(t *T) *item {
	l := e.parent
	//nolint:staticcheck
	if l.diversion < 0 || t.env[l.diversion] == nil {
		return emptyItem
	}
	key := fmt.Sprintf("%s_%s", divertKey(t, l), e.seed)
	index := Registry.policy(key) % 100
	child := e.bucketToGroup[uint16(index)]
	if child == nil {
		return emptyItem
	}
	return child
}
