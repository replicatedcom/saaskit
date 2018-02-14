package param

import (
	"log"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
)

var (
	DefaultParamCache *ParamCache
)

func Init(sess *session.Session) {
	useSSM := os.Getenv("USE_EC2_PARAMETERS") != ""
	DefaultParamCache = NewParamCache(sess, useSSM)
}

func Lookup(envName string, ssmName string, decrypt bool) string {
	if DefaultParamCache == nil {
		panic("must initialize param package")
	}
	return DefaultParamCache.Lookup(envName, ssmName, decrypt)
}

func NewParamCache(sess *session.Session, useSSM bool) *ParamCache {
	var svc *ssm.SSM
	if useSSM {
		svc = ssm.New(sess)
	}
	return &ParamCache{
		ssm: svc,
		m:   map[string]string{},
	}
}

type ParamCache struct {
	ssm *ssm.SSM
	m   map[string]string
	sync.RWMutex
}

func (c *ParamCache) Lookup(envName, ssmName string, decrypt bool) string {
	if c.ssm == nil || ssmName == "" {
		return c.envGet(envName)
	}
	if val, ok := c.mapGet(ssmName); ok {
		return val
	}
	val, _ := c.ssmGet(ssmName, decrypt)
	return val
}

func (c *ParamCache) envGet(envName string) string {
	return os.Getenv(envName)
}

func (c *ParamCache) ssmGet(ssmName string, decrypt bool) (string, bool) {
	resp, err := c.ssm.GetParameters(&ssm.GetParametersInput{
		Names: []*string{
			&ssmName,
		},
		WithDecryption: aws.Bool(decrypt),
	})
	if err != nil {
		log.Printf("Failed to get ssm param %s: %v", ssmName, err)
		return "", false
	}
	if len(resp.InvalidParameters) > 0 {
		for _, p := range resp.InvalidParameters {
			log.Printf("Ssm param %s invalid", *p)
		}
		return "", false
	}

	val := *(resp.Parameters[0].Value)
	c.mapSet(ssmName, val)
	return val, true
}

func (c *ParamCache) mapGet(ssmName string) (string, bool) {
	c.RLock()
	defer c.RUnlock()
	val, ok := c.m[ssmName]
	return val, ok
}

func (c *ParamCache) mapSet(ssmName string, val string) {
	c.Lock()
	defer c.Unlock()
	c.m[ssmName] = val
}
