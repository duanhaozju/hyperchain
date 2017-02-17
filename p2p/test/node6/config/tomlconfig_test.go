package config
import "testing"

func TestReadToml(t *testing.T) {
	ReadToml("./caconfig.toml")
}

func TestWriteConfig(t *testing.T) {
	WriteConfig("./caconfig.toml")
}