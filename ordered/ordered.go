package ordered

import (
	"sort"
)

type UnsafeInterfaceMap struct {
	keys      []interface{}
	keyValues map[interface{}]interface{}
}

func NewUnsafeInterfaceMap() *UnsafeInterfaceMap {
	return &UnsafeInterfaceMap{keyValues: make(map[interface{}]interface{})}
}

func (this *UnsafeInterfaceMap) Get(key interface{}) interface{} {
	return this.keyValues[key]
}

func (this *UnsafeInterfaceMap) Set(key, value interface{}) {
	this.keys = append(this.keys, key)
	this.keyValues[key] = value
}

func (this *UnsafeInterfaceMap) Keys() []interface{} {
	return this.keys
}

// func (this *UnsafeInterfaceMap) Filter(fn func(key, value interface{}) bool) *UnsafeInterfaceMap {
// 	filtered := NewUnsafeInterfaceMap()
// 	for _, k := range this.keys {
// 		if fn(k, this.keyValues[k]) {
// 			filtered.Set(k, this.keyValues[k])
// 		}
// 	}
// 	return filtered
// }

func (this *UnsafeInterfaceMap) Len() int { return len(this.keys) }
func (this *UnsafeInterfaceMap) Swap(i, j int) {
	this.keys[i], this.keys[j] = this.keys[j], this.keys[i]
}

type StringStringMap struct {
	*UnsafeInterfaceMap
}

func NewStringStringMap() *StringStringMap {
	m := NewUnsafeInterfaceMap()
	return &StringStringMap{m}
}

func (this *StringStringMap) Get(key string) string {
	v, _ := this.UnsafeInterfaceMap.Get(key).(string)
	return v
}

func (this *StringStringMap) Keys() []string {
	s := make([]string, len(this.keys))
	for i, k := range this.keys {
		s[i] = k.(string)
	}
	return s
}

func (this *StringStringMap) Filter(fn func(key, value string) bool) *StringStringMap {
	filtered := NewStringStringMap()
	for _, k := range this.keys {
		if fn(k.(string), this.keyValues[k].(string)) {
			filtered.Set(k.(string), this.keyValues[k].(string))
		}
	}
	return filtered
}

func (this *StringStringMap) SortByKey() {
	sort.Sort(stringStringMapByKey{this})
}

func (this *StringStringMap) SortByValue() {
	sort.Sort(stringStringMapByValue{this})
}

type stringStringMapByKey struct {
	*StringStringMap
}

func (this stringStringMapByKey) Less(i, j int) bool {
	return this.keys[i].(string) < this.keys[j].(string)
}

type stringStringMapByValue struct {
	*StringStringMap
}

func (this stringStringMapByValue) Less(i, j int) bool {
	return this.keyValues[this.keys[i].(string)].(string) < this.keyValues[this.keys[j].(string)].(string)
}
