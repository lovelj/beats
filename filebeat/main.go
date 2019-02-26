// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package main

import (
	"os"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"runtime"
	"github.com/drone/routes"
	"github.com/elastic/beats/filebeat/cmd"
	"github.com/elastic/beats/filebeat/webconfig"
)
var filebeatconfig webconfig.FilebeatConfig

func getallfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")

	b, _ := json.Marshal(filebeatconfig)
	fmt.Println(string(b))
	w.Write([]byte(b)) //这个写入到w的是输出到客户端的

	//	params := r.URL.Query()
	//	uid := params.Get(":uid")
	//	fmt.Fprintf(w, "you are get allinfo :%s:", uid)
}
func addfile(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	//w.Header().Set("Content-Type", "application/json")

	//len := r.ContentLength
	//body := make([]byte, len)

	r.ParseForm()
	fmt.Println("Path:", r.Form["Path"][0])
	fmt.Println("LogType:", r.Form["LogType"][0])
	fmt.Println("LogPattern:", r.Form["LogPattern"][0])
	fmt.Println("Producer:", r.Form["Producer"][0])
	mfile := webconfig.MonitorFile{}
	mfile.Path = r.Form["Path"][0]
	mfile.LogType = r.Form["LogType"][0]
	mfile.LogPattern = r.Form["LogPattern"][0]
	mfile.Producer = r.Form["Producer"][0]
	//	result, _ := ioutil.ReadAll(r.Body) //r.Body.Read(body)
	//	r.Body.Close()
	//	fmt.Println(string(result)) ///
	//	var mfile MonitorFile
	//	err := json.Unmarshal(result, &mfile)
	//	if err != nil {
	//		w.Write([]byte("{\"status\":1}"))
	//		return
	//	}

	filebeatconfig.Files[mfile.Path] = mfile
	filebeatconfig.SaveInputFile()
	//	b, _ := json.Marshal(filebeatconfig)
	//	fmt.Println(string(b))
	w.Write([]byte("ok"))
}
func delfile(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	//w.Header().Set("content-type", "application/json")
	r.ParseForm()
	filepath := r.Form["Path"][0]
	//	query := r.URL.Query()
	//	filepath := query["filename"][0]
	//	fmt.Println(filepath)
	delete(filebeatconfig.Files, filepath)

	//	b, _ := json.Marshal(filebeatconfig)
	//	fmt.Println(string(b))
	filebeatconfig.SaveInputFile()
	w.Write([]byte("ok"))
}

func saveoutput(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	//w.Header().Set("content-type", "application/json")
	r.ParseForm()
	output := r.Form["Output"][0]

	fmt.Println(output)
	filebeatconfig.Output = output

	b, _ := json.Marshal(filebeatconfig)
	filebeatconfig.SaveOutput()
	fmt.Println(string(b))
	w.Write([]byte("ok"))
}
func hello(w http.ResponseWriter, r *http.Request) {
	//	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	//	w.WriteHeader(http.StatusOK)
	w.Header().Set("Access-Control-Allow-Origin", "*")             //允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers", "Content-Type") //header的类型
	w.Header().Set("content-type", "application/json")
	w.Write([]byte("{\"Maincfgfile\":\"\"}"))
}
func home(w http.ResponseWriter, r *http.Request) {
	content, _ := readFile("./filebeatwebpage/index.html")
	fmt.Fprintf(w, "%s", content)
}
func readFile(file_name string) ([]byte, error) {
	fi, err := os.Open(file_name)
	if err != nil {
		panic(err)
	}
	defer fi.Close()
	return ioutil.ReadAll(fi)
}

var commands = map[string]string{
	"windows": "cmd /c start",
	"darwin":  "open",
	"linux":   "xdg-open",
}

func Open(url string)  {
	//run, ok := commands[runtime.GOOS]
	//if !ok {
	//	return fmt.Errorf("don't know how to open things on %s platform", runtime.GOOS)
	//}
	//
	//cmd := exec.Command(run, uri)
	//return cmd.Start()
	var err error

	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
		err = fmt.Errorf("unsupported platform")
	}
	if err != nil {
		log.Fatal(err)
	}
}
func startserver(){
	err:=http.ListenAndServe(":54321", nil) //设置监听的端口
	fmt.Println("webserver started")
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
// The basic model of execution:
// - input: finds files in paths/globs to harvest, starts harvesters
// - harvester: reads a file, sends events to the spooler
// - spooler: buffers events until ready to flush to the publisher
// - publisher: writes to the network, notifies registrar
// - registrar: records positions of files read
// Finally, input uses the registrar information, on restart, to
// determine where in each file to restart a harvester.
func main() {
	filebeatconfig = webconfig.FilebeatConfig{}

	filebeatconfig.ParseInputFromFile("./reloadconfig/inputfile.yml")
	filebeatconfig.ParseOutputFromFile("./filebeat.yml")

	mux := routes.New()
	mux.Get("/getallfile", getallfile)
	mux.Post("/delfile", delfile)
	mux.Get("/hello", hello)
	mux.Post("/addfile", addfile)
	mux.Post("/saveoutput", saveoutput)
	http.Handle("/", mux)
	http.HandleFunc("/index", home)

	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("./filebeatwebpage/css"))))
	http.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("./filebeatwebpage/img"))))
	//	http.Handle("/img/", http.FileServer(http.Dir("D:\\workspace\\goweb\\filebeatwebpage\\img")))
	//	http.Handle("/css/", http.FileServer(http.Dir("D:\\workspace\\goweb\\filebeatwebpage\\css")))
	//	http.Handle("/js/", http.FileServer(http.Dir("D:\\workspace\\goweb\\filebeatwebpage\\js")))
	//	http.HandleFunc("/", sayhelloName)       //设置访问的路由
	go  startserver() //设置监听的端口
	//fmt.Println("webserver started")
	//if err != nil {
	//	log.Fatal("ListenAndServe: ", err)
	//}
	Open("http://localhost:54321/index")

	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}

}
