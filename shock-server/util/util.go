package util

import (
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/jaredwilkening/goweb"
	"math/rand"
	"strconv"
	"time"
)

const chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz1234567890"

type UrlResponse struct {
	Url       string `json:"url"`
	ValidTill string `json:"validtill"`
}

func RandString(l int) (s string) {
	rand.Seed(time.Now().UTC().UnixNano())
	c := make([]byte, l)
	for i := 0; i < l; i++ {
		c[i] = chars[rand.Intn(len(chars))]
	}
	return string(c)
}

func ToInt(s string) (i int) {
	i, _ = strconv.Atoi(s)
	return
}

func ApiUrl(cx *goweb.Context) string {
	if conf.Conf["api-url"] != "" {
		return conf.Conf["api-url"]
	}
	return "http://" + cx.Request.Host
}
