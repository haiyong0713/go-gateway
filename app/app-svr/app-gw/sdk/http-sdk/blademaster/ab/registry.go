package ab

import (
	"sync/atomic"
)

var Registry = newRegistry()

type registry struct {
	tree     atomic.Value // atomic point to tree
	env      atomic.Value // atomic value of a map, env name -> *envVar
	policy   DivertPolicy
	reporter Reporter

	flags   []*flag                         // all flags defined arranged in a slice
	flagMap map[string]*flag                // flag name -> flag
	udp     map[string]UserDefinedPredicate // func name -> user defined function
}

func newRegistry() (r *registry) {
	r = &registry{
		flagMap: make(map[string]*flag),
		udp:     make(map[string]UserDefinedPredicate),
	}
	return
}

// RegisterEnv registers env variable to global registry.
func (r *registry) RegisterEnv(kvs ...KV) {
	newMap := make(map[string]envVar)
	oldMap := r.loadEnv()
	if oldMap != nil {
		for k, v := range r.loadEnv() {
			newMap[k] = v
		}
	}
	for _, kv := range kvs {
		if _, ok := newMap[kv.Key]; ok {
			continue
		}
		newMap[kv.Key] = envVar{
			index: uint(len(newMap)),
			kv:    kv,
		}
	}
	r.env.Store(newMap)
}

// RegisterFlag registers flag to global registry.
func (r *registry) RegisterFlag(f *flag) {
	f.index = uint(len(r.flags))
	r.flags = append(r.flags, f)
	r.flagMap[f.name] = f
}

// RegisterPolicy replaces diversion policy.
func (r *registry) RegisterPolicy(p DivertPolicy) {
	r.policy = p
}

// RegisterReporter replaces reporter.
func (r *registry) RegisterReporter(rp Reporter) {
	r.reporter = rp
}

// RegisterPredicate registers user defined predicate to global registry.
func (r *registry) RegisterPredicate(name string, p UserDefinedPredicate) {
	r.udp[name] = p
}

func (r *registry) loadEnv() map[string]envVar {
	x := r.env.Load()
	if x == nil {
		return nil
	}
	return x.(map[string]envVar)
}

// Tree returns current conf tree.
func (r *registry) Tree() (t *tree) {
	x := r.tree.Load()
	if x == nil {
		return nil
	}
	return x.(*tree)
}

func (r *registry) storeTree(t *tree) {
	r.tree.Store(t)
}

// Reset is only for tinker test. Never use this in your code!!!
func (r *registry) Reset() {
	r.flags = make([]*flag, 0)
	r.flagMap = make(map[string]*flag)
	r.udp = make(map[string]UserDefinedPredicate)

	r.env.Store(make(map[string]envVar))
}
