// Package asyncapispec wraps the minimum of the AsyncAPI specification
// that Mox Populi has to work with.
package asyncapispec

import "github.com/lithictech/moxpopuli/schema"

type Specification map[string]interface{}

func (s Specification) GetOrAddChannels() Channels {
	return getOrAddMap(s, "channels")
}

func (s Specification) GetOrAddServers() Servers {
	return getOrAddMap(s, "servers")
}

func getOrAddMap(o map[string]interface{}, key string) map[string]interface{} {
	v, ok := o[key]
	if ok {
		return v.(map[string]interface{})
	}
	m := map[string]interface{}{}
	o[key] = m
	return m
}

func getOrAddSchema(o map[string]interface{}, key string) schema.Schema {
	v, ok := o[key]
	if !ok {
		r := schema.Schema{}
		o[key] = r
		return r
	}
	if t, ok := v.(schema.Schema); ok {
		return t
	}
	sch := schema.FromMap(v.(map[string]interface{}))
	o[key] = sch
	return sch
}
