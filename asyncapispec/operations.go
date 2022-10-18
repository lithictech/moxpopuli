package asyncapispec

import "github.com/lithictech/moxpopuli/schema"

type Operation map[string]interface{}

func (o Operation) GetOrAddBindings() OperationBindings {
	return getOrAddMap(o, "bindings")
}

func (o Operation) GetOrAddMessage() Message {
	return getOrAddMap(o, "message")
}

type OperationBindings map[string]interface{}

func (o OperationBindings) GetOrAdd(key string) OperationBinding {
	return getOrAddMap(o, key)
}

func (o OperationBindings) GetOrAddHttp() HttpOperationBinding {
	return HttpOperationBinding(o.GetOrAdd("http"))
}

type OperationBinding map[string]interface{}

type HttpOperationBinding map[string]interface{}

func (o HttpOperationBinding) Method() string {
	if s, ok := o["method"]; ok {
		return s.(string)
	}
	return ""
}

func (o HttpOperationBinding) GetOrAddOrTypeQuery() schema.Schema {
	return getOrAddSchema(o, "query")
}
