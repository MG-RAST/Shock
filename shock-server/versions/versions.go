package versions

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"time"

	"github.com/MG-RAST/Shock/shock-server/conf"
	"github.com/MG-RAST/Shock/shock-server/db"
	"github.com/MG-RAST/Shock/shock-server/node"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
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
	// get ACL versions
	confVersionACL, ok1 := conf.VERSIONS["ACL"]
	dbVersionACL, ok2 := VersionMap["ACL"]

	// skip version updates if database is empty / new shock deploy
	session := db.Connection.Session.Copy()
	c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
	num, err := c.Count()
	session.Close()
	if err != nil {
		return err
	}
	if num == 0 {
		return nil
	}

	// Upgrading databases with ACL schema before version 2
	if (ok1 && confVersionACL >= 2) && (!ok2 || (ok2 && dbVersionACL < 2)) {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("The ACL schema version in your database needs updating to version 2.  Would you like the update to run? (y/n): ")
		//text, _ := consoleReader.ReadString('\n')
		text := ""
		if conf.FORCE_YES {
			text = "y"
		} else {
			text, _ = consoleReader.ReadString('\n')
		}
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
				if conf.FORCE_YES {
					text = "y"
				} else {
					text, _ = consoleReader.ReadString('\n')
				}
			}

			if text[0] == 'y' {
				fmt.Println("Updating ACL's to version 2")
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
				fmt.Println("ACL schema version 2 update complete.")
			} else {
				os.Exit(0)
			}
		} else {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}

	// get Node versions
	confVersionNode, ok1 := conf.VERSIONS["Node"]
	dbVersionNode, ok2 := VersionMap["Node"]

	// Updating databases with Node schema before version 2
	// removing 'public' field
	if (ok1 && confVersionNode >= 2) && (!ok2 || (ok2 && dbVersionNode < 2)) {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("The Node schema version in your database needs updating to version 2.  Would you like the update to run? (y/n): ")
		text := ""
		if conf.FORCE_YES {
			text = "y"
		} else {
			text, _ = consoleReader.ReadString('\n')
		}
		if text[0] == 'y' {
			session := db.Connection.Session.Copy()
			defer session.Close()
			c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
			fmt.Println("Updating Nodes to version 2")
			if _, err = c.UpdateAll(bson.M{}, bson.M{"$unset": bson.M{"public": 1}}); err != nil {
				return err
			}
			fmt.Println("Node schema version 2 update complete.")
		} else {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}

	// Updating databases with Node schema before version 3
	// move parts info from file in node dir to part of node document
	if (ok1 && confVersionNode >= 3) && (!ok2 || (ok2 && dbVersionNode < 3)) {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("The Node schema version in your database needs updating to version 3.  Would you like the update to run? (y/n): ")
		text := ""
		if conf.FORCE_YES {
			text = "y"
		} else {
			text, _ = consoleReader.ReadString('\n')
		}
		if text[0] == 'y' {
			// query for parts nodes with no md5sum
			var n = new(node.Node)
			updated := 0
			query := bson.M{"$and": []bson.M{bson.M{"type": "parts"}, bson.M{"file.checksum.md5": bson.M{"$exists": false}}}}
			// get node iter
			session := db.Connection.Session.Copy()
			defer session.Close()
			c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
			iter := c.Find(query).Iter()
			defer iter.Close()
			for iter.Next(n) {
				//fmt.Println("Processing node: " + n.Id)
				pfile, perr := ioutil.ReadFile(n.Path() + "/parts/parts.json")
				// have file and no parts in node document - fix it
				if (perr == nil) && (n.Parts == nil) {
					pl := &node.PartsList{}
					if err = json.Unmarshal(pfile, &pl); err != nil {
						return err
					}
					if err = os.RemoveAll(n.Path() + "/parts/parts.json"); err != nil {
						return err
					}
					n.Parts = pl
					n.Save()
					updated += 1
				}
			}
			fmt.Println(fmt.Sprintf("Node schema version 3 update complete, updated %d nodes.", updated))
		} else {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}

	// Updating databases with Node schema before version 4
	// add timestamp for file info and each index info
	if (ok1 && confVersionNode >= 4) && (!ok2 || (ok2 && dbVersionNode < 4)) {
		consoleReader := bufio.NewReader(os.Stdin)
		fmt.Print("The Node schema version in your database needs updating to version 4.  Would you like the update to run? (y/n): ")
		utext := ""
		if conf.FORCE_YES {
			utext = "y"
		} else {
			utext, _ = consoleReader.ReadString('\n')
		}
		if utext[0] == 'y' {
			fmt.Print("Would you like to update node info with timestamps from disk (otherwise current time is used)? (y/n): ")
			ftext := ""
			if conf.FORCE_YES {
				ftext = "y"
			} else {
				ftext, _ = consoleReader.ReadString('\n')
			}
			// query for all nodes with a file md5sum
			var n = new(node.Node)
			updated := 0
			query := bson.M{"$and": []bson.M{bson.M{"file.checksum.md5": bson.M{"$exists": true}}, bson.M{"file.checksum.md5": bson.M{"$ne": ""}}}}
			// get node iter
			session := db.Connection.Session.Copy()
			defer session.Close()
			c := session.DB(conf.MONGODB_DATABASE).C("Nodes")
			iter := c.Find(query).Iter()
			defer iter.Close()
			for iter.Next(n) {
				//fmt.Println("Processing node: " + n.Id)
				if ftext[0] == 'y' {
					// update file info from disk
					if fileStat, ferr := os.Stat(n.FilePath()); ferr == nil {
						n.File.CreatedOn = fileStat.ModTime()
					} else {
						n.File.CreatedOn = time.Now()
					}
				} else {
					// update file info with current
					n.File.CreatedOn = time.Now()
				}
				// update index time
				for idxtype, idxinfo := range n.Indexes {
					if ftext[0] == 'y' {
						idxpath := n.IndexPath() + "/" + idxtype + ".idx"
						if idxStat, ierr := os.Stat(idxpath); ierr == nil {
							idxinfo.CreatedOn = idxStat.ModTime()
						} else {
							idxinfo.CreatedOn = time.Now()
						}
					} else {
						idxinfo.CreatedOn = time.Now()
					}
					n.Indexes[idxtype] = idxinfo
				}
				n.Save()
				updated += 1
			}
			fmt.Println(fmt.Sprintf("Node schema version 4 update complete, updated %d nodes.", updated))
		} else {
			fmt.Println("Exiting.")
			os.Exit(0)
		}
	}

	return
}
