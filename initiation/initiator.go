package initiation

import (
	"flag"
	"fmt"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/jsonmerger"
	"github.com/wolferton/quilt/facility/logger"
	"github.com/wolferton/quilt/ioc"
	"os"
	"strings"
)

const initiatorComponentName string = "quiltFrameworkInitiator"
const jsonMergerComponentName string = "quiltJsonMerger"
const configInjectorComponentName string = "quiltConfigInjector"

type Initiator struct {
	logger logger.Logger
}

func (i *Initiator) Start(protoComponents []*ioc.ProtoComponent) {

	if config.QuiltHome() == "" {
		fmt.Println("QUILT_HOME environment variable not set")
		os.Exit(-1)
	}

	var params map[string]string
	params = i.parseArgs()

	bootstrapLogLevel := logger.LogLevelFromLabel(params["logLevel"])

	facilitiesInitialisor := new(FacilitiesInitialisor)

	protoComponents, frameworkLoggingManager := facilitiesInitialisor.BootstrapFrameworkLogging(protoComponents, bootstrapLogLevel)
	i.logger = frameworkLoggingManager.CreateLogger(initiatorComponentName)

	i.logger.LogInfo("Creating framework components")

	configAccessor := i.loadConfigIntoAccessor(params["config"], frameworkLoggingManager)
	facilitiesInitialisor.ConfigAccessor = configAccessor

	injectorLogger := frameworkLoggingManager.CreateLogger(configInjectorComponentName)
	configInjector := config.ConfigInjector{injectorLogger, configAccessor}

	facilitiesInitialisor.ConfigInjector = &configInjector
	facilitiesInitialisor.UpdateFrameworkLogLevel()

	protoComponents = facilitiesInitialisor.InitialiseApplicationLogger(protoComponents)
	protoComponents = facilitiesInitialisor.InitialiseHttpServer(protoComponents, configAccessor, frameworkLoggingManager)
	protoComponents = facilitiesInitialisor.InitialiseQueryManager(protoComponents)

	container := ioc.CreateContainer(protoComponents, frameworkLoggingManager, configAccessor, &configInjector)

	i.logger.LogInfo("Starting components")
	container.StartComponents()
	i.logger.LogInfo("Ready")

}

func (i *Initiator) loadConfigIntoAccessor(configPath string, frameworkLoggingManager *logger.ComponentLoggerManager) *config.ConfigAccessor {
	configFiles := i.builtInConfigPaths()

	expandedPaths, err := config.ExpandToFiles(i.splitConfigPaths(configPath))

	if err != nil {
		i.logger.LogFatal("Unable to load specified config files: " + err.Error())
		return nil
	}

	configFiles = append(configFiles, expandedPaths...)

	if i.logger.IsLevelEnabled(logger.Debug) {

		i.logger.LogDebug("Loading configuration from: ")

		for _, fileName := range configFiles {
			i.logger.LogDebug(fileName)
		}
	}

	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = frameworkLoggingManager.CreateLogger(jsonMergerComponentName)

	mergedJson := jsonMerger.LoadAndMergeConfig(configFiles)

	return &config.ConfigAccessor{mergedJson}
}

func (i *Initiator) parseArgs() map[string]string {
	configFilePtr := flag.String("c", "resources/config", "Path to container configuration files")
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
		i.logger.LogFatal("Unable to load config from folder " + configFolder + " " + err.Error())
	}

	return files

}
