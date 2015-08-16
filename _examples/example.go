package main

import (
	"fmt"
	"time"

	"github.com/araddon/gou"

	"github.com/lytics/confl"
)

type Config struct {
	Title       string
	Hand        handOfKing
	Location    address `confl:"address"`
	Seenwith    map[string]character
	Seasons     []string
	Description string
}

/*
hand {
  name = "Tyrion"
  organization = "Lannisters"
  bio = "Imp"                 // comments on fields
  dob = 1979-05-27T07:32:00Z  # dates, and more comments on fields
}
*/
type handOfKing struct {
	Name     string
	Org      string `confl:"organization"`
	Bio      string
	DOB      time.Time
	Deceased bool
}

type address struct {
	Street  string
	City    string
	Region  string
	ZipCode int
}

type character struct {
	Episode string
	Season  string
}

func main() {
	gou.SetupLogging("debug")
	gou.SetColorOutput()

	var config Config
	if _, err := confl.DecodeFile("example.conf", &config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Title: %s\n", config.Title)
	fmt.Printf("Hand: %s %s, %s. Born: %s, Deceased? %v\n",
		config.Hand.Name, config.Hand.Org, config.Hand.Bio, config.Hand.DOB, config.Hand.Deceased)
	fmt.Printf("Location: %#v\n", config.Location)
	for name, person := range config.Seenwith {
		fmt.Printf("Seen With: %s (%s, %s)\n", name, person.Episode, person.Season)
	}
	fmt.Printf("Seasons: %v\n", config.Seasons)
	fmt.Printf("Description: %v\n", config.Description)
}
