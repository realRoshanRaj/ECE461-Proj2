package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	gq "pkgmanager/internal/metrics/api/graphql"
	"pkgmanager/internal/metrics/api/rest"
	dep "pkgmanager/pkg/utils"
	"regexp"
	"sort"
	"strconv"
	"strings"
	// These are dependencies must be installed with go get make sure in makefile
	// "github.com/joho/godotenv"
)

const (
	output_json = "output.ndjson"
)

var token string
var log_file string
var log_level int
var repos *dep.Repos

// func init() {
// 	// Loads token into environment variables along with other things in the .env file
// 	// godotenv.Load(".env")
// 	var err error
// 	token = os.Getenv("GITHUB_TOKEN")
// 	if err != nil {
// 		log.Fatal(err, "couldn't find GITHUB_TOKEN environment variable")
// 	}
// 	log_file = os.Getenv("LOG_FILE")
// 	if err != nil {
// 		log.Fatal(err, "couldn't find LOG_FILE environment variable")
// 	}
// 	// Clears file
// 	empty := []byte {};
// 	storeLog(log_file, empty , "", true)

//		log_level , err = strconv.Atoi(os.Getenv("LOG_LEVEL"))
//		if err != nil {
//			log.Fatal(err, "couldn't find LOG_LEVEL environment variable")
//		}
//		repos = &dep.Repos{}
//	}
func main() {

	//init

	var err error
	token = os.Getenv("GITHUB_TOKEN")

	// fmt.Println(token)
	log_file = os.Getenv("LOG_FILE")

	// Clears file
	empty := []byte{}
	storeLog(log_file, empty, "", true)

	log_level, err = strconv.Atoi(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.Fatal(err, "couldn't find LOG_LEVEL environment variable")
	}
	repos = &dep.Repos{}

	//

	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Printf("Please enter ./run help for help\n")
		os.Exit(0)
	}

	// Expects File path to be first arguement
	urlfile, err := os.Open(args[0])
	if err != nil {
		log.Fatal(err)
	}
	defer urlfile.Close()

	// Read URLS from the file
	var urls []string
	scanner := bufio.NewScanner(urlfile)
	for scanner.Scan() {
		// fmt.Println(scanner.Text())
		urls = append(urls, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// For each URL fetch data
	for i := 0; i < len(urls); i++ {
		//if url is npm turn into github url

		convertUrl(&urls[i])

		// Used for Graphql
		split_url := strings.Split(urls[i], "/")
		repo_owner := split_url[3]
		repo_name := split_url[4]

		// fmt.Printf("SPLIT URL: %s\n", split_url)
		// fmt.Printf("REPO OWNER: %s\n", repo_owner)
		// fmt.Printf("REPO NAME: %s\n", repo_name)

		// Gets HTTP response from Rest API

		repo_resp := rest.GetRepoResponse(urls[i]) // repository data
		// fmt.Println(token)

		contri_resp := rest.GetContributorResponse(urls[i]) //contributor data

		totalCommits, err := rest.GetTotalCommitsInMergedPRs(repo_owner, repo_name, token)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Total Commits with PR: ", totalCommits)

		numCommits, err := rest.GetNumCommits(repo_owner, repo_name, token)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Num Commits: ", numCommits)

		fraction := float64(totalCommits) / float64(numCommits)
		fmt.Println("Fraction: ", fraction)

		// Gets Intermediate metric values from Graphql NOT FINAL SCORES
		metrics := gq.Graphql_func(repo_owner, repo_name, token)

		// Inserts the metrics into final function to do math on them and make a new struct out of them

		repos.Construct(repo_resp, contri_resp, metrics[0], metrics[1], metrics[2], metrics[3], metrics[4])

		if log_level >= 2 {
			log.Println(urls[i])
		}
	}

	sort.SliceStable((*repos), func(i, j int) bool {
		return (*repos)[i].NET_SCORE > (*repos)[j].NET_SCORE
	})

	repos.Print()
	repos.Store(output_json)
}

// Converts npm url to github url
func convertUrl(url *string) {
	if strings.HasPrefix(*url, "https://www.npmjs") {
		// data, err := exec.Command("node", "giturl.js", *url).Output()
		// if err != nil {
		// 	fmt.Println("Error:", err)
		// } else {
		// 	fmt.Println("URL: ", *url)
		// 	*url = strings.TrimSuffix(string(data), "\n")
		// 	fmt.Println("URL: ", *url)
		// }
		rawgithubURL := rest.GetGithubURL(*url)

		gitLinkMatch := regexp.MustCompile(".*github.com/(.*).git")
		parsed := gitLinkMatch.FindStringSubmatch(rawgithubURL)[1]
		*url = "https://github.com/" + parsed
	}
}

func storeLog(filename string, data []byte, header string, clear bool) error {
	var f *os.File
	var err error

	if clear {
		f, err = os.OpenFile(log_file, os.O_CREATE|os.O_WRONLY, 0644)
	} else {
		f, err = os.OpenFile(log_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	}

	if err != nil {
		log.Println(err)
	}
	defer f.Close()

	var logger *log.Logger = log.New(f, header, log.LstdFlags)
	if log_level >= 1 {
	} else {
		logger.SetFlags(0)
		logger.SetOutput(io.Discard)
	}

	logger.Println(string(data))
	return err
}
