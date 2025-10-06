package constants

import "backend-go/config"

var DOMAIN = config.GetEnv("DOMAIN", "localhost")
