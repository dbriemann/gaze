package main

import "errors"

var (
	//	ErrShowExists = errors.New("show already exists")
	ErrDirNotFile = errors.New("expected file but got directory")
	ErrFileNotDir = errors.New("expected directory but got file")
)
