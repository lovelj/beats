package webconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
)

type FilebeatConfig struct {
	Maincfgfile   string
	Reloadcfgfile string
	Output        string
	Files         map[string]MonitorFile
}

func (f *FilebeatConfig) ParseInputFromFile(filepath string) {
	f.Files = make(map[string]MonitorFile)
	f.Reloadcfgfile = filepath
	p, err := os.Open(filepath)
	defer p.Close()
	if nil == err {
		buff := bufio.NewReader(p)
		for {
			line, err := buff.ReadString('\n')
			line = strings.TrimSpace(line)

			if err != nil || io.EOF == err {
				break
			}
			fmt.Println(line)
			length := len(line)
			if length < 1 {
				continue
			}

			ind := strings.Index(line, "#")

			if ind == 0 {
				continue
			}

			pre := strings.HasPrefix(line, "- type: log")
			if pre {
				mfile := MonitorFile{}
				for {
					cline, err := buff.ReadString('\n')
					cline = strings.TrimSpace(cline)

					if err != nil || io.EOF == err {
						break
					}
					fmt.Println(cline)
					clength := len(cline)
					if clength < 1 {
						continue
					}

					cind := strings.Index(cline, "#")

					if cind == 0 {
						continue
					}

					phost := strings.HasPrefix(cline, "paths")
					if phost {
						for {
							pline, err := buff.ReadString('\n')

							if err != nil || io.EOF == err {
								break
							}
							pline = strings.TrimSpace(pline)
							ppre := strings.HasPrefix(pline, "#")
							if ppre {
								continue
							}
							if len(pline) < 1 {
								continue
							}
							ppre = strings.HasPrefix(pline, "-")
							if ppre {
								mfile.Path = strings.TrimSpace(substring(pline, 1, len(pline)))

								fmt.Println(mfile.Path)
								break
							}
						}
					}

					if strings.HasPrefix(cline, "logformat") {
						str := strings.TrimSpace(substring(cline, strings.Index(cline, ":")+1, len(cline)))
						mfile.LogType = str
					}
					if strings.HasPrefix(cline, "logpattern") {
						str := strings.TrimSpace(substring(cline, strings.Index(cline, ":")+1, len(cline)))
						mfile.LogPattern = str
					}
					if strings.HasPrefix(cline, "producer") {
						str := strings.TrimSpace(substring(cline, strings.Index(cline, ":")+1, len(cline)))
						mfile.Producer = str
					}
					cpre := strings.HasPrefix(cline, "fields_under_root")
					if cpre {
						break
					}
				}
				f.Files[mfile.Path] = mfile

			}

		}
	}

	//return fd
}

func (f *FilebeatConfig) ParseOutputFromFile(filepath string) {
	p, err := os.Open(filepath)
	f.Maincfgfile = filepath
	defer p.Close()
	if nil == err {
		buff := bufio.NewReader(p)
		for {
			line, err := buff.ReadString('\n')
			line = strings.TrimSpace(line)

			if err != nil || io.EOF == err {
				break
			}
			fmt.Println(line)
			length := len(line)
			if length < 1 {
				continue
			}

			ind := strings.Index(line, "#")

			if ind == 0 {
				continue
			}

			prehosts := strings.HasPrefix(line, "hosts:")
			if prehosts {

				inds := strings.Index(line, "\"")
				inde := strings.LastIndex(line, "\"")
				substr := substring(line, inds+1, inde)

				f.Output = strings.TrimSpace(substr)
				fmt.Println(f.Output)
				break
			}
		}
	}
}

func (f *FilebeatConfig) SaveOutput() {
	content := f.Out2String()
	file, err := os.OpenFile(f.Maincfgfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Open file error!")
		return
	}
	defer file.Close()
	file.WriteString(content)
}
func (f *FilebeatConfig) SaveInputFile() {
	file, err := os.OpenFile(f.Reloadcfgfile, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, os.ModePerm)
	if err != nil {
		fmt.Println("Open file error!")
		return
	}
	defer file.Close()
	splitstr := "\r\n"
	if "windows" == strings.ToLower(runtime.GOOS) {

	} else {
		splitstr = "\n"
	}
	content := ""
	for _, v := range f.Files {
		content = content + splitstr + v.ToString()
	}
	file.WriteString(content)
}

func (f *FilebeatConfig) Out2String() string {
	splitstr := "\r\n"
	if "windows" == strings.ToLower(runtime.GOOS) {

	} else {
		splitstr = "\n"
	}
	str := "filebeat.config.inputs: " + splitstr +
		"  enabled: true" + splitstr +
		"  path: reloadconfig/inputfile.yml" + splitstr +
		"  reload.enabled: true" + splitstr +
		"  reload.period: 5s" + splitstr +
		"setup.template.name: \"logsystem\"" + splitstr +
		"setup.template.pattern: \"logsystem-*\"" + splitstr +
		"setup.template.fields: \"fields.yml\"" + splitstr +
		"setup.template.settings:" + splitstr +
		"  index.number_of_shards: 3" + splitstr +
		"output.elasticsearch: " + splitstr +
		"   " + "hosts: [\"" + f.Output + "\"]" + splitstr +
		"   " + "index: \"logsystem-%{[logdate]}\""
	fmt.Println(str)
	return str
}
func substring(source string, start int, end int) string {

	//return string(str[begin:end])
	//var r = []rune(source)
	length := len(source)

	if start < 0 || end > length || start > end {
		return ""
	}

	if start == 0 && end == length {
		return source
	}

	return string(source[start:end])
}
