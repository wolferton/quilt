package cli

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
)

const (
	componentConfArgName = "conf"
	componentConfDefault = "conf/components.json"
	bindingsOutFileArgName = "out"
	bindingsOutFileDefault = "bindings/bindings.go"

	nameField = "name"
	typeField = "type"

	deferSeparator = ":"
	refPrefix = "ref"
	refAlias = "r"
	confPrefix = "conf"
	confAlias = "c"
)

type CreateBindingsCommand struct {
	name             string
	shortDescription string
	logger           logger.Logger
}

func (cbc *CreateBindingsCommand) Name() string {
	return cbc.name
}

func (cbc *CreateBindingsCommand) ShortDescription() string {
	return cbc.shortDescription
}

func BuildCreateBindingsCommand() Command {

	command := new(CreateBindingsCommand)
	command.name = "bind"
	command.shortDescription = "Generate Go source files to add custom components to the Quilt IOC container."

	return command
}

func (cbc *CreateBindingsCommand) Execute() int {

	argumentParser := cbc.createArgumentParser()
	flag.Parse()


	argumentParser.Parse(flag.Args())

	confPath := argumentParser.StringValue(componentConfArgName)
	outPath := argumentParser.StringValue(bindingsOutFileArgName)

	logLevelLabel := argumentParser.StringValue(LogLevelArgName)
	cbc.configureLogging(logLevelLabel)

	jsonMerger := new(jsonmerger.JsonMerger)
	jsonMerger.Logger = cbc.logger

	mergedConfig := jsonMerger.LoadAndMergeConfig([]string{confPath})

	configAccessor := config.ConfigAccessor{mergedConfig}

	cbc.writeBindingsSource(outPath, &configAccessor)

	return 0
}

func (cbc *CreateBindingsCommand) createArgumentParser() *ArgumentParser {
	argumentParser := CreateArgumentParser()
	argumentParser.RegisterArgument(componentConfArgName, "c")
	argumentParser.SetDefault(componentConfArgName, componentConfDefault)

	argumentParser.RegisterArgument(bindingsOutFileArgName, "o")
	argumentParser.SetDefault(bindingsOutFileArgName, bindingsOutFileDefault)

	argumentParser.RegisterArgument(LogLevelArgName, "l")
	argumentParser.SetDefault(LogLevelArgName, logger.ErrorLabel)

	return argumentParser
}

func (cbc *CreateBindingsCommand) configureLogging(logLevelLabel string) {
	logLevel := logger.LogLevelFromLabel(logLevelLabel)

	logger := new(logger.LevelAwareLogger)


	logger.SetThreshold(logLevel)
	logger.SetLoggerName(cbc.name)
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

