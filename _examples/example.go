package main

import (
	"fmt"
	"time"

	"github.com/lytics/confl"
)

type Config struct {
	Title   string
	Hand    handOfKing
	DB      database `confl:"database"`
	Servers map[string]server
	Clients clients
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
	Name string
	Org  string `confl:"organization"`
	Bio  string
	DOB  time.Time
}

type database struct {
	Server  string
	Ports   []int
	ConnMax int `confl:"connection_max"`
	Enabled bool
}

type server struct {
	IP string
	DC string
}

type clients struct {
	Data  [][]interface{}
	Hosts []string
}

func main() {
	var config Config
	if _, err := confl.DecodeFile("example.conf", &config); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Title: %s\n", config.Title)
	fmt.Printf("Hand: %s (%s, %s), Born: %s\n",
		config.Hand.Name, config.Hand.Org, config.Hand.Bio, config.Hand.DOB)
	fmt.Printf("Database: %s %v (Max conn. %d), Enabled? %v\n",
		config.DB.Server, config.DB.Ports, config.DB.ConnMax, config.DB.Enabled)
	for serverName, server := range config.Servers {
		fmt.Printf("Server: %s (%s, %s)\n", serverName, server.IP, server.DC)
	}
	fmt.Printf("Client data: %v\n", config.Clients.Data)
	fmt.Printf("Client hosts: %v\n", config.Clients.Hosts)
}
