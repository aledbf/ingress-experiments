package common

type Configuration struct {
	Certificate string
	Key         string

	PodIP   string
	PodName string

	Debug bool
}
