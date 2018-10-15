package sidecar

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"
)

func TestIfConfigTemplateIsResolved(t *testing.T) {
	//given
	envoyConfig := EnvoyConfig{
		serviceName:    "lorem",
		instanceId:     "lorem-12345",
		bindIp:         "127.1.2.3",
		bindPort:       "778",
		logsDir:        "/sandbox-lorem-12345/envoy-logs",
		baseConfigPath: "../testdata/envoy-config.yaml",
	}

	//when
	config, err := envoyConfig.resolveConfigTemplate()

	//then
	assert.NoError(t, err)
	configStr := string(config)
	assert.Contains(t, configStr, "port_value: 778")
	assert.Contains(t, configStr, "for instance: lorem-12345")
	assert.Contains(t, configStr, "/sandbox-lorem-12345/envoy-logs/envoy-access.log")
	assert.NotContains(t, configStr, "{{.")
}

func TestIfGenerateBaseId(t *testing.T) {
	//given
	instance1 := "instanceId-abc21334eeuie323239023uieh"
	instance2 := "instanceId-jefuwh8w9e8944uiwehhiuhiuh"

	//when
	baseIdInstance1 := generateBaseId(instance1)
	baseIdInstance2 := generateBaseId(instance2)
	baseIdInstance1Again := generateBaseId(instance1)

	//then
	assert.Regexp(t, "[0-9]+", baseIdInstance1)
	assert.Regexp(t, "[0-9]+", baseIdInstance2)
	assert.Equal(t, baseIdInstance1, baseIdInstance1Again)
	assert.NotEqual(t, baseIdInstance1, baseIdInstance2)
}

func TestIfPrepareEnvoy(t *testing.T) {
	//given
	baseConfig, err := ioutil.ReadFile("../testdata/envoy-config.yaml")
	assert.NoError(t, err)
	os.MkdirAll("../target/test-output/sandbox-ipsum-1/envoy/", 0755)
	assert.NoError(t, err)
	err = ioutil.WriteFile("../target/test-output/sandbox-ipsum-1/envoy/envoy-config-template.yaml", baseConfig, 0644)
	assert.NoError(t, err)

	expectedCommand := []string{
		"../target/test-output/sandbox-ipsum-1/envoy/envoy",
		"-c", "envoy-config.yaml",
		"--service-cluster", "ipsum",
		"--service-node", "ipsum-1",
		"--log-path", "../target/test-output/sandbox-ipsum-1/envoy/logs/envoy-output.log",
		"--base-id", "3200151633",
	}

	//when
	cmd, err := PrepareEnvoy("ipsum", "ipsum-1", "127.1.20.1", "21009", "../target/test-output/sandbox-ipsum-1")
	defer os.RemoveAll("../target/test-output/sandbox-ipsum-1")

	//then
	assert.NoError(t, err)
	content, err := ioutil.ReadFile("../target/test-output/sandbox-ipsum-1/envoy/envoy-config.yaml")
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
	assert.Equal(t, expectedCommand, cmd)
}