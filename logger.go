package orm

import (
	"context"
	"errors"
	"github.com/ericnts/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"regexp"
	"time"
)

type Logger struct {
	LogLevel logger.LogLevel
}

// LogMode log mode
func (l *Logger) LogMode(level logger.LogLevel) logger.Interface {
	l.LogLevel = level
	return l
}

func (l *Logger) Info(_ context.Context, sql string, params ...interface{}) {
	log.Info(sql, params)
}
func (l *Logger) Warn(_ context.Context, sql string, params ...interface{}) {
	log.Warn(sql, params)
}
func (l *Logger) Error(_ context.Context, sql string, params ...interface{}) {
	log.Error(sql, params)
}

func (l Logger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	sql, rowsAffected := fc()
	sql = regexp.MustCompile("\\s+").ReplaceAllString(sql, " ")
	if time.Now().After(begin.Add(time.Second * 3)) {
		log.Warnf("<%v> slow sql statement: %s, rowsAffected: %d",
			time.Since(begin), sql, rowsAffected)
		return
	}
	if l.LogLevel == logger.Info && err == nil {
		log.Debugf("<%v> sql statement: %s, rowsAffected: %d",
			time.Since(begin), sql, rowsAffected)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		log.Debugf("<%v> sql statement no record: %s",
			time.Since(begin), sql)
		return
	}
	if err != nil {
		log.Errorf("bad sql: %s, error: %v", sql, err)
	}
}
