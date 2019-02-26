package webconfig

import (
	"runtime"
	"strings"
)

type MonitorFile struct {
	Path       string
	Producer   string
	LogType    string
	LogPattern string
}

func (m *MonitorFile) ToString() string {
	adfpattern := "^\\d{2}\\:\\d{2}\\:\\d{2}"
	log4jpattern := "^\\d{4}\\-\\d{2}\\-\\d{2}"

	splitstr := "\r\n"
	if "windows" == strings.ToLower(runtime.GOOS) {

	} else {
		splitstr = "\n"
	}
	pattern := ""
	cfgpattern := ""
	if m.LogType == "adflog" {
		pattern = adfpattern

	} else {
		pattern = log4jpattern
		cfgpattern = "      logpattern  : " + m.LogPattern + splitstr
	}

	return "- type: log" + splitstr +
		"  enabled: true" + splitstr +
		"  tail_files: true " + splitstr +
		"  paths:  " + splitstr +
		"    - " + m.Path + splitstr +
		"  multiline.pattern: " + pattern + splitstr +
		"  multiline.negate: true" + splitstr +
		"  multiline.match: after" + splitstr +
		"  multiline.timeout: 10s" + splitstr +
		"  encoding: utf-8" + splitstr +
		"  fields: " + splitstr +
		"      logformat  : " + m.LogType + splitstr +
		"      producer  : " + m.Producer + splitstr +
		cfgpattern +
		"  fields_under_root: true" + splitstr

	return ""

}
