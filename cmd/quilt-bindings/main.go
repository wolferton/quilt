package main

import (
	"flag"
	"github.com/wolferton/quilt/config"
	"github.com/wolferton/quilt/facility/logger"
	"os"
	"path"
	"bufio"
	"strconv"
	"strings"
	"github.com/wolferton/quilt/facility/jsonmerger"
	"fmt"
)

const (
	cdfLocationFlagName string = "c"
	cdfLocationDefault string = "resource/components"
	cdfLocationHelp string = "A comma separated list of config files or directories containing component definition files"

	ofLocationFlagName string = "o"
	ofLocationDefault string = "bindings/bindings.go"
	ofLocationHelp string = "Path of the Go source file that will be generated"

	llFlagName string = "l"
	llDefault string = "ERROR"
	llHelp string = "Minimum importance of logging to be displayed (TRACE, DEBUG, INFO, WARN, ERROR, FATAL)"

	nameField = "name"
	typeField = "type"

	deferSeparator = ":"
	refPrefix = "ref"
	refAlias = "r"
	confPrefix = "conf"
	confAlias = "c"
)

func main() {

	var cdf = flag.String(cdfLocationFlagName, cdfLocationDefault, cdfLocationHelp)
	var of = flag.String(ofLocationFlagName, ofLocationDefault, ofLocationHelp)
	var ll = flag.String(llFlagName, llDefault, llHelp)

	flag.Parse()

	command := new(CreateBindingsCommand)

	expandedFileList, err := config.ExpandToFiles(splitConfigPaths((*cdf)))

	if err != nil {
		fmt.Println("Unable to expand " + *cdf + " to a list of config files: " + err.Error())
		os.Exit(-1)
	}

	command.ComponentDefinitions = expandedFileList
	command.OutputFile = *of
	command.LogLevel = *ll

	command.Execute()

}

func splitConfigPaths(pathArgument string) []string {
	return strings.Split(pathArgument, ",")
}


type CreateBindingsCommand struct {
	logger	logger.Logger
	OutputFile string
	ComponentDefinitions []string
	LogLevel string
}


func (cbc *CreateBindingsCommand) Execute() int {

	cbc.configureLogging()

	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = cbc.logger

	mergedConfig := jsonMerger.LoadAndMergeConfig(cbc.ComponentDefinitions)

	configAccessor := config.ConfigAccessor{mergedConfig}

	cbc.writeBindingsSource(cbc.OutputFile, &configAccessor)

	return 0
}

func (cbc *CreateBindingsCommand) configureLogging() {
	logLevel := logger.LogLevelFromLabel(cbc.LogLevel)

	logger := new(logger.LevelAwareLogger)


	logger.SetThreshold(logLevel)
	logger.SetLoggerName("")
	cbc.logger = logger
}

func (cbc *CreateBindingsCommand) writeBindingsSource(outPath string, configAccessor *config.ConfigAccessor) {

	cbc.logger.LogInfo("Writing binding file " + outPath)

	os.MkdirAll(path.Dir(outPath), 0777)
	file, err := os.Create(outPath)

	if (err != nil) {
		cbc.logger.LogFatal(err.(*os.PathError).Error())
		os.Exit(-1)
	}

	defer file.Close()

	writer := bufio.NewWriter(file)
	cbc.writeImports(writer, configAccessor)

	components := configAccessor.ObjectVal("components")
	componentCount := len(components)

	writer.WriteString("func Components() []*ioc.ProtoComponent {\n")

	writer.WriteString("\tpc := make([]*ioc.ProtoComponent, ")
	writer.WriteString(strconv.Itoa(componentCount))
	writer.WriteString(")\n")

	index := 0

	for name, componentJson := range components {
		component := componentJson.(map[string]interface{})

		writer.WriteString("\n\t//")
		writer.WriteString(name)
		writer.WriteString("\n")

		instanceVariableName := cbc.writeInstance(writer, configAccessor, component, name)
		componentProtoName := cbc.writeComponentWrapper(writer, configAccessor, component, name, index, instanceVariableName)

		for fieldName, fieldContents := range component {

			if !cbc.reservedFieldName(fieldName) {

				switch config.JsonType(fieldContents) {
				case config.JsonMap:
					cbc.writeMapValue(writer, instanceVariableName, fieldName, fieldContents.(map[string]interface{}))
				case config.JsonString:
					cbc.writeStringValue(writer, instanceVariableName, fieldName, fieldContents.(string), componentProtoName)
				case config.JsonUnknown:
					cbc.logger.LogError("Unknown JSON type for field " + fieldName + " on component " + name)
				}

			}

		}

		writer.WriteString("\n\n")
		index = index + 1
	}

	writer.WriteString("\treturn pc\n")
	writer.WriteString("}\n")
	writer.Flush()
}

