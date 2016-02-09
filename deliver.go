/*
 *Name: Carter Currin
 *File: deliver.go
 *Description: Continuously pulls Postback objects from the local Redis server on port 6379 for in the following form:
 *  {  
 *     "method":"GET",
 *     "url":"http://sample-domain-endpoint.com/data?key={key}&value={value}&foo={bar}",
 *     "data":
 *     {  
 *        "key":"Azureus",
 *        "value":"Dendrobates"
 *     }
 *  }
 *
 * Once the data is parsed, the url has it's {xxx} elements replaced with the values coresponding to xxx keys in data.
 * If the key has no match in data then a configurable constant is used (default is empty string).
 * The data is then delivered to the endpoint url using the provided method (either GET or POST).
 * A GET delivery would look like:
 *   http://sample-domain-endpoint.com/data?key=Azureus&value=Dendrobates&foo=
 * A POST delivery would look like:
 *   headers:
 *   (POST) http://sample-domain-endpoint.com/data?key=Azureus&value=Dendrobates&foo=
 *   HTTP/1.1 1 1
 *   Content-Type : application/json
 *   body:
 *   {
 *      "bar":"",
 *      "key":"Phyllobates",
 *      "value":"Terribilis"
 *   }
 *
 * Delivery Time, response code, response time, and response body as well as any warnings and error are logged in
 * a log file (default is log file is "deliver.go.log" in same directory as this file).
 */
 
package main

import (
	"fmt"
	"bytes"
	"strings"
	"log"
	"os"
	"io"
	"io/ioutil"
	"regexp"
	"net/http"
	"encoding/json"
	"github.com/garyburd/redigo/redis"
)

const (
	//Configurable default value for unmatched url {keys}
	UNMATCHED_URL_KEY_VALUE = ""
	//set to the name of the Redis List to check for Postback Objects
	REDIS_LIST_NAME = "request"
	//set to where you want the logs written
	LOG_FILE_NAME = "deliver.go.log"
	//set if you want to see trace log outputs in standard out or discarded
	SHOW_TRACES = false
)

var (
	//Regular Expression that matches the imbedded keys with postback urls:
	//http://sample-domain-endpoint.com/data?key={key}&value={value}&foo={bar}
	//(matches)                                  ^^^^^       ^^^^^^^     ^^^^^
	curlyBracketMatch = regexp.MustCompile("{.*?}")
	
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

//Used to store the JSON Postback object input from Redis
type PostbackObject struct {
	Method string              `json:"method"`
	Url    string              `json:"url"`
	Data   map[string]string   `json:"data"`
}

//Initializes the different log levels
//example log format:
//HEADER: yyyy/mm/dd hh:mm:ss.mcrsec /the/full/file/path/deliver.go:line#: Hello World!
func InitLoggers(traceHandle io.Writer, infoHandle io.Writer, warningHandle io.Writer, errorHandle io.Writer) {
	Trace =   log.New(traceHandle,   "TRACE: ",   log.Ldate|log.Lmicroseconds|log.Llongfile)
	Info =    log.New(infoHandle,    "INFO: ",    log.Ldate|log.Lmicroseconds|log.Llongfile)
	Warning = log.New(warningHandle, "WARNING: ", log.Ldate|log.Lmicroseconds|log.Llongfile)
	Error =   log.New(errorHandle,   "ERROR: ",   log.Ldate|log.Lmicroseconds|log.Llongfile)
}

//Finds all the keys in Postback Object's url of the form {xxx} and replaces them with the values they
//    correspond to in the Postback Object's data.
//Precondition: postback has the url to be parsed and the data to replace with.
//Postcondition: the postback's url has {xxx}'s replaced with yyy values from xxx keys from data
func matchUrlKeysToValues(postback *PostbackObject) {
	stringIndexesMatches := curlyBracketMatch.FindStringIndex(postback.Url)
	for stringIndexesMatches != nil {
		bracketMatchString := curlyBracketMatch.FindString(postback.Url)
		matchString := bracketMatchString[1:(len(bracketMatchString) - 1)]
		replaceString, keyHasValue := postback.Data[matchString]
		if !keyHasValue {
			replaceString = UNMATCHED_URL_KEY_VALUE
			postback.Data[matchString] = UNMATCHED_URL_KEY_VALUE
		}
		postback.Url = postback.Url[:stringIndexesMatches[0]] + replaceString + postback.Url[stringIndexesMatches[1]:]
		stringIndexesMatches = curlyBracketMatch.FindStringIndex(postback.Url)
	}
}

//Given a sent PostbackObject and the response received from sending it, response information is logged to file.
//Precondition: postback is initialed to the information delivered to the endpoint
//              response is initialed to the response received from the delivery
//Postcondition: Response time, code, and body are logged to file using Info log level.
func logEndpointResponseInfo(response *http.Response, postback PostbackObject) {
	Info.Println("Received response from: <" + postback.Url + ">")
	Info.Println("response code:", response.StatusCode)
	//Info.Println("response headers:", response.Header)
	body, _ := ioutil.ReadAll(response.Body)
	Info.Println("response body:", string(body))
}

//Delivers a PostbackObject using the HTTP GET method
//Precondition: postback.url is set the the endpoint url to where the postback is delivered and contains the
//              desired GET information.
//Postcondition: The postback info is sent via HTTP GET to postback.url
//               Delivery time is logged using the Info log level.
//               Response info is logged according to logEndpointResponseInfo()
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

//Delivers a PostbackObject using the HTTP POST method
//Precondition: postback.url is set the the endpoint url to where the postback is delivered.
//              postback.data is initialed to the data to be sent via POST
//Postcondition: The postback.data is sent via HTTP POST in JSON form to postback.url
//               Delivery time is logged using the Info log level
//               Response info is logged according to logEndpointResponseInfo()
func deliverPost(postback PostbackObject) {
	requestBody, _ := json.Marshal(postback.Data)
	Trace.Println("requestBody: " + string(requestBody))
	request, err :=  http.NewRequest("POST", postback.Url, bytes.NewBuffer(requestBody))
	request.Header.Set("Content-Type", "application/json")
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

//Processes a Postback object received from the local Redis Server.
//Precondition: redisServer is dialed into a Redis server that is meant to contain postback objects in a 
//              Redis List called REDIS_LIST_NAME
//              Postback Objects received must have one of two acceptable methods: GET or POST
//Postcondition: A Postback Object is read in from the Redis server, its url is parsed with keys being 
//               replaced with values according to matchUrlKeysToValues(), and it is sent to its endpoint
//               via HTTP GET or POST with all relevent information logged.
func processPostbackObject(redisServer redis.Conn) {
	endpoint, err := redis.String(redisServer.Do("LPOP", REDIS_LIST_NAME))
	if err == nil && endpoint != "" {
		postback := PostbackObject{}
		json.Unmarshal([]byte(endpoint), &postback)
		Trace.Println("endpoint: " + endpoint)
		Trace.Println("postback: " + fmt.Sprint(postback))
		matchUrlKeysToValues(&postback)
		Trace.Println("postback.Url: " + postback.Url)
		if strings.ToUpper(postback.Method) == "GET" {
			deliverGet(postback)
		} else if strings.ToUpper(postback.Method) == "POST" {
			deliverPost(postback)
		} else {
			Error.Println("Unsupported Postback Method.")
		}
	} else if fmt.Sprint(err) == "redigo: nil returned" {
		//There is no data yet to deliver.
	} else if err != nil {
		Warning.Println("Redis Problem: " + fmt.Sprint(err))
	} else {
		Warning.Println("Received empty Redis Object.")
	}
}

//Continually processes postback objects from the local Redis database server
//Precondition: The file at LOG_FILE_NAME must be acessable to write logs
//              A local Redis database server must be running on port 6379
//Postcondition: Any Postback objects received from the Redis database's List, REDIS_LIST_NAME, are processed
//               Log levels are set up for the rest of the program to use, all pertinent inforamtion is logged
func main() {
	logFile, logFileError := os.OpenFile(LOG_FILE_NAME, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if logFileError != nil {
		log.Fatalln("Failed to open log file: ", logFileError)
	}
	defer logFile.Close()

	var traceOutput io.Writer
	if SHOW_TRACES {
		traceOutput = os.Stdout
	} else {
		traceOutput = ioutil.Discard
	}
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
