package main

import (
    "fmt"
    "gopkg.in/yaml.v2"
    "io/ioutil"
    "os"
    "github.com/davecgh/go-spew/spew"
)

type Location struct {
	ID          string `bson:"ID" json:"ID" yaml:"ID" `                   // e.g. ANLs3 or local for local store
	Description string `bson:"Description" json:"Description" yaml:"Description"` // e.g. ANL official S3 service
	Type        string `bson:"type" json:"type" yaml:"Type" `               // e.g. S3
	URL         string `bson:"url" json:"url" yaml:"URL"`                 // e.g. http://s3api.invalid.org/download&id=
	Prefix      string `bson:"prefix" json:"-" yaml:"Prefix"`                // e.g.g S3 Bucket or username
	AuthKey     string `bson:"AuthKey" json:"-" yaml:"AuthKey"`               // e.g.g AWS auth-key
	SecretKey   string `bson:"SecretKey" json:"-" yaml:"SecretKey" `             // e.g.g AWS secret-key
}
type Config struct {
   Locations []Location `bson:"Locations" json:"Locations" yaml:"Locations" ` 
}

func main() {
    filename := os.Args[1]
    var config Config
    source, err := ioutil.ReadFile(filename)
    if err != nil {
        panic(err)
    }
    err = yaml.Unmarshal(source, &config)
    if err != nil {
        panic(err)
    }

    spew.Dump(config)

    var Locations map[string]*Location
    _ = Locations
    Locations = make( map[string]*Location )

    for i, _ := range config.Locations {
     loc := &config.Locations[i]

    Locations[loc.ID]=loc
    spew.Dump(Locations[loc.ID])
}
//    spew.Dump(Locations)


//    Locations{"anls3_anlseq"}
    fmt.Printf("Value: %#v\n", Locations["anls3_anlseq"])
    fmt.Printf("Value: %#v\n", Locations["anls3_anlseq"].SecretKey)
}
