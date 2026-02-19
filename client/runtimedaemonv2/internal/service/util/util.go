package util

import (
	"runtime"
	"strings"
)

const (
	WindowsOS = "windows"
)

func Ptr[T any](v T) *T {
	return &v
}

func IsWindows() bool {
	os := runtime.GOOS

	if strings.HasPrefix(os, WindowsOS) {
		return true
	}

	return false
}

func InList[T comparable](v T, list []T) bool {
	for _, item := range list {
		if item == v {
			return true
		}
	}

	return false
}

func SmartDivision(a, b int64) float64 {
	if b == 0 {
		return 0
	}

	return float64(a) / float64(b)
}
