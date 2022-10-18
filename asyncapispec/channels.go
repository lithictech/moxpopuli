package asyncapispec

type Channels map[string]interface{}

func (c Channels) GetOrAddItem(key string) ChannelItem {
	return getOrAddMap(c, key)
}

type ChannelItem map[string]interface{}

func (c ChannelItem) GetOrAddSubscribe() Operation {
	return getOrAddMap(c, "subscribe")
}
