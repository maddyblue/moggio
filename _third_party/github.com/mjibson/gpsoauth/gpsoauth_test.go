package gpsoauth

import (
	"io/ioutil"
	"testing"
)

func TestLogin(t *testing.T) {
	username, err := ioutil.ReadFile("username")
	if err != nil {
		t.Fatal(err)
	}
	password, err := ioutil.ReadFile("password")
	if err != nil {
		t.Fatal(err)
	}
	auth, err := Login(string(username), string(password), GetNode(), "sj")
	if err != nil {
		t.Fatal(err)
	}
	_ = auth
}
