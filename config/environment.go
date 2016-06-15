package config

import "os"

func QuiltHome() string {
	return os.Getenv("QUILT_HOME")
}
