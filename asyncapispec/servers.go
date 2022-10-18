package asyncapispec

type Servers map[string]interface{}

func (c Servers) GetOrAddServer(key string) Server {
	return getOrAddMap(c, key)
}

type Server map[string]interface{}

func (s Server) Protocol() string {
	return s["protocol"].(string)
}

func (s Server) Url() string {
	return s["url"].(string)
}
