// Copyright 2016 Palantir Technologies, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package integration_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"regexp"
	"testing"
	"time"

	"github.com/nmiyake/archiver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/palantir/godel/apps/distgo/pkg/osarch"
	"github.com/palantir/godel/framework/godellauncher"
	"github.com/palantir/godel/framework/pluginapi"
)

var echoPluginTmpl = fmt.Sprintf(`#!/bin/sh
if [ "$1" = "%s" ]; then
    echo '%s'
    exit 0
fi

echo $@
`, pluginapi.InfoCommandName, `%s`)

func TestPlugins(t *testing.T) {
	pluginName := fmt.Sprintf("tester-integration-%d", time.Now().Unix())

	testProjectDir := setUpGödelTestAndDownload(t, testRootDir, gödelTGZ, version)
	src := `package main

	import "fmt"

	func main() {
		fmt.Println("hello, world!")
	}
`
	err := ioutil.WriteFile(path.Join(testProjectDir, "main.go"), []byte(src), 0644)
	require.NoError(t, err)

	cfgDir, err := godellauncher.ConfigDirPath(testProjectDir)
	require.NoError(t, err)

	cfg, err := godellauncher.ReadGödelConfig(cfgDir)
	require.NoError(t, err)

	cfgContent := fmt.Sprintf(`
plugins:
  resolvers:
    - %s/repo/{{GroupPath}}/{{Product}}/{{Version}}/{{Product}}-{{OS}}-{{Arch}}-{{Version}}.tgz
  plugins:
    - locator:
        id: "com.palantir:%s:1.0.0"
`, testProjectDir, pluginName)
	err = yaml.Unmarshal([]byte(cfgContent), &cfg)
	require.NoError(t, err)

	pluginDir := path.Join(testProjectDir, "repo", "com", "palantir", pluginName, "1.0.0")
	err = os.MkdirAll(pluginDir, 0755)
	require.NoError(t, err)

	pluginInfo := pluginapi.MustNewInfo("com.palantir", pluginName, "1.0.0", "echo.yml", pluginapi.MustNewTaskInfo(
		"echo-task",
		"Echoes input",
		pluginapi.TaskInfoGlobalFlagOptions(pluginapi.NewGlobalFlagOptions(
			pluginapi.GlobalFlagOptionsParamDebugFlag("--debug"),
			pluginapi.GlobalFlagOptionsParamProjectDirFlag("--project-dir"),
			pluginapi.GlobalFlagOptionsParamGodelConfigFlag("--godel-config"),
			pluginapi.GlobalFlagOptionsParamConfigFlag("--config"),
		)),
		pluginapi.TaskInfoCommand("echo"),
		pluginapi.TaskInfoVerifyOptions(pluginapi.NewVerifyOptions(
			pluginapi.VerifyOptionsApplyFalseArgs("--verify"),
		)),
	))
	pluginInfoJSON, err := json.Marshal(pluginInfo)
	require.NoError(t, err)

	pluginScript := path.Join(pluginDir, pluginName+"-1.0.0")
	err = ioutil.WriteFile(pluginScript, []byte(fmt.Sprintf(echoPluginTmpl, string(pluginInfoJSON))), 0755)
	require.NoError(t, err)

	pluginTGZPath := path.Join(pluginDir, fmt.Sprintf("%s-%s-1.0.0.tgz", pluginName, osarch.Current()))
	err = archiver.TarGz(pluginTGZPath, []string{pluginScript})
	require.NoError(t, err)

	cfgBytes, err := yaml.Marshal(cfg)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(cfgDir, godellauncher.GödelConfigYML), cfgBytes, 0644)
	require.NoError(t, err)

	// plugin is resolved on first run
	gotOutput := execCommand(t, testProjectDir, "./godelw", "version")
	wantOutput := "(?s)" + regexp.QuoteMeta(fmt.Sprintf(`Getting package from %s/repo/com/palantir/%s/1.0.0/%s-%s-1.0.0.tgz...`, testProjectDir, pluginName, pluginName, osarch.Current())) + ".+"
	assert.Regexp(t, wantOutput, gotOutput)

	gotOutput = execCommand(t, testProjectDir, "./godelw", "echo-task", "foo", "--bar", "baz")
	wantOutput = fmt.Sprintf("--project-dir %s --godel-config %s/godel/config/godel.yml --config %s/godel/config/echo.yml echo foo --bar baz\n", testProjectDir, testProjectDir, testProjectDir)
	assert.Equal(t, wantOutput, gotOutput)

	gotOutput = execCommand(t, testProjectDir, "./godelw", "verify", "--skip-test")
	wantOutput = fmt.Sprintf(`Running format...
Running generate...
Running imports...
Running license...
Running check...
Running compiles...
Running deadcode...
Running errcheck...
Running extimport...
Running golint...
Running govet...
Running importalias...
Running ineffassign...
Running nobadfuncs...
Running novendor...
Running outparamcheck...
Running unconvert...
Running varcheck...
Running echo-task...
--project-dir %s --godel-config %s/godel/config/godel.yml --config %s/godel/config/echo.yml echo
`, testProjectDir, testProjectDir, testProjectDir)
	assert.Equal(t, wantOutput, gotOutput)

	gotOutput = execCommand(t, testProjectDir, "./godelw", "verify", "--skip-test", "--apply=false")
	wantOutput = fmt.Sprintf(`Running format...
Running generate...
Running imports...
Running license...
Running check...
Running compiles...
Running deadcode...
Running errcheck...
Running extimport...
Running golint...
Running govet...
Running importalias...
Running ineffassign...
Running nobadfuncs...
Running novendor...
Running outparamcheck...
Running unconvert...
Running varcheck...
Running echo-task...
--project-dir %s --godel-config %s/godel/config/godel.yml --config %s/godel/config/echo.yml echo --verify
`, testProjectDir, testProjectDir, testProjectDir)
	assert.Equal(t, wantOutput, gotOutput)
}
