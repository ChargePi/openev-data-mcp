package database

import (
	"context"
	"errors"
	"time"

	gormlogger "gorm.io/gorm/logger"
	"go.uber.org/zap"
)

type zapGormLogger struct {
	logger                    *zap.Logger
	level                     gormlogger.LogLevel
	slowThreshold             time.Duration
	ignoreRecordNotFoundError bool
}

func newZapGormLogger(logger *zap.Logger) gormlogger.Interface {
	return &zapGormLogger{
		logger:                    logger,
		level:                     gormlogger.Warn,
		slowThreshold:             200 * time.Millisecond,
		ignoreRecordNotFoundError: true,
	}
}

func (l *zapGormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	clone := *l
	clone.level = level
	return &clone
}

func (l *zapGormLogger) Info(_ context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Info {
		l.logger.Sugar().Infof(msg, args...)
	}
}

func (l *zapGormLogger) Warn(_ context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Warn {
		l.logger.Sugar().Warnf(msg, args...)
	}
}

func (l *zapGormLogger) Error(_ context.Context, msg string, args ...any) {
	if l.level >= gormlogger.Error {
		l.logger.Sugar().Errorf(msg, args...)
	}
}

func (l *zapGormLogger) Trace(_ context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	if l.level <= gormlogger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.String("sql", sql),
		zap.Int64("rows", rows),
		zap.Duration("elapsed", elapsed),
	}

	switch {
	case err != nil && (!l.ignoreRecordNotFoundError || !errors.Is(err, gormlogger.ErrRecordNotFound)):
		l.logger.Error("gorm query error", append(fields, zap.Error(err))...)
	case elapsed > l.slowThreshold:
		l.logger.Warn("gorm slow query", fields...)
	case l.level >= gormlogger.Info:
		l.logger.Debug("gorm query", fields...)
	}
}