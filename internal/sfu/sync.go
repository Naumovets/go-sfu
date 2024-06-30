package sfu

import "sync"

var (
	ListLock sync.RWMutex
)
