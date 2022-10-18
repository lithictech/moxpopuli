package asyncapispec

import (
	"github.com/lithictech/moxpopuli/schema"
	"strings"
)

type Message map[string]interface{}

func (o Message) GetOrAddBindings() MessageBindings {
	return getOrAddMap(o, "bindings")
}

func (o Message) GetOrAddPayload() schema.Schema {
	return getOrAddSchema(o, "payload")
}

func (o Message) GetOrAddHeaders() schema.Schema {
	return getOrAddSchema(o, "headers")
}

func (o Message) ContentType() string {
	if s, ok := o["contentType"]; ok {
		return s.(string)
	}
	return ""
}

func (o Message) CorrelationIdHeaderKey() (string, bool) {
	cid, ok := o["correlationId"]
	if !ok {
		return "", false
	}
	cidm := cid.(map[string]interface{})
	loc := cidm["location"].(string)
	headerloc := "$message.header#/"
	if !strings.HasPrefix(loc, headerloc) {
		return "", false
	}
	return loc[len(headerloc):], true
}

type MessageBindings map[string]interface{}

func (o MessageBindings) GetOrAdd(key string) MessageBinding {
	return getOrAddMap(o, key)
}

func (o MessageBindings) GetOrAddHttp() HttpMessageBinding {
	return HttpMessageBinding(o.GetOrAdd("http"))
}

type MessageBinding map[string]interface{}

type HttpMessageBinding map[string]interface{}

func (o HttpMessageBinding) GetOrAddHeaders() schema.ObjectSchema {
	h := getOrAddSchema(o, "headers")
	if _, ok := h["type"]; !ok {
		h["type"] = "object"
		h["properties"] = map[string]schema.Schema{}
	}
	return h.MustObject()
}
