package logger

import (
    "strings"
    "fmt"
	"time"
)

const (
    NoLogging = 0
    Trace = 10
    Debug = 20
    Info = 40
    Warn = 50
    Error = 60
    Fatal = 70
)


type FrameworkLogSource interface {
    SetFrameworkLogger(logger Logger)
}


type Logger interface {
    LogTrace(message string)
    LogDebug(message string)
    LogInfo(message string)
    LogWarn(message string)
    LogError(message string)
    LogFatal(message string)
    LogAtLevel(level int, levelLabel string, message string)
	IsLevelEnabled(level int) bool
}

type LogThresholdControl interface{
    SetGlobalThreshold(threshold int)
    SetLocalThreshold(threshold int)
}

const TraceLabel = "TRACE"
const DebugLabel = "DEBUG"
const InfoLabel = "INFO"
const WarnLabel = "WARN"
const ErrorLabel = "ERROR"
const FatalLabel = "FATAL"



func LogLevelFromLabel(label string) int{
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
    loggerName string
}

func (lal *LevelAwareLogger) IsLevelEnabled(level int) bool{
	return level >= lal.globalLogThreshold && level >= lal.localLogThreshhold
}

func (lal *LevelAwareLogger) log(prefix string, level int, message string) {

    if(level >= lal.localLogThreshhold || level >= lal.globalLogThreshold){

		t := time.Now()
        fmt.Printf("%s %s %s %s\n", t.Format(time.RFC3339), prefix, lal.loggerName, message)
    }

}

func (lal *LevelAwareLogger) LogAtLevel(level int, levelLabel string, message string) {
    lal.log(levelLabel, level, message)
}

func (lal *LevelAwareLogger) LogTrace(message string) {
    lal.log(TraceLabel, Trace, message)
}

func (lal *LevelAwareLogger) LogDebug(message string) {
    lal.log(DebugLabel, Debug, message)
}

func (lal *LevelAwareLogger) LogInfo(message string) {
    lal.log(InfoLabel, Info, message)
}

func (lal *LevelAwareLogger) LogWarn(message string) {
    lal.log(WarnLabel, Warn, message)
}

func (lal *LevelAwareLogger) LogError(message string) {
    lal.log(ErrorLabel, Error, message)
}

func (lal *LevelAwareLogger) LogFatal(message string) {
    lal.log(FatalLabel, Fatal, message)
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
    componentsLogger map[string]LogThresholdControl
	initalComponentLogLevels map[string]interface{}
    globalThreshold int
}

func CreateComponentLoggerManager(globalThreshold int, initalComponentLogLevels map[string]interface{}) *ComponentLoggerManager{
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

func (clm *ComponentLoggerManager) CreateLogger(componentId string) Logger{

	threshold := clm.globalThreshold

	if clm.initalComponentLogLevels != nil {

		if levelLabel, ok := clm.initalComponentLogLevels[componentId]; ok {
			threshold = LogLevelFromLabel(levelLabel.(string))
		}

	}


	return clm.CreateLoggerAtLevel(componentId, threshold)
}

func (clm *ComponentLoggerManager) CreateLoggerAtLevel(componentId string, threshold int) Logger{
    logger := new(LevelAwareLogger)
    logger.globalLogThreshold = clm.globalThreshold
    logger.localLogThreshhold = threshold
    logger.loggerName = componentId

    clm.componentsLogger[componentId] = logger

    return logger
}




