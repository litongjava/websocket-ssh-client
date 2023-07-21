package config

type Config struct {
  App *App `yaml:"app"`
}

type App struct {
  Host     string `yaml:"host"`
  Port     int    `yaml:"port"`
  EndPoint string `yaml:"endPoint"`
}
