package config

const (
    doorOpenDuration  = 3000
    travelDuration    = 2500
    includeCab        = false
)

type ClearRequestType int

const (
    all ClearRequestType = iota
    inDirn
)

var clearRequestType = inDirn // Initial value is inDirn
