package dependencies

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"math"
)

type Cont []struct { //best contributor
	Contributions int `json:"contributions"`
}

// type NCont struct { //nested info about contributor
// 	Contributions int `json:"contributions"`
// 	Id	int `json:"id"`
// }

type Repo struct { //Structure that will recieve important information from REST API request
	URL         string `json:"html_url"`
	NetScore	float64 
	RampUp		float64	
	Correctness float64
	BusFactor float64
	ResponsiveMaintainer float64
	License LName `json:"license"`
	// Name string
}

type LName struct { //substructure to hold nested json fields
	Name string	`json:"name"`
}

type Repos []Repo

func (r *Repos) Search(task string, resp *http.Response, resp1 *http.Response, RU float64, C float64, totalCommits float64, RM float64) {

    var repo Repo
    json.NewDecoder(resp.Body).Decode(&repo) //decodes response and stores info in repo struct
	//fmt.Println(repo.License.Name)

	var cont Cont
    json.NewDecoder(resp1.Body).Decode(&cont) //decodes response and stores info in repo struct
	//fmt.Println(cont[0].Contributions)

	fmt.Println(repo.URL)
	fmt.Println(task)
    new_repo := Repo{ //setting values in repo struct, mostly hard coded for now.
        URL:         repo.URL,
        RampUp:        RU,
        Correctness: C,
        BusFactor: RoundFloat(1 - (float64(cont[0].Contributions) / totalCommits), 3),
        ResponsiveMaintainer: RM,
        License: repo.License,
    }

	var LicenseComp float64
	if (new_repo.License.Name != "") {
		LicenseComp = 1
	} else {
		LicenseComp = 0
	}
	new_repo.NetScore = RoundFloat((LicenseComp*(new_repo.Correctness + 3*new_repo.ResponsiveMaintainer + new_repo.BusFactor+ 2*new_repo.RampUp))/7.0, 3)

    *r = append(*r, new_repo)

}

func (r *Repos) Load(filename string) error { //reads the json
	file, err := os.ReadFile(filename)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}

	if len(file) == 0 {
		return err
	}

	err = json.Unmarshal([]byte(file), r)
	if err != nil {
		return err
	}

	return nil
}

func (r *Repos) Store(filename string) error {
	data, err := json.Marshal(r)
	if err != nil {
		return err
	}
	
	return os.WriteFile(filename, data, 0644)
}


func (r *Repos) Print() {
	for _, repo := range *r {
		fmt.Printf("%s\n", repo.URL)
	}
}

func RoundFloat(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}