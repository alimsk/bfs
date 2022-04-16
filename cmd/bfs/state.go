package main

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"os"

	"github.com/alimsk/shopee"
	jsoniter "github.com/json-iterator/go"
)

// implements json.Marshaler and json.Unmarshaler for http.CookieJar
type CookieJarMarshaler struct{ http.CookieJar }

var _ json.Marshaler = (*CookieJarMarshaler)(nil)
var _ json.Unmarshaler = (*CookieJarMarshaler)(nil)

func (c *CookieJarMarshaler) MarshalJSON() ([]byte, error) {
	return jsoniter.MarshalIndent(c.Cookies(shopee.ShopeeUrl), "", "  ")
}

func (c *CookieJarMarshaler) UnmarshalJSON(data []byte) error {
	var v []*http.Cookie
	if err := jsoniter.Unmarshal(data, &v); err != nil {
		return err
	}
	c.CookieJar, _ = cookiejar.New(nil)
	c.SetCookies(shopee.ShopeeUrl, v)
	return nil
}

type State struct {
	Cookies []*CookieJarMarshaler
}

func loadStateFile(name string) (*State, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var s State
	err = jsoniter.NewDecoder(f).Decode(&s)
	return &s, err
}

func (s *State) saveAsFile(name string) error {
	f, err := os.Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	e := jsoniter.NewEncoder(f)
	e.SetIndent("", "  ")
	return e.Encode(s)
}
