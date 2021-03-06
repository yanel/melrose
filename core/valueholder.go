package core

import (
	"fmt"
	"github.com/emicklei/melrose/notify"
)

func String(h Valueable) string {
	if h == nil {
		return ""
	}
	val := h.Value()
	if val == nil {
		return ""
	}
	if v, ok := val.(string); ok {
		return v
	}
	return ""
}

func Float(h Valueable) float64 {
	if h == nil {
		return 0.0
	}
	val := h.Value()
	if val == nil {
		return 0.0
	}
	if v, ok := val.(float64); ok {
		return v
	}
	return 0.0
}

func Int(h Valueable) int {
	// TODO notify somehow
	if h == nil {
		return 0
	}
	val := h.Value()
	if val == nil {
		return 0
	}
	if v, ok := val.(int); ok {
		return v
	}
	// maybe the value is a Valueable
	if vv, ok := val.(Valueable); ok {
		return Int(vv)
	}
	notify.Print(notify.Warningf("Int() expected [int] but got [%T], return 0", h.Value()))
	return 0
}

func ToValueable(v interface{}) Valueable {
	if w, ok := v.(Valueable); ok {
		return w
	}
	return &ValueHolder{Any: v}
}

// ValueHolder is decorate any object to become a Valueable.
type ValueHolder struct {
	Any interface{}
}

func (h ValueHolder) Value() interface{} {
	return h.Any
}

func (h *ValueHolder) SetValue(newAny interface{}) {
	h.Any = newAny
}

func (h ValueHolder) Storex() string {
	return fmt.Sprintf("%v", h.Any)
}

func (h ValueHolder) String() string {
	return h.Storex()
}

func On(v interface{}) *ValueHolder {
	return &ValueHolder{Any: v}
}
