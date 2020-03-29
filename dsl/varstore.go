package dsl

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/emicklei/melrose"
	"github.com/emicklei/melrose/notify"
)

type Variable struct {
	Name  string
	store *VariableStore
}

func (v Variable) Value() interface{} {
	m, ok := v.store.Get(v.Name)
	if !ok {
		return nil
	}
	return m
}

func (v Variable) Storex() string {
	return fmt.Sprintf(`var(%s)`, v.Name)
}

func (v Variable) String() string {
	return fmt.Sprintf(`%s:%T`, v.Name, v.Value())
}

func (v Variable) S() melrose.Sequence {
	if s, ok := v.Value().(melrose.Sequenceable); ok {
		return s.S()
	}
	return melrose.Sequence{}
}

type VariableStore struct {
	mutex     sync.RWMutex
	variables map[string]interface{}
}

func NewVariableStore() *VariableStore {
	return &VariableStore{
		variables: map[string]interface{}{},
	}
}

func (v *VariableStore) NameFor(value interface{}) string {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	for k, v := range v.variables {
		if reflect.DeepEqual(value, v) {
			return k
		}
	}
	return "" // not found
}

func (v *VariableStore) Get(key string) (interface{}, bool) {
	v.mutex.RLock()
	e, ok := v.variables[key]
	v.mutex.RUnlock()
	return e, ok
}

func (v *VariableStore) Put(key string, value interface{}) {
	v.mutex.Lock()
	v.variables[key] = value
	v.mutex.Unlock()
}

func (v *VariableStore) Delete(key string) {
	v.mutex.Lock()
	delete(v.variables, key)
	v.mutex.Unlock()
}

func (v *VariableStore) Variables() map[string]interface{} {
	v.mutex.RLock()
	copy := map[string]interface{}{}
	for k, v := range v.variables {
		copy[k] = v
	}
	v.mutex.RUnlock()
	return copy
}

func (v *VariableStore) ListVariables(entry string) notify.Message {
	v.mutex.RLock()
	defer v.mutex.RUnlock()
	keys := []string{}
	width := 0
	for k, _ := range v.variables {
		if len(k) > width {
			width = len(k)
		}
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := v.variables[k]
		if s, ok := v.(melrose.Storable); ok {
			fmt.Printf("%s = %s\n", strings.Repeat(" ", width-len(k))+k, s.Storex())
		} else {
			fmt.Printf("%s = (%T) %v\n", strings.Repeat(" ", width-len(k))+k, v, v)
		}
	}
	return nil
}

type Snapshot struct {
	Variables     map[string]string `json:"variables"`
	Configuration map[string]string `json:"configuration"`
}

const defaultStorageFilename = "melrose.json"

func (s *VariableStore) LoadMemoryFromDisk(entry string) notify.Message {
	f, err := os.Open(defaultStorageFilename)
	if err != nil {
		return notify.Errorf("unable to load:%v", err)
	}
	defer f.Close()

	snap := Snapshot{}
	dec := json.NewDecoder(f)
	if err := dec.Decode(&snap); err != nil {
		return notify.Error(err)
	}

	// if var is used and its value is not known we do a second pass.
	// TODO workaround fix
	secondsPass := map[string]string{}
	// load into existing
	for k, storex := range snap.Variables {
		v, err := Evaluate(s, storex)
		if err != nil {
			secondsPass[k] = storex
			continue
		}
		if r, ok := v.(EvaluationResult); ok {
			s.Put(k, r.Result)
		}
	}
	for k, storex := range secondsPass {
		v, err := Evaluate(s, storex)
		if err != nil {
			return notify.Errorf("unable to evaluate [%s = %s] because: %v", k, storex, err)
		}
		if r, ok := v.(EvaluationResult); ok {
			s.Put(k, r.Result)
		}
	}
	// handle configuration
	if v, ok := snap.Configuration["bpm"]; ok {
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return notify.Errorf("invalid value for bpm (beat-per-minute):%v", v)
		} else {
			melrose.CurrentDevice().SetBeatsPerMinute(f)
		}
	}
	//printInfo(fmt.Sprintf("loaded %d variables. use \":v\" to list them", len(snap.Variables)))
	return nil
}

func (s *VariableStore) SaveMemoryToDisk(entry string) notify.Message {
	f, err := os.Create(defaultStorageFilename)
	if err != nil {
		return notify.Errorf("unable to save:%v", err)
	}
	defer f.Close()

	storeMap := map[string]string{}
	s.mutex.RLock()
	for k, v := range s.variables {
		if s, ok := v.(melrose.Storable); ok {
			storeMap[k] = s.Storex()
		} else {
			return notify.Errorf("cannot store %q:%T\n", k, v)
		}
	}

	snap := Snapshot{
		Variables: storeMap,
		Configuration: map[string]string{
			"bpm": fmt.Sprintf("%v", melrose.CurrentDevice().BeatsPerMinute()),
		},
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "\t")
	if err := enc.Encode(snap); err != nil {
		return notify.Errorf("%v", err)
	}
	return notify.Infof("saved %d variables. use \":v\" to list them", len(storeMap))
}