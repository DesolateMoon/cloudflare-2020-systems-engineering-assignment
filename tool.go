package main

import (
	"fmt"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
//	"strings"
	"time"
)

/*########################################
  ########################################
  Structs
  (Profile)
  ########################################
  ########################################
*/
//Will profile all the necessary data from the website
type Profile struct {
	timeElapsed		time.Duration
	response		httpResponse 
}

/*########################################
  ########################################
  Threaded Helpers
  (handleRequests)
  ########################################
  ########################################
*/
//Threaded helper that will request the 
//website and store necessary data
func handleRequests(request chan Profile, newURL uniformResourceLocator) {

	/*-- Final Variables --*/
	URLREDIRECT		:= 300
	URLBADREQUEST	:= 400
	PARSEHEADER		:= "Location"	

	profile 		:= Profile{}
	getRequest 		:= httpRequest{"GET", newURL.domain, newURL.subDirectory}
	start 			:= time.Now()
	data 			:= sendGetRequest(getRequest)
	elapsed			:= time.Since(start)
	response		:= responseParser(data)

	//if response is valid and not redirected record the necessary data
	if !(response.status >= URLREDIRECT && response.status < URLBADREQUEST) {
		profile.response = response
		profile.timeElapsed = elapsed
		request <- profile
	} else {
		redirectURL, _ := redirectParser(response.header[PARSEHEADER])
		go handleRequests(request, redirectURL)
	}
}

/*########################################
  ########################################
  Main Wrapper Logic
  ########################################
  ########################################
*/
//CLI Tool which will make an HTTP request
//to the url and either prints the entire
//response or a summary of statistics to 
//the console. Will also enable flags such
//as --help, --url, and --profile. 
func main() {

	/*-- Final Variables --*/
	SUCCESSFULEXIT 	:= 0
	NONSUCCESSEXIT 	:= 1
	SUCESSSTATUS 	:= 200
	HELPFLAG		:= "--help"
	PROFILEFLAG 	:= "--profile"
	URLFLAG			:= "--url"

	/*-- Flag Variables --*/
	args 			:= os.Args[1:]
	url 			:= ""
	numRequest 		:= 1
	doStats	 		:= false
	index 			:= 0

	/*######################################################
      Parses all args flags from input while error handling
      ######################################################
    */
	for index = 0; index < len(args); index++ {
		switch arg := args[index]; arg {
			case HELPFLAG:
				fmt.Printf("Usage: go run . --profile <NUM_REQUEST> --url <URL>\n")
				fmt.Printf("--profile <NUM_REQUEST> number of profile requests to be made\n")
				fmt.Printf("--url <URL> Case sensitive link that will be requested to print the statistics")
				os.Exit(SUCCESSFULEXIT)

			case PROFILEFLAG:
				if len(args) < index + 1 {
					fmt.Printf("Please enter missing <REQUEST_COUNT> after --profile flag\n")
					os.Exit(NONSUCCESSEXIT)
				}

				count, err := strconv.Atoi(args[index+1])

				if err != nil {
					fmt.Printf("<REQUEST_COUNT> must be a number, please try again\n")
					os.Exit(NONSUCCESSEXIT)
				}

				numRequest = count
				doStats = true

			case URLFLAG:
				if len(args) < index + 1 {
					fmt.Printf("Please enter missing <URL> after --url flag\n")
					os.Exit(NONSUCCESSEXIT)
				}

				/* TODO: Has some edgecases that still need to be handled*/
				/* https://stackoverflow.com/questions/161738/what-is-the-best-regular-expression-to-check-if-a-string-is-a-valid-url */
				//url is case sensitive
				url = args[index+1]		
				isURL,_ := regexp.MatchString(
					`[-a-zA-Z0-9@:%._\+~#=]{1,256}\.[a-zA-Z0-9()]{1,6}\b([-a-zA-Z0-9()@:%_\+.~#?&//=]*)`, 
					url)

				if isURL == false {
					fmt.Printf("Please enter acceptable <URL>, it is not in acceptable form\n")
					os.Exit(NONSUCCESSEXIT)
				} else if url == "" {
					fmt.Printf("Please enter missing <URL> by --url <URL> or refer to --help for more info\n")
					os.Exit(NONSUCCESSEXIT)
				}
		}
	}

	/*-- Request & Statistics Variables --*/
	newURL, _ 		:= uniformResourceLocatorParser(url)
	requests 		:= make(chan Profile, numRequest)
	profile		 	:= make([]Profile, 0, numRequest)
	times 			:= make([]int, 0, numRequest)
	size 			:= make([]int, 0, numRequest)
	errors 			:= make([]int, 0, numRequest)
	
	/*#######################################
      Delegate to Multi-Thread helper method
      #######################################
    */
	index = 0
	for index = 0; numRequest > index; index++  {
		go handleRequests(requests, newURL)
	}
	
	/*#######################################
      Organize data and spit out error codes
      #######################################
    */
	index = 0
	for request := range requests {
		profile = append(profile, request)
		index++
		
		if numRequest <= index {
			close(requests)
		}
	}

	index = 0
	if doStats == false {
		fmt.Println("**--- Retrieving Body Data for " + url + " ---**")
		fmt.Printf("%s", profile[index].response.body)
		os.Exit(SUCCESSFULEXIT)
	}

	for _, profData := range profile {
		if SUCESSSTATUS != profData.response.status {
			errors = append(errors, profData.response.status)
		}
		times = append(times, int(profData.timeElapsed.Milliseconds()))
		size = append(size, profData.response.size)
		index = index + int(profData.timeElapsed.Milliseconds())
	}

	/*##############################
      Logic to Output Statistics
      ##############################
    */
	sort.Ints(times)
	sort.Ints(size)
	mid 			:= int(math.Floor(float64(len(times))/2.0))
	
	fastestTime 	:= times[ 0 ]
	slowestTime 	:= times[ len(times)-1 ]
	medianTime 		:= times[ mid ]
	meanTime 		:= index / numRequest
	okRequests 		:= numRequest - len(errors)
	percentSuccess 	:= float64(okRequests / numRequest) * 100.0
	smallestBytes 	:= size[ 0 ]
	largestBytes 	:= size[ len(size)-1 ]


	/*##############################
      Outputs Statistics to Console
      ##############################
    */
	fmt.Println("**--- Profiling Statistics for " + url + " ---**")
	fmt.Println("Number of Requests:", numRequest)
	fmt.Printf("Fastest Time: %d (ms)\n", fastestTime)
	fmt.Printf("Slowest Time: %d (ms)\n", slowestTime)
	fmt.Printf("Median Time: %d (ms)\n", medianTime)
	fmt.Printf("Mean Time: %d (ms)\n", meanTime)
	fmt.Printf("Percent of successful requests: %.f%%\n", percentSuccess)
	if len(errors) > 0 {
		fmt.Println("Error Codes:", errors)
	} else {
		fmt.Println("Error codes: None")
	}
	fmt.Printf("Size of the smallest response: %d (bytes)\n", smallestBytes)
	fmt.Printf("Size of the largest response: %d (bytes)", largestBytes)
}