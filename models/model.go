package models

type Source struct {
	Protocol string
	Name     string
	Blob     []byte
}

type Delete struct {
	Protocol string
	Name     string
}
