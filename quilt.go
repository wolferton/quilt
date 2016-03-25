package main

import (

    "github.com/wolferton/quilt/cli"
    "fmt"
    "os"
)

const defaultHelpMessage string = "For a list of available commands, run quiltframework LIST"

func main() {

    commands := buildCommands()

    args := os.Args


    if(len(args) == 1){
        fmt.Println("Not enough arguments")
        fmt.Println(defaultHelpMessage)
        os.Exit(-1)
    }

    commandName := args[1]
    command := commands[commandName]

    if(command == nil){
        fmt.Printf("Unknown command %s\n", commandName)
        fmt.Println(defaultHelpMessage)
        os.Exit(-1)
    }

    if(command.Name() == "LIST"){
        listCommandsAndExit(commands)
    } else{
        os.Exit(command.Execute())
    }


}

func listCommandsAndExit(commands map[string]cli.Command){

    for _, command := range commands {

        fmt.Printf("%s\t\t\t%s\n", command.Name(), command.ShortDescription())

    }

   os.Exit(0)

}

func buildCommands() map[string]cli.Command {


    commands := make(map[string]cli.Command)

    createBinding := cli.BuildCreateBindingsCommand()

    commands[createBinding.Name()] = createBinding

    listCommand := new(cli.ListCommand)

    commands[listCommand.Name()] = listCommand

    return commands

}







