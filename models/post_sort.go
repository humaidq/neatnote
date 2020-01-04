package models

import (
	"time"
)

var hotnessDelta float64 = 86400 // NOTE: 86400 is 1 day in seconds

// HotPosts implements sort.Interface for []Post based on iota score diminished
// by time.
type HotPosts []Post

func (p HotPosts) Len() int {
	return len(p)
}

func (p HotPosts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func getHotScore(p Post) float64 {
	t := time.Now().Sub(time.Unix(p.CreatedUnix, 0)).Seconds() / hotnessDelta
	if t < 1 {
		return float64(p.Iota)
	}
	return float64(p.Iota) / t
}

func (p HotPosts) Less(i, j int) bool {
	return getHotScore(p[i]) > getHotScore(p[j])
}

// TopPosts implements sort.Interface for []Post based on highest iota.
type TopPosts []Post

func (p TopPosts) Len() int {
	return len(p)
}

func (p TopPosts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p TopPosts) Less(i, j int) bool {
	return p[i].Iota > p[j].Iota
}

// NewPosts implements sort.Interface for []Post based on new posts.
type NewPosts []Post

func (p NewPosts) Len() int {
	return len(p)
}

func (p NewPosts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p NewPosts) Less(i, j int) bool {
	return p[i].CreatedUnix > p[j].CreatedUnix
}
