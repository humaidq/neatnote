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

func (p HotPosts) Less(i, j int) bool {
	// Time since posting i and j
	ti := time.Now().Sub(time.Unix(p[i].CreatedUnix, 0))
	tj := time.Now().Sub(time.Unix(p[j].CreatedUnix, 0))
	// Scores of i and j
	si := float64(p[i].Iota) / (ti.Seconds() / hotnessDelta)
	sj := float64(p[j].Iota) / (tj.Seconds() / hotnessDelta)
	return si > sj
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
