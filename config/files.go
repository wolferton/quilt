package config
import (
    "io/ioutil"
    "strings"
)

func FindConfigFilesInDir(dirPath string) ([]string, error) {
    return findConfigFilesInDir(dirPath, []string{})
}

func findConfigFilesInDir(dirPath string, filesFound []string) ([]string, error){

    contents, err := ioutil.ReadDir(dirPath)

    files := make([]string, 0)


    if err != nil {
        return nil, err
    }

    for _, info := range contents{

        fileName := info.Name()

        if info.Mode().IsDir() {

        } else if strings.HasSuffix(fileName, ".json") {

            files = append(files, dirPath + "/" + fileName)
        }

    }

    return files, nil
}