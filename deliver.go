package main

import (
	"fmt"
	"log"
	"os"
	"io"
	"bytes"
	"strings"
	"regexp"
	"flag"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

var (
	home   = os.Getenv("HOME")
	user   = os.Getenv("USER")
	gopath = os.Getenv("GOPATH")
	
	curlyBracketMatch = regexp.MustCompile("{.*?}")
	
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

type PostbackObject struct {
	Method string              `json:"method"`
	Url    string              `json:"url"`
	Data   map[string]string   `json:"data"`
}

func Init() {
	if user == "" {
		log.Fatalln("$USER not set")
	}
	if home == "" {
		home = "/home/" + user
	}
	if gopath == "" {
		gopath = home + "/go"
	}
	// gopath may be overridden by --gopath flag on command line.
	flag.StringVar(&gopath, "gopath", gopath, "override default GOPATH")
}

func InitLoggers(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace =   log.New(traceHandle,   "TRACE: ",   log.Ldate|log.Lmicroseconds|log.Llongfile)
	Info =    log.New(infoHandle,    "INFO: ",    log.Ldate|log.Lmicroseconds|log.Llongfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Lmicroseconds|log.Llongfile)
	Error =   log.New(errorHandle,   "ERROR: ",   log.Ldate|log.Lmicroseconds|log.Llongfile)
}

func replaceUrlWithData(postback *PostbackObject) {
	stringIndexesMatches := curlyBracketMatch.FindStringIndex(postback.Url)
	for stringIndexesMatches != nil {
		bracketMatchString := curlyBracketMatch.FindString(postback.Url)
		matchString := bracketMatchString[1:(len(bracketMatchString) - 1)]
		replaceString := postback.Data[matchString]
		postback.Url = postback.Url[:stringIndexesMatches[0]] + replaceString + postback.Url[stringIndexesMatches[1]:]
		stringIndexesMatches = curlyBracketMatch.FindStringIndex(postback.Url)
	}
}

func logEndpointResponseInfo(response *http.Response, postback PostbackObject) {
	Info.Println("Received response from: <" + postback.Url + ">")
	Info.Println("response code:", response.StatusCode)
	//Info.Println("response headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	Info.Println("response body:", string(body))
}

func deliverGet(postback PostbackObject) {
	Info.Println("Delivering. url: <" + postback.Url + "> method: " + postback.Method)
	response, err := http.Get(postback.Url)
	if err != nil {
		Warning.Println("Could not send GET request to: <" + postback.Url + ">")
	} else {
		defer response.Body.Close()
		logEndpointResponseInfo(response, postback)
	}
}

func deliverPost(postback PostbackObject) {
	requestBody, _ := json.Marshal(postback.Data)
	Trace.Println("requestBody: " + string(requestBody))
	request, err :=  http.NewRequest("POST", postback.Url, bytes.NewBuffer(requestBody))
	//request.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	Trace.Println("request: " + fmt.Sprint(request))
	Info.Println("Delivering. url: <" + postback.Url + "> method: " + postback.Method)
	response, err := client.Do(request)
	if err != nil {
		Warning.Println("Could not send POST request to: <" + postback.Url + ">")
	} else {
		defer response.Body.Close()
		logEndpointResponseInfo(response, postback)
	}
}

func processPostbackObject(redisServer redis.Conn) {
	endpoint, err := redis.String(redisServer.Do("LPOP", "request"))
	if err == nil && endpoint != "" {
		postback := PostbackObject{}
		json.Unmarshal([]byte(endpoint), &postback)
		Trace.Println("endpoint: " + endpoint)
		Trace.Println("postback: " + fmt.Sprint(postback))
		replaceUrlWithData(&postback)
		Trace.Println("postback.Url: " + postback.Url)
		if strings.ToUpper(postback.Method) == "GET" {
			deliverGet(postback)
		} else if strings.ToUpper(postback.Method) == "POST" {
			deliverPost(postback)
		} else {
			Error.Println("Unsupported Postback Method.")
		}
	} else if endpoint == "" {
		//Warning.Println("Received empty Redis Object.")
	} else {
		Warning.Println("Redis Problem: " + fmt.Sprint(err))
	}
}

func main() {
	logFile, logFileError := os.OpenFile("deliver.go.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if logFileError != nil {
		log.Fatalln("Failed to open lof file: ", logFileError)
	}
	defer logFile.Close()

	traceOutput := os.Stdout //use "ioutil.Discard" to ignore Trace logs
	infoOutput := io.MultiWriter(logFile, os.Stdout)
	warningOutput := io.MultiWriter(logFile, os.Stdout)
	errorOutput := io.MultiWriter(logFile, os.Stderr)
	InitLoggers(traceOutput, infoOutput, warningOutput, errorOutput)
	
	redisServer, err := redis.Dial("tcp", ":6379")
	if err != nil {
		Error.Fatalln(err)
	}
	defer redisServer.Close()
	for {
		processPostbackObject(redisServer)
	}
}
