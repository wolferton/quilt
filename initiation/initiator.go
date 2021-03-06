package initiation

import (
	"flag"
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/jsonmerger"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/logging"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"
)

const initiatorComponentName string = ioc.FrameworkPrefix + "FrameworkInitiator"
const jsonMergerComponentName string = ioc.FrameworkPrefix + "JsonMerger"
const configAccessorComponentName string = ioc.FrameworkPrefix + "ConfigAccessor"
const facilityInitialisorComponentName string = ioc.FrameworkPrefix + "FacilityInitialisor"

type Initiator struct {
	logger logging.Logger
}

func (i *Initiator) Start(customComponents []*ioc.ProtoComponent) {

	start := time.Now()

	if config.QuiltHome() == "" {
		fmt.Println("QUILT_HOME environment variable not set")
		os.Exit(-1)
	}

	var params map[string]string
	/*protoComponents := make(map[string]*ioc.ProtoComponent)

	for _, c := range customComponents {
		protoComponents[c.Component.Name] = c
	}*/

	params = i.parseArgs()

	bootstrapLogLevel := logging.LogLevelFromLabel(params["logLevel"])
	frameworkLoggingManager, logManageProto := BootstrapFrameworkLogging(bootstrapLogLevel)
	i.logger = frameworkLoggingManager.CreateLogger(initiatorComponentName)

	i.logger.LogInfof("Starting components")

	configAccessor := i.loadConfigIntoAccessor(params["config"], frameworkLoggingManager)
	container := ioc.NewContainer(frameworkLoggingManager, configAccessor)

	container.AddProto(logManageProto)
	container.AddProtos(customComponents)

	facilitiesInitialisor := NewFacilitiesInitialisor(container, frameworkLoggingManager)
	facilitiesInitialisor.Logger = frameworkLoggingManager.CreateLogger(facilityInitialisorComponentName)

	err := facilitiesInitialisor.Initialise(configAccessor)
	i.shutdownIfError(err, container)

	err = container.Populate()
	i.shutdownIfError(err, container)

	runtime.GC()

	err = container.StartComponents()
	i.shutdownIfError(err, container)

	elapsed := time.Since(start)
	i.logger.LogInfof("Ready (startup time %s)", elapsed)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)

	go func() {
		<-c
		i.shutdown(container)
		os.Exit(1)
	}()

	for {
		time.Sleep(100000000000)
	}
}

func (i *Initiator) shutdownIfError(err error, c *ioc.ComponentContainer) {

	if err != nil {
		i.logger.LogFatalf(err.Error())
		i.shutdown(c)
		os.Exit(-1)
	}

}

func (i *Initiator) shutdown(container *ioc.ComponentContainer) {
	i.logger.LogInfof("Shutting down")

	container.ShutdownComponents()

}

func (i *Initiator) loadConfigIntoAccessor(configPath string, frameworkLoggingManager *logging.ComponentLoggerManager) *config.ConfigAccessor {
	configFiles := i.builtInConfigPaths()

	expandedPaths, err := config.ExpandToFiles(i.splitConfigPaths(configPath))
	fl := frameworkLoggingManager.CreateLogger(configAccessorComponentName)

	if err != nil {
		i.logger.LogFatalf("Unable to load specified config files: %s", err.Error())
		return nil
	}

	configFiles = append(configFiles, expandedPaths...)

	if i.logger.IsLevelEnabled(logging.Debug) {

		i.logger.LogDebugf("Loading configuration from: ")

		for _, fileName := range configFiles {
			i.logger.LogDebugf(fileName)
		}
	}

	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = frameworkLoggingManager.CreateLogger(jsonMergerComponentName)

	mergedJson := jsonMerger.LoadAndMergeConfig(configFiles)

	return &config.ConfigAccessor{mergedJson, fl}
}

func (i *Initiator) parseArgs() map[string]string {
	configFilePtr := flag.String("c", "resource/config", "Path to container configuration files")
	startupLogLevel := flag.String("l", "INFO", "Logging threshold for messages from components during bootstrap")
	flag.Parse()

	var params map[string]string
	params = make(map[string]string)

	params["config"] = *configFilePtr
	params["logLevel"] = *startupLogLevel

	return params

}

func (i *Initiator) splitConfigPaths(pathArgument string) []string {
	return strings.Split(pathArgument, ",")
}

func (i *Initiator) builtInConfigPaths() []string {

	const builtInConfigPath = "/resource/facility-config"

	configFolder := config.QuiltHome() + builtInConfigPath

	files, err := config.FindConfigFilesInDir(configFolder)

	if err != nil {
		i.logger.LogFatalf("Unable to load config from folder %s: %s", configFolder, err.Error())
	}

	return files

}
