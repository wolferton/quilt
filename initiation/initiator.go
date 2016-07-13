package initiation

import (
	"flag"
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/jsonmerger"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/facility/serviceerror"
	"github.com/wolferton/quilt/ioc"
	"github.com/wolferton/quilt/ws/json"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const initiatorComponentName string = "quiltFrameworkInitiator"
const jsonMergerComponentName string = "quiltJsonMerger"
const configInjectorComponentName string = "quiltConfigInjector"

type Initiator struct {
	logger logger.Logger
}

func (i *Initiator) Start(customComponents []*ioc.ProtoComponent) {

	start := time.Now()

	if config.QuiltHome() == "" {
		fmt.Println("QUILT_HOME environment variable not set")
		os.Exit(-1)
	}

	var params map[string]string
	protoComponents := make(map[string]*ioc.ProtoComponent)

	for _, c := range customComponents {
		protoComponents[c.Component.Name] = c
	}

	params = i.parseArgs()

	bootstrapLogLevel := logger.LogLevelFromLabel(params["logLevel"])

	facilitiesInitialisor := new(FacilitiesInitialisor)

	frameworkLoggingManager := facilitiesInitialisor.BootstrapFrameworkLogging(protoComponents, bootstrapLogLevel)
	i.logger = frameworkLoggingManager.CreateLogger(initiatorComponentName)

	i.logger.LogInfof("Creating framework components")

	configAccessor := i.loadConfigIntoAccessor(params["config"], frameworkLoggingManager)
	facilitiesInitialisor.ConfigAccessor = configAccessor

	injectorLogger := frameworkLoggingManager.CreateLogger(configInjectorComponentName)
	configInjector := config.ConfigInjector{injectorLogger, configAccessor}

	facilitiesInitialisor.ConfigInjector = &configInjector
	facilitiesInitialisor.UpdateFrameworkLogLevel()

	facilitiesInitialisor.InitialiseApplicationLogger(protoComponents)
	facilitiesInitialisor.InitialiseHttpServer(protoComponents)
	facilitiesInitialisor.InitialiseQueryManager(protoComponents)
	facilitiesInitialisor.InitialiseDatabaseAccessor(protoComponents)

	json.InitialiseJsonHttp(frameworkLoggingManager, configAccessor, protoComponents)
	serviceerror.InitialiseServiceErrorManager(frameworkLoggingManager, configAccessor, protoComponents)

	container := ioc.CreateContainer(protoComponents, frameworkLoggingManager, configAccessor, &configInjector)

	i.logger.LogInfof("Starting components")
	err := container.StartComponents()

	if err != nil {
		i.logger.LogFatalf(err.Error())
		i.shutdown(container)
		os.Exit(1)
	} else {
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
}

func (i *Initiator) shutdown(container *ioc.ComponentContainer) {
	i.logger.LogInfof("Shutting down")

	container.ShutdownComponents()

}

func (i *Initiator) loadConfigIntoAccessor(configPath string, frameworkLoggingManager *logger.ComponentLoggerManager) *config.ConfigAccessor {
	configFiles := i.builtInConfigPaths()

	expandedPaths, err := config.ExpandToFiles(i.splitConfigPaths(configPath))

	if err != nil {
		i.logger.LogFatalf("Unable to load specified config files: %s", err.Error())
		return nil
	}

	configFiles = append(configFiles, expandedPaths...)

	if i.logger.IsLevelEnabled(logger.Debug) {

		i.logger.LogDebugf("Loading configuration from: ")

		for _, fileName := range configFiles {
			i.logger.LogDebugf(fileName)
		}
	}

	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = frameworkLoggingManager.CreateLogger(jsonMergerComponentName)

	mergedJson := jsonMerger.LoadAndMergeConfig(configFiles)

	return &config.ConfigAccessor{mergedJson}
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
