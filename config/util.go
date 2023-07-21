package config

import (
  "fmt"
  "gopkg.in/yaml.v2"
  "io/ioutil"
  "log"
)

var CONFIG *Config

func ReadFile(filename string) {
  yamlFile, err := ioutil.ReadFile(filename)
  if err != nil {
    log.Println(err.Error())
  }

  err = yaml.Unmarshal(yamlFile, &CONFIG)
  if err != nil {
    fmt.Println("error", err.Error())
  }
}
