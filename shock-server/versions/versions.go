package versions

import (
	"bufio"
	"fmt"
	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/golib/mgo"
	"github.com/MG-RAST/golib/mgo/bson"
	"os"
	"strconv"
)

type Version struct {
	Name    string `bson:"name" json:"name"`
	Version int    `bson:"version" json:"version"`
}

type Versions []Version

var VersionMap = make(map[string]int)

func Initialize() (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Versions")
	c.EnsureIndex(mgo.Index{Key: []string{"name"}, Unique: true})
	var versions = new(Versions)
	err = c.Find(bson.M{}).All(versions)
	for _, v := range *versions {
		VersionMap[v.Name] = v.Version
	}
	return
}

func Print() (err error) {
	fmt.Printf("##### Versions ####\n")
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Versions")
	var versions = new(Versions)
	if err = c.Find(bson.M{}).All(versions); err != nil {
		return err
	}
	for _, v := range *versions {
		fmt.Printf("name: %v\tversion number: %v\n", v.Name, v.Version)
	}
	fmt.Println("")
	return
}

func PushVersionsToDatabase() (err error) {
	session := db.Connection.Session.Copy()
	defer session.Close()
	c := session.DB(conf.MONGODB_DATABASE).C("Versions")
	for k, v := range conf.VERSIONS {
		if _, err = c.Upsert(bson.M{"name": k}, bson.M{"$set": bson.M{"name": k, "version": v}}); err != nil {
			return err
		}
	}
	return
}

func RunVersionUpdates() (err error) {
	// Upgrading databases with ACL schema before version 2
	confVersionACL, ok1 := conf.VERSIONS["ACL"]
	dbVersionACL, ok2 := VersionMap["ACL"]
	if (ok1 && confVersionACL >= 2) && (!ok2 || (ok2 && dbVersionACL < confVersionACL)) {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("The ACL schema version in your database needs updating.  Would you like the update to run? (y/n): ")
		text, _ := consoleReader.ReadString('\n')
		if text[0] == 'y' {
			// Checking database to see if "public" already exists in a Node's ACL's somewhere.
			session := db.Connection.Session.Copy()
			defer session.Close()
			c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
			nCount, err := c.Find(bson.M{}).Count()
			if err != nil {
				return err
			}
			oCount, err := c.Find(bson.M{"acl.owner": "public"}).Count()
			if err != nil {
				return err
			}
			rCount, err := c.Find(bson.M{"acl.read": "public"}).Count()
			if err != nil {
				return err
			}
			wCount, err := c.Find(bson.M{"acl.write": "public"}).Count()
			if err != nil {
				return err
			}
			dCount, err := c.Find(bson.M{"acl.delete": "public"}).Count()
			if err != nil {
				return err
			}
			if oCount > 0 || rCount > 0 || wCount > 0 || dCount > 0 {
				fmt.Println("There are \"public\" strings already present in some of your Node ACL's:")
				fmt.Println("Total number of nodes: " + strconv.Itoa(nCount))
				fmt.Println("Number of nodes with \"public\" in owner ACL: " + strconv.Itoa(oCount))
				fmt.Println("Number of nodes with \"public\" in read ACL: " + strconv.Itoa(rCount))
				fmt.Println("Number of nodes with \"public\" in write ACL: " + strconv.Itoa(wCount))
				fmt.Println("Number of nodes with \"public\" in delete ACL: " + strconv.Itoa(dCount))
				fmt.Print("Your database may be in a mixed state, would you like to run the update anyway? (y/n): ")
				text, _ = consoleReader.ReadString('\n')
			}
			if text[0] == 'y' {
				fmt.Println("Updating ACL's to version: " + strconv.Itoa(confVersionACL))
				if _, err = c.UpdateAll(bson.M{"acl.owner": ""}, bson.M{"$set": bson.M{"acl.owner": "public"}}); err != nil {
					return err
				}
				if _, err = c.UpdateAll(bson.M{"acl.read": bson.M{"$size": 0}}, bson.M{"$push": bson.M{"acl.read": "public"}}); err != nil {
					return err
				}
				if _, err = c.UpdateAll(bson.M{"acl.write": bson.M{"$size": 0}}, bson.M{"$push": bson.M{"acl.write": "public"}}); err != nil {
					return err
				}
				if _, err = c.UpdateAll(bson.M{"acl.delete": bson.M{"$size": 0}}, bson.M{"$push": bson.M{"acl.delete": "public"}}); err != nil {
					return err
				}
				session := db.Connection.Session.Copy()
				defer session.Close()
				c := session.DB(conf.MONGODB_DATABASE).C("Versions")
				if _, err = c.Upsert(bson.M{"name": "ACL"}, bson.M{"$set": bson.M{"name": "ACL", "version": 2}}); err != nil {
					return err
				}
				fmt.Println("ACL schema version update complete.")
			} else {
				os.Exit(0)
			}
		} else {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}

	// Updating databases with Node schema before version 2
	confVersionNode, ok1 := conf.VERSIONS["Node"]
	dbVersionNode, ok2 := VersionMap["Node"]
	if (ok1 && confVersionNode >= 2) && (!ok2 || (ok2 && dbVersionNode < confVersionNode)) {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("The Node schema version in your database needs updating.  Would you like the update to run? (y/n): ")
		text, _ := consoleReader.ReadString('\n')
		if text[0] == 'y' {
			session := db.Connection.Session.Copy()
			defer session.Close()
			c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
			fmt.Println("Updating Nodes to version: " + strconv.Itoa(confVersionNode))
			if _, err = c.UpdateAll(bson.M{}, bson.M{"$unset": bson.M{"public": 1}}); err != nil {
				return err
			}
			fmt.Println("Node schema version update complete.")
		} else {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}
	return
}
