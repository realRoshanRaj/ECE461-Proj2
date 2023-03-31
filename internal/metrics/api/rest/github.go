package rest

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

var token string

func init() {
	token = os.Getenv("GITHUB_TOKEN")
}

type Repository struct {
	PullRequests struct {
		PageInfo struct {
			EndCursor   githubv4.String
			HasNextPage bool
		}
		Edges []struct {
			Node struct {
				Additions githubv4.Int
				Deletions githubv4.Int
			}
		}
	} `graphql:"pullRequests(states: MERGED, first: 100, after: $pullRequestCursor)"`
}

type Response struct {
	Repository Repository `graphql:"repository(owner: $repositoryOwner, name: $repositoryName)"`
}

func GetTotalChanges(url, token string) (int, error) {
	//given github url, split it to get owner and name
	link := strings.Split(url, "/")
	owner := link[len(link)-2]
	name := link[len(link)-1]

	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubv4.NewClient(httpClient)

	variables := map[string]interface{}{
		"repositoryOwner":   githubv4.String(owner),
		"repositoryName":    githubv4.String(name),
		"pullRequestCursor": (*githubv4.String)(nil),
	}

	var totalAdditions int
	var totalDeletions int
	for {
		var query Response
		err := client.Query(context.Background(), &query, variables)
		if err != nil {
			return 0, err
		}

		for _, pr := range query.Repository.PullRequests.Edges {
			totalAdditions += int(pr.Node.Additions)
			totalDeletions += int(pr.Node.Deletions)
		}

		if !query.Repository.PullRequests.PageInfo.HasNextPage {
			break
		}

		variables["pullRequestCursor"] = githubv4.NewString(query.Repository.PullRequests.PageInfo.EndCursor)
	}

	totalChanges := totalAdditions - totalDeletions

	return totalChanges, nil
}

// TODO: change the log printf functions to new log
func GetPullRequestsResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] + "/pulls?state=closed" //converting github repo url to API url
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	pr_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	//defer pr_resp.Body.Close()

	return pr_resp
}

func GetPullRequestResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	req, err := http.NewRequest(http.MethodGet, httpUrl, nil)
	if err != nil {
		log.Fatalln(err)
	}

	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	pr_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}

	return pr_resp
}

func GetCodeTabResponse(httpurl string) string {
	//client := &http.Client{}

	link := strings.Split(httpurl, "https://github.com/")
	Code_tab_link := "https://api.codetabs.com/v1/loc?github=" + link[len(link)-1] //converting github repo url to API url

	return Code_tab_link
}

func GetRepoResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is to the repository main page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] //converting github repo url to API url
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitH-ub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo_resp.Body.Close()

	/* Dumps the contents of the body of the request and the response
	*  into readable formats as in the html
	 */
	// LOGGING STUFF FOR DEBUGGING HTTP REQUESTS AND RESPONSES
	responseDump, err := httputil.DumpResponse(repo_resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Response: %s", responseDump)
	// Here the 0666 is the same as chmod parameters in linux
	// os.WriteFile("responseDumpRepo.log", responseDump, 0666) // Deprecated
	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	_requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	// os.WriteFile("requestDumpRepo.log", requestDump, 0666) // Deprecated
	log.Printf("Request: %s", _requestDump)
	// storeLog(log_file, requestDump, "Repo request dump\n", false)
	// storeLog(log_file, responseDump, "Repo response dump\n", false)

	return repo_resp
}

func GetContributorResponse(httpUrl string) *http.Response {
	client := &http.Client{}

	// Make sure the URL is the contributors page
	link := strings.Split(httpUrl, "https://github.com/")
	REST_api_link := "https://api.github.com/repos/" + link[len(link)-1] + "/contributors" //converting github repo url to API url
	// fmt.Println(REST_api_link)
	req, err := http.NewRequest(http.MethodGet, REST_api_link, nil)
	if err != nil {
		log.Fatalln(err)
	}
	req.Header.Add("Authorization", "Bearer "+token)

	// Make the GET request to the GitHub API
	repo_resp, err := client.Do(req)
	if err != nil {
		log.Fatalln(err)
	}
	defer repo_resp.Body.Close()

	// LOGGING STUFF FOR DEBUGGING HTTP REQUESTS AND RESPONSES
	responseDump, err := httputil.DumpResponse(repo_resp, true)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Response: %s", responseDump)
	// Here the 0666 is the same as chmod parameters in linux
	// os.WriteFile(log_file, responseDump, 0666) // Deprecated
	// This will DUMP your AUTHORIZATION token be careful! add to .gitignore if you haven't already
	_requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		log.Fatalln(err)
	}
	log.Printf("Request: %s", _requestDump)
	// os.WriteFile("requestDumpContributor.log", requestDump, 0666) // Deprecate

	// storeLog(log_file, requestDump, "Contributor request dump\n", true)
	// storeLog(log_file, responseDump, "Contributor response dump\n", true)

	return repo_resp
}
