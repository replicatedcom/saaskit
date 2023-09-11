package store

import (
	"sync"

	"database/sql/driver"

	"github.com/replicatedcom/saaskit/tracing/datadog"
)

// Datadog driver is not threadsafe and can panic during startup
var ddLock = sync.Mutex{}

func RegisterDatadogDriver(driverName string, driver driver.Driver, dbName string) {
	ddLock.Lock()
	defer ddLock.Unlock()

	// Safe to call multiple times, isRegistered function is not exported
	datadog.RegisterSQL(driverName, driver, dbName)
}
