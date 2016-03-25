package cli
import (
    "strings"
)

const argPrefix = "-"
const argValSeparator = "="

type ArgumentParser struct {

    argumentValues map[string]string
    argumentAliasCanonical map[string]string

}

func (ap *ArgumentParser) RegisterArgument(canonical string, aliases ...string) {

    ap.argumentAliasCanonical[canonical] = canonical

    for _, alias := range aliases {
        ap.argumentAliasCanonical[alias] = canonical
    }

}

func (ap *ArgumentParser) SetDefault(canonical string, value string) {
    ap.argumentValues[canonical] = value
}

func (ap *ArgumentParser) RegisterFlag(canonical string, aliases ...string) {

}

func (ap *ArgumentParser) StringValue(argumentName string) string {
    return ap.argumentValues[argumentName]
}

func (ap *ArgumentParser) Parse(elements []string) {

    for _, element := range elements {
        if( strings.HasPrefix(element, argPrefix) ){
            ap.extractArgumentValue(element)
        }
    }
}

func (ap *ArgumentParser) extractArgumentValue(element string) {

    argVal := strings.Split(strings.TrimLeft(element, argPrefix), argValSeparator)

    if(len(argVal) == 2) {
        arg := argVal[0]
        val := argVal[1]

        if _, present := ap.argumentAliasCanonical[arg]; present {
            canonicalArgName := ap.argumentAliasCanonical[arg]
            ap.argumentValues[canonicalArgName] = val
        }
    }


}

func CreateArgumentParser() *ArgumentParser {

    parser := new(ArgumentParser)

    parser.argumentAliasCanonical = make(map[string]string)
    parser.argumentValues = make(map[string]string)

    return parser

}