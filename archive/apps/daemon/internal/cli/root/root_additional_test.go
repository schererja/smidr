package root

import (
	"os"
	"testing"

	"github.com/spf13/viper"
)

func TestInitConfig_NoFile(t *testing.T) {
	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(oldwd)

	viper.Set("verbose", true)
	initConfig() // should not panic or print error when config is missing
}

func TestInitConfig_WithFile_Verbose(t *testing.T) {
	tmp := t.TempDir()
	oldwd, _ := os.Getwd()
	_ = os.Chdir(tmp)
	defer os.Chdir(oldwd)

	// create a minimal valid yaml file
	if err := os.WriteFile("smidr.yaml", []byte("name: test\n"), 0o644); err != nil {
		t.Fatalf("write yaml: %v", err)
	}
	viper.Set("verbose", true)
	initConfig() // should read config and (with verbose) print Using config file
}
