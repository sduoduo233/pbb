package controllers

type D map[string]any

func If[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}
