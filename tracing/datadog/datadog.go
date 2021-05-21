package datadog

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"os"
	"runtime"

	"github.com/gin-gonic/gin"
	sqltrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/database/sql"
	gintrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/gin-gonic/gin"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

var ddIsEnabled bool

func init() {
	ddIsEnabled = os.Getenv("USE_DATADOG_APM") == "true"
}

// DDFinishable confirms ddtrace to Finishable
type DDFinishable interface {
	Finish(...ddtrace.FinishOption)
}

// Finishable is a wrapped ddtrace, which can Finish
type Finishable struct {
	toFinish DDFinishable
}

// Finish a Finishable, safe if nil
func (f *Finishable) Finish() {
	if f.toFinish != nil {
		f.toFinish.Finish()
	}
}

// StartSpanFromGin takes a gin context and returns a wrapped Span plus
// the span's context. If Datadog APM isn't enabled it simply returns
// a wrapped Nil, which is safe to Finish() and the gin context's HTTP
// request context
func StartSpanFromGin(c *gin.Context) (*Finishable, context.Context) {
	if !datadogEnabled() {
		return &Finishable{nil}, c.Request.Context()
	}
	var spanName string
	rctx := c.Request.Context()
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		spanName = "undefined"
	} else {
		spanName = runtime.FuncForPC(pc).Name()
	}
	span, ctx := tracer.StartSpanFromContext(rctx, spanName)
	return &Finishable{
		toFinish: span,
	}, ctx
}

// GinMiddleware wraps gin tracer's middleware
// which already plays nice when Datadog APM is disabled
func GinMiddleware(service string) gin.HandlerFunc {
	return gintrace.Middleware(service)
}

// StartTracer starts the tracer, taking a Service name and version
func StartTracer(serviceName, version string) {
	if !datadogEnabled() {
		return
	}
	tracer.Start(
		tracer.WithService(serviceName),
		tracer.WithServiceVersion(version),
	)
}

// StopTracer stops the tracer, typically called with defer in the same
// scope as StartTracer.
func StopTracer() {
	if datadogEnabled() {
		tracer.Stop()
	}
}

// RegisterSQL Registers the SQL driver with a service name in Datadog
// when Datadog APM is enabled
func RegisterSQL(driverName string, driver driver.Driver, dbName string) {
	if datadogEnabled() {
		sqltrace.Register(driverName, driver, sqltrace.WithServiceName(dbName))
	}
}

// OpenSQL Opens a DB with the Datadog tracer when Datadog APM
// is enabled. Otherwise it uses the standrd library's sql.Open
func OpenSQL(driverName, dataSourceName string) (*sql.DB, error) {
	if datadogEnabled() {
		return sqltrace.Open(driverName, dataSourceName)
	}
	return sql.Open(driverName, dataSourceName)
}

func datadogEnabled() bool {
	return ddIsEnabled
}
