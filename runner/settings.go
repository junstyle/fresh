package runner

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/pilu/config"
)

const (
	envSettingsPrefix   = "RUNNER_"
	mainSettingsSection = "Settings"
)

var settings = map[string]string{
	"config_path":       "./runner.conf",
	"root":              ".",
	"watch_paths":       "", //watched pathes, seperate by semi;
	"tmp_path":          "./tmp",
	"build_name":        "runner-build",
	"build_args":        "",
	"build_log":         "runner-build-errors.log",
	"run_args":          "",
	"valid_ext":         ".go, .tpl, .tmpl, .html",
	"no_rebuild_ext":    ".tpl, .tmpl, .html",
	"ignored":           "assets, tmp",
	"build_delay":       "600",
	"colors":            "1",
	"log_color_main":    "cyan",
	"log_color_build":   "yellow",
	"log_color_runner":  "green",
	"log_color_watcher": "magenta",
	"log_color_app":     "",
}

var colors = map[string]string{
	"reset":          "0",
	"black":          "30",
	"red":            "31",
	"green":          "32",
	"yellow":         "33",
	"blue":           "34",
	"magenta":        "35",
	"cyan":           "36",
	"white":          "37",
	"bold_black":     "30;1",
	"bold_red":       "31;1",
	"bold_green":     "32;1",
	"bold_yellow":    "33;1",
	"bold_blue":      "34;1",
	"bold_magenta":   "35;1",
	"bold_cyan":      "36;1",
	"bold_white":     "37;1",
	"bright_black":   "30;2",
	"bright_red":     "31;2",
	"bright_green":   "32;2",
	"bright_yellow":  "33;2",
	"bright_blue":    "34;2",
	"bright_magenta": "35;2",
	"bright_cyan":    "36;2",
	"bright_white":   "37;2",
}

func logColor(logName string) string {
	settingsKey := fmt.Sprintf("log_color_%s", logName)
	colorName := settings[settingsKey]

	return colors[colorName]
}

func loadEnvSettings() {
	for key, _ := range settings {
		envKey := fmt.Sprintf("%s%s", envSettingsPrefix, strings.ToUpper(key))
		if value := os.Getenv(envKey); value != "" {
			settings[key] = value
		}
	}
}

func loadRunnerConfigSettings() {
	if _, err := os.Stat(configPath()); err != nil {
		return
	}

	logger.Printf("Loading settings from %s", configPath())
	sections, err := config.ParseFile(configPath(), mainSettingsSection)
	if err != nil {
		return
	}

	for key, value := range sections[mainSettingsSection] {
		settings[key] = value
	}
}

func initSettings() {
	loadEnvSettings()
	loadRunnerConfigSettings()
}

func getenv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func root() string {
	return settings["root"]
}

func watchPaths() []string {
	s := strings.TrimSpace(settings["watch_paths"])
	if s != "" {
		r := root()
		ps := []string{}
		for _, ss := range strings.Split(s, ",") {
			ps = append(ps, strings.TrimSpace(ss))
		}
		if !inArray(ps, r) {
			ps = append(ps, r)
		}
		return ps
	} else {
		return []string{}
	}
}

func tmpPath() string {
	return settings["tmp_path"]
}

func buildName() string {
	return settings["build_name"]
}
func buildPath() string {
	p := filepath.Join(tmpPath(), buildName())
	if runtime.GOOS == "windows" && filepath.Ext(p) != ".exe" {
		p += ".exe"
	}
	return p
}

func buildArgs() []string {
	// s := regexp.MustCompile(`(-\w+\s+|-\w+\s*$)`).ReplaceAllString(settings["build_args"], "|^|$1|^|")
	// args := []string{}
	// for _, a := range strings.Split(s, "|^|") {
	// 	if strings.TrimSpace(a) != "" {
	// 		args = append(args, strings.TrimSpace(a))
	// 	}
	// }
	// return args
	args, _ := parseCommandLine(settings["build_args"])
	return args
}

func buildErrorsFileName() string {
	return settings["build_log"]
}

func buildErrorsFilePath() string {
	return filepath.Join(tmpPath(), buildErrorsFileName())
}

// runArgs 运行app时的参数
func runArgs() []string {
	// s := regexp.MustCompile(`(-\w+\s+|-\w+\s*$)`).ReplaceAllString(settings["run_args"], "|^|$1|^|")
	// args := []string{}
	// for _, a := range strings.Split(s, "|^|") {
	// 	if strings.TrimSpace(a) != "" {
	// 		args = append(args, strings.TrimSpace(a))
	// 	}
	// }
	// return args
	args, _ := parseCommandLine(settings["run_args"])
	return args
}

func configPath() string {
	return settings["config_path"]
}

func buildDelay() time.Duration {
	value, _ := strconv.Atoi(settings["build_delay"])

	return time.Duration(value)
}

func inArray(a []string, s string) bool {
	for _, item := range a {
		if item == s {
			return true
		}
	}
	return false
}

func parseCommandLine(command string) ([]string, error) {
	command = strings.TrimSpace(command)

	var args []string
	state := "start"
	current := ""
	quote := "\""
	escapeNext := true
	for _, c := range command {

		if state == "quotes" {
			if string(c) != quote {
				current += string(c)
			} else {
				args = append(args, current)
				current = ""
				state = "start"
			}
			continue
		}

		if escapeNext {
			current += string(c)
			escapeNext = false
			continue
		}

		if c == '\\' {
			escapeNext = true
			continue
		}

		if c == '"' || c == '\'' {
			state = "quotes"
			quote = string(c)
			continue
		}

		if state == "arg" {
			if c == ' ' || c == '\t' {
				args = append(args, current)
				current = ""
				state = "start"
			} else {
				current += string(c)
			}
			continue
		}

		if c != ' ' && c != '\t' {
			state = "arg"
			current += string(c)
		}
	}

	if state == "quotes" {
		return []string{}, errors.New(fmt.Sprintf("Unclosed quote in command line: %s", command))
	}

	if current != "" {
		args = append(args, current)
	}

	return args, nil
}
