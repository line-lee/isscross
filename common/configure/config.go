package configure

var SystemConfig *Config

type Config struct {
	Port int `yaml:"port"`
}