func (cbc *CreateBindingsCommand) writeStringValue(writer *bufio.Writer, instanceName string, fieldName string, fieldContents string, componentProtoName string) {

	valueElements := strings.SplitN(fieldContents, deferSeparator, 2)

	if len(valueElements) == 2 {

		prefix := valueElements[0]
		instruction := valueElements[1]



		if prefix == refPrefix || prefix == refAlias {

			writer.WriteString("\t")
			writer.WriteString(componentProtoName)
			writer.WriteString(".AddDependency(\"")
			writer.WriteString(fieldName)
			writer.WriteString("\", \"")
			writer.WriteString(instruction)
			writer.WriteString("\")\n")

		} else if prefix == confPrefix || prefix == confAlias {

			writer.WriteString("\t")
			writer.WriteString(componentProtoName)
			writer.WriteString(".AddConfigPromise(\"")
			writer.WriteString(fieldName)
			writer.WriteString("\", \"")
			writer.WriteString(instruction)
			writer.WriteString("\")\n")
		}

	}

}

func (cbc *CreateBindingsCommand) writeMapValue(writer *bufio.Writer, instanceName string, fieldName string, fieldContents map[string]interface{}) {

	directField := instanceName + "." + fieldName

	writer.WriteString(directField)
	writer.WriteString(" = make(map[string]string)\n")

	for key, value := range fieldContents {
		writer.WriteString("\t")
		writer.WriteString(directField)
		writer.WriteString("[\"")
		writer.WriteString(key)
		writer.WriteString("\"] = \"")
		writer.WriteString(value.(string))
		writer.WriteString("\"\n")
	}

}

func (cbc *CreateBindingsCommand) reservedFieldName(field string) bool {
	return field == nameField || field == typeField
}

func (cbc *CreateBindingsCommand) writeComponentWrapper(writer *bufio.Writer, configAccessor *config.ConfigAccessor, component map[string]interface{}, name string, index int, instanceName string) string {

	componentProtoName := name + "Proto"

	writer.WriteString(componentProtoName)
	writer.WriteString(" := ioc.CreateProtoComponent(")
	writer.WriteString(instanceName)
	writer.WriteString(", \"")
	writer.WriteString(name)
	writer.WriteString("\")\n\t")
	writer.WriteString("pc[")
	writer.WriteString(strconv.Itoa(index))
	writer.WriteString("] = ")
	writer.WriteString(componentProtoName);
	writer.WriteString("\n\t")
	writer.WriteString(componentProtoName)
	writer.WriteString(".Component.Name = \"")
	writer.WriteString(name)
	writer.WriteString("\"\n\t")

	return componentProtoName
}

func (cbc *CreateBindingsCommand) writeInstance(writer *bufio.Writer, configAccessor *config.ConfigAccessor, component map[string]interface{}, name string) string {
	instanceType := configAccessor.StringFieldVal("type", component)
	instanceName := name + "Instance"
	writer.WriteString("\t")
	writer.WriteString(instanceName)
	writer.WriteString(" := new(")
	writer.WriteString(instanceType)
	writer.WriteString(")\n\t")

	return instanceName
}

func (cbc *CreateBindingsCommand) writeImports(writer *bufio.Writer, configAccessor *config.ConfigAccessor) {
	packages := configAccessor.Array("packages")

	writer.WriteString("package bindings\n\n")
	writer.WriteString("import (\n")

	for _, packageName := range packages {
		writer.WriteString("\t\"")
		writer.WriteString(packageName.(string))
		writer.WriteString("\"\n")
	}

	writer.WriteString("\t\"github.com/wolferton/quilt/ioc\"")
	writer.WriteString("\n)\n\n")
}

