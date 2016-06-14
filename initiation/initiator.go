package initiation
import (
    "flag"
    "github.com/wolferton/quilt/config"
    "github.com/wolferton/quilt/ioc"
    "github.com/wolferton/quilt/facility/logger"
    "fmt"
    "os"
    "strings"
    "github.com/wolferton/quilt/facility/jsonmerger"
)

const initiatorComponentName string = "quiltFrameworkInitiator"
const jsonMergerComponentName string = "quiltJsonMerger"

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

    var configPath = params["config"]


    configFiles := i.builtInConfigPaths()
    configFiles = append(configFiles, i.splitConfigPaths(configPath)...)

    i.logger.LogInfo("Loading configuration from: ")

    for _, fileName := range configFiles {
        i.logger.LogInfo(fileName)
    }


    jsonMerger := new(jsonmerger.JsonMerger)
    jsonMerger.Logger = frameworkLoggingManager.CreateLogger(jsonMergerComponentName)

    mergedJson := jsonMerger.LoadAndMergeConfig(configFiles)

    configAccessor := config.ConfigAccessor{mergedJson}

	facilitiesInitialisor.UpdateFrameworkLogLevel(frameworkLoggingManager, &configAccessor)

	protoComponents = facilitiesInitialisor.InitialiseApplicationLogger(protoComponents, &configAccessor, frameworkLoggingManager)
    protoComponents = facilitiesInitialisor.InitialiseHttpServer(protoComponents, &configAccessor, frameworkLoggingManager)

    container := ioc.CreateContainer(protoComponents, frameworkLoggingManager, &configAccessor)

    i.logger.LogInfo("Starting components")
    container.StartComponents()
    i.logger.LogInfo("Ready")

}

func (i *Initiator) parseArgs() (map[string]string) {
    configFilePtr := flag.String("c", "config.json", "Path to container configuration file")
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

    const builtInConfigPath = "/resource/conf"

    configFolder := config.QuiltHome() + builtInConfigPath

    files, err := config.FindConfigFilesInDir(configFolder)

    if err != nil {
        i.logger.LogFatal("Unable to load config from folder " + configFolder + " " + err.Error())
    }

    return files

}
