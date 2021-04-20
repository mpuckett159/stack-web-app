package db

import (
	log "github.com/sirupsen/logrus"
)

// Declare top level context logger
var ContextLogger = log.WithFields(log.Fields{
	"package": "db",
})