package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

/*########################################
  ########################################
  Structs
  (httpRequest, httpResponse, 
   uniformResourceLocator)
  ########################################
  ########################################
*/
//Will hold all data for the httpRequest
type httpRequest struct {
	get 			string
	domain 			string
	subDirectory 	string
}

//Will hold all data for the httpResponse
type httpResponse struct {
	header 			map[string]string
	body 			string
	size 			int
	status  		int
}

//Will hold all data for the URL
type uniformResourceLocator struct {
	domain 			string
	subDirectory	string
}

/*########################################
  ########################################
  Helper Functions
  (redirectParser, parseResponse, 
   parseURL, sendRequest)
  ########################################
  ########################################
*/
//Parses redirect to get the proper redirect URL
func redirectParser(header string) (uniformResourceLocator, error) {
	subDirectory 	:= "/"
	temp 			:= regexp.MustCompile("^(https://|http://)")
	newHeader 		:= temp.ReplaceAllString(header, "")
	domain 			:= newHeader
	location 		:= strings.Index(newHeader, "/")

	//checks if location is valid
	if location != -1 {
		domain = newHeader[:location]
		subDirectory = newHeader[location:]
	}	

	return uniformResourceLocator{domain, subDirectory}, nil
}

//Parses data to get the proper HTTP Response
func responseParser(data []byte) httpResponse {
	offset			:= 4
	markdown		:= "\r\n\r\n"
	header 			:= make(map[string]string)
	input			:= bufio.NewScanner(strings.NewReader(string(data)))
	firstTime 		:= true
	var status int
	
	for input.Scan() {
		line := input.Text()
		if firstTime == false {
			temp := strings.SplitN(line, ": ", 2)
			if len(temp) >= 2 {
				header[temp[0]] = temp[1]
			} else {
				break
			}
		} else {
			temp := strings.Split(line, " ")
			status, _ = strconv.Atoi(temp[1])
		}
		firstTime = false
	}

	index := offset + strings.Index(string(data), markdown)

	return httpResponse{header, string(data)[index:],  len(data), status}
}

//Parses URL, case sensitive
func uniformResourceLocatorParser(url string) (uniformResourceLocator, error) {
	subDirectory 	:= "/"
	temp 			:= regexp.MustCompile("^(https://|http://|HTTPS://|HTTP://)")
	newURL 			:= temp.ReplaceAllString(url, "")
	base 			:= newURL
	location 		:= strings.Index(newURL, subDirectory)

	if location != -1 {
		subDirectory = newURL[location:]
		base = newURL[:location]
	}	

	return uniformResourceLocator{base, subDirectory}, nil
}

//Sends get request using crpyto/tls library
func sendGetRequest(getRequest httpRequest) []byte {
	/*-- Final Variables --*/
	NONSUCCESSEXIT 	:= 1
	PARSETIME		:= "3s"
	TCP				:= "tcp"
	HTTPS			:= ":https"
	HTTP			:= "HTTP/1.0"
	MARKDOWN		:= "\r\n"
	HOST			:= "Host: "

	timeout, _ := time.ParseDuration(PARSETIME)
	dialer := net.Dialer{
		Timeout: timeout,
	}
	connection, err := tls.DialWithDialer(&dialer, TCP, getRequest.domain + HTTPS, nil)

	if err != nil {
		fmt.Printf("Error %s\n", err.Error())
		os.Exit(NONSUCCESSEXIT)
	}

	defer connection.Close()

	connection.Write([]byte(getRequest.get + " " + getRequest.subDirectory + " " + HTTP + MARKDOWN))
	connection.Write([]byte(HOST + getRequest.domain + MARKDOWN))
	connection.Write([]byte(MARKDOWN))

	line, err := ioutil.ReadAll(connection)
	
	if err != nil {
		fmt.Printf("Error %s\n", err.Error())
		os.Exit(NONSUCCESSEXIT)
	}

	return line
}