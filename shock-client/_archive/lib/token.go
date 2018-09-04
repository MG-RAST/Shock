package lib

import (
	"encoding/json"
	"io/ioutil"
	"os/user"
	"syscall"
	"time"
)

func (t *Token) ExpiresInDays() int {
	d := time.Duration(t.ExpiresIn) * time.Second
	return int(d.Hours()) / 24
}

func (t *Token) Path() (path string) {
	u, _ := user.Current()
	return u.HomeDir + "/.shock-client.token"
}

func (t *Token) Delete() (err error) {
	if err := syscall.Unlink(t.Path()); err != nil {
		if err.Error() == "no such file or directory" {
			return nil
		} else {
			return err
		}
	}
	return nil
}

func (t *Token) Load() (err error) {
	if f, err := ioutil.ReadFile(t.Path()); err != nil {
		return err
	} else {
		if err = json.Unmarshal(f, &t); err != nil {
			return err
		}
	}
	return
}

func (t *Token) Store() (err error) {
	m, _ := json.Marshal(t)
	return ioutil.WriteFile(t.Path(), m, 0600)
}
