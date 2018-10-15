package sidecar

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const (
	envoyDir            = "envoy"
	logsDir             = envoyDir + "/logs"
	accessLogFile       = "envoy-access.log"
	envoyLogFile        = "envoy-output.log"
	envoyBaseConfigFile = "envoy-config-template.yaml"
	envoyConfigFile     = "envoy-config.yaml"
	envoyBinary         = "envoy"
)

func PrepareEnvoy(
	serviceName string, instanceId string, bindIp string, bindPort string,
	sandboxDir string) ([]string, error) {

	logsAbsDir := filepath.Join(sandboxDir, logsDir)

	config := &EnvoyConfig{
		serviceName:    serviceName,
		instanceId:     instanceId,
		bindIp:         bindIp,
		bindPort:       bindPort,
		baseConfigPath: filepath.Join(sandboxDir, envoyDir, envoyBaseConfigFile),
		configPath:     filepath.Join(sandboxDir, envoyDir, envoyConfigFile),
		logsDir:        logsAbsDir,
	}

	if err := config.createConfigFile(); err != nil {
		return nil, errors.Wrap(err, "Preparing config for Envoy failed.")
	}

	os.MkdirAll(logsAbsDir, 0755)

	return []string{
		filepath.Join(sandboxDir, envoyDir, envoyBinary),
		"-c", envoyConfigFile,
		"--service-cluster", serviceName,
		"--service-node", instanceId,
		"--log-path", filepath.Join(logsAbsDir, envoyLogFile),
		"--base-id", generateBaseId(instanceId),
	}, nil
}

type EnvoyConfig struct {
	serviceName    string
	instanceId     string
	bindIp         string
	bindPort       string
	baseConfigPath string
	logsDir        string
	configPath     string
}

type EnvoyConfigVariables struct {
	ListenerAddress string
	ListenerPort    string
	AccessLogPath   string
	InstanceId      string
}

func generateBaseId(instanceId string) string {
	checksum := md5.Sum([]byte(instanceId))
	return fmt.Sprint(binary.BigEndian.Uint32(checksum[:4]))
}

func (c *EnvoyConfig) resolveConfigTemplate() ([]byte, error) {
	config, err := ioutil.ReadFile(c.baseConfigPath)
	if err != nil {
		return nil, errors.Wrap(err, "Cannot read file")
	}

	tpl, err := template.New("config").Parse(string(config))
	if err != nil {
		return nil, errors.Wrap(err, "Cannot parse template")
	}
	vars := EnvoyConfigVariables{
		ListenerAddress: c.bindIp,
		ListenerPort:    c.bindPort,
		AccessLogPath:   filepath.Join(c.logsDir, accessLogFile),
		InstanceId:      c.instanceId,
	}

	var buffer bytes.Buffer
	if err := tpl.Execute(&buffer, vars); err != nil {
		return nil, errors.Wrap(err, "Cannot resolve template")
	}
	return buffer.Bytes(), nil
}

func (c *EnvoyConfig) createConfigFile() error {
	configJson, err := c.resolveConfigTemplate()
	if err != nil {
		return errors.Wrapf(err, "Invalid envoy base config file (%v)", c.baseConfigPath)
	}

	if err := ioutil.WriteFile(c.configPath, configJson, 0644); err != nil {
		return errors.Wrapf(err, "Cannot write to envoy config file (%s)", c.configPath)
	}
	return nil
}
