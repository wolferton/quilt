package logger

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	NoLogging = 0
	Trace     = 10
	Debug     = 20
	Info      = 40
	Warn      = 50
	Error     = 60
	Fatal     = 70
)

type FrameworkLogSource interface {
	SetFrameworkLogger(logger Logger)
}

type Logger interface {
	LogTracef(format string, a ...interface{})
	LogDebugf(format string, a ...interface{})
	LogInfof(format string, a ...interface{})
	LogWarnf(format string, a ...interface{})
	LogErrorf(format string, a ...interface{})
	LogErrorfWithTrace(format string, a ...interface{})
	LogFatalf(format string, a ...interface{})
	LogAtLevelf(level int, levelLabel string, format string, a ...interface{})
	IsLevelEnabled(level int) bool
}

type LogThresholdControl interface {
	SetGlobalThreshold(threshold int)
	SetLocalThreshold(threshold int)
}

const TraceLabel = "TRACE"
const DebugLabel = "DEBUG"
const InfoLabel = "INFO"
const WarnLabel = "WARN"
const ErrorLabel = "ERROR"
const FatalLabel = "FATAL"

func LogLevelFromLabel(label string) int {
	switch strings.ToUpper(label) {
	case TraceLabel:
		return Trace
	case DebugLabel:
		return Debug
	case InfoLabel:
		return Info
	case ErrorLabel:
		return Error
	case FatalLabel:
		return Fatal
	}

	return NoLogging
}

type LevelAwareLogger struct {
	globalLogThreshold int
	localLogThreshhold int
	loggerName         string
}

func (lal *LevelAwareLogger) IsLevelEnabled(level int) bool {
	return level >= lal.localLogThreshhold || level >= lal.globalLogThreshold
}

func (lal *LevelAwareLogger) log(prefix string, level int, message string) {

	if lal.IsLevelEnabled(level) {
		t := time.Now()
		fmt.Printf("%s %s %s %s\n", t.Format(time.RFC3339), prefix, lal.loggerName, message)
	}

}
func (lal *LevelAwareLogger) logf(levelLabel string, level int, format string, a ...interface{}) {

	if lal.IsLevelEnabled(level) {
		t := time.Now()
		message := fmt.Sprintf(format, a...)
		fmt.Printf("%s %s %s %s\n", t.Format(time.RFC3339), levelLabel, lal.loggerName, message)
	}

}

func (lal *LevelAwareLogger) LogAtLevel(level int, levelLabel string, message string) {
	lal.log(levelLabel, level, message)
}

func (lal *LevelAwareLogger) LogAtLevelf(level int, levelLabel string, format string, a ...interface{}) {
	lal.logf(levelLabel, level, format, a...)
}

func (lal *LevelAwareLogger) LogTracef(format string, a ...interface{}) {
	lal.logf(TraceLabel, Trace, format, a...)
}

func (lal *LevelAwareLogger) LogDebugf(format string, a ...interface{}) {
	lal.logf(DebugLabel, Debug, format, a...)
}

func (lal *LevelAwareLogger) LogInfof(format string, a ...interface{}) {
	lal.logf(InfoLabel, Info, format, a...)
}

func (lal *LevelAwareLogger) LogWarnf(format string, a ...interface{}) {
	lal.logf(WarnLabel, Warn, format, a...)
}

func (lal *LevelAwareLogger) LogErrorf(format string, a ...interface{}) {
	lal.logf(ErrorLabel, Error, format, a...)
}

func (lal *LevelAwareLogger) LogErrorfWithTrace(format string, a ...interface{}) {
	trace := make([]byte, 2048)
	runtime.Stack(trace, false)

	format = format + "\n%s"
	a = append(a, trace)

	lal.logf(ErrorLabel, Error, format, a...)

}

func (lal *LevelAwareLogger) LogFatalf(format string, a ...interface{}) {
	lal.logf(FatalLabel, Fatal, format, a...)
}

func (lal *LevelAwareLogger) SetGlobalThreshold(threshold int) {
	lal.globalLogThreshold = threshold
}

func (lal *LevelAwareLogger) SetLocalThreshold(threshold int) {
	lal.localLogThreshhold = threshold
}

func (lal *LevelAwareLogger) SetThreshold(threshold int) {
	lal.SetGlobalThreshold(threshold)
	lal.SetLocalThreshold(threshold)
}

func (lal *LevelAwareLogger) SetLoggerName(name string) {
	lal.loggerName = name
}

type ComponentLoggerManager struct {
	componentsLogger         map[string]LogThresholdControl
	initalComponentLogLevels map[string]interface{}
	globalThreshold          int
}

func CreateComponentLoggerManager(globalThreshold int, initalComponentLogLevels map[string]interface{}) *ComponentLoggerManager {
	loggers := make(map[string]LogThresholdControl)
	manager := new(ComponentLoggerManager)
	manager.componentsLogger = loggers
	manager.globalThreshold = globalThreshold
	manager.initalComponentLogLevels = initalComponentLogLevels

	return manager
}

func (clm *ComponentLoggerManager) UpdateGlobalThreshold(globalThreshold int) {
	clm.globalThreshold = globalThreshold

	for _, v := range clm.componentsLogger {
		v.SetGlobalThreshold(globalThreshold)
	}
}

func (clm *ComponentLoggerManager) UpdateLocalThreshold(threshold int) {
	clm.globalThreshold = threshold

	for _, v := range clm.componentsLogger {
		v.SetLocalThreshold(threshold)
	}
}

func (clm *ComponentLoggerManager) CreateLogger(componentId string) Logger {

	threshold := clm.globalThreshold

	if clm.initalComponentLogLevels != nil {

		if levelLabel, ok := clm.initalComponentLogLevels[componentId]; ok {
			threshold = LogLevelFromLabel(levelLabel.(string))
		}

	}

	return clm.CreateLoggerAtLevel(componentId, threshold)
}

func (clm *ComponentLoggerManager) CreateLoggerAtLevel(componentId string, threshold int) Logger {
	logger := new(LevelAwareLogger)
	logger.globalLogThreshold = clm.globalThreshold
	logger.localLogThreshhold = threshold
	logger.loggerName = componentId

	clm.componentsLogger[componentId] = logger

	return logger
}

func CreateAnonymousLogger(componentId string, threshold int) Logger {
	logger := new(LevelAwareLogger)
	logger.globalLogThreshold = threshold
	logger.localLogThreshhold = threshold
	logger.loggerName = componentId

	return logger
}
