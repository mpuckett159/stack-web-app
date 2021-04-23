package db

import (
	log "github.com/sirupsen/logrus"
)

// ContextLogger declares top level context logger to use in package
var ContextLogger = log.WithFields(log.Fields{
	"package": "db",
})
