package lightning

import "strings"

type HeaderMap struct {
	Map map[string]string
	ctx *Ctx
}

func (hm *HeaderMap) init(headers map[string]string) *HeaderMap {
	hm.Map = make(map[string]string)
	hm.Map = headers
	return hm
}

func (hm *HeaderMap) Set(key, value string) {
	hm.Map[strings.ToLower(key)] = value
	hm.ctx.Set(key, value)
}

func (hm *HeaderMap) ContentType() string {
	return string(hm.ctx.Request().Header.ContentType())
}

func (hm *HeaderMap) All() map[string]string {
	return hm.Map
}

// Get retrieves the value of a given header
func (hm *HeaderMap) Get(key string, defaultString ...string) string {
	if val, ok := hm.Map[key]; ok {
		if ok {
			return val
		}
	}
	return hm.ctx.Get(key, defaultString...)
}
