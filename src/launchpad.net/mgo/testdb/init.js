
// Setup auth on db2.
for (var i = 0; i != 60; i++) {
	try {
        db2 = new Mongo("127.0.0.1:40002").getDB("admin")
		break
	} catch(err) {
		print("Can't connect yet...")
	}
	sleep(1000)
}
db2.addUser("root", "rapadura")
db2.auth("root", "rapadura")
db2.addUser("reader", "rapadura", true)

//var settings = {heartbeatSleep: 0.05, heartbeatTimeout: 0.5}
var settings = {}

// We know the master of the first set (pri=1), but not of the second.
var rs1cfg = {_id: "rs1",
              members: [{_id: 1, host: "127.0.0.1:40011", priority: 1},
                        {_id: 2, host: "127.0.0.1:40012", priority: 0},
                        {_id: 3, host: "127.0.0.1:40013", priority: 0}],
              settings: settings}
var rs2cfg = {_id: "rs2",
              members: [{_id: 1, host: "127.0.0.1:40021", priority: 1},
                        {_id: 2, host: "127.0.0.1:40022", priority: 1},
                        {_id: 3, host: "127.0.0.1:40023", priority: 1}],
              settings: settings}

for (var i = 0; i != 60; i++) {
	try {
		db1 = new Mongo("127.0.0.1:40001").getDB("admin")
		db2 = new Mongo("127.0.0.1:40002").getDB("admin")
		rs1a = new Mongo("127.0.0.1:40011").getDB("admin")
		rs2a = new Mongo("127.0.0.1:40021").getDB("admin")
		break
	} catch(err) {
		print("Can't connect yet...")
	}
	sleep(1000)
}

rs2a.runCommand({replSetInitiate: rs2cfg})
rs1a.runCommand({replSetInitiate: rs1cfg})

function configShards() {
    cfg1 = new Mongo("127.0.0.1:40201").getDB("admin")
    cfg1.runCommand({addshard: "127.0.0.1:40001"})
    cfg1.runCommand({addshard: "rs1/127.0.0.1:40011"})

    cfg2 = new Mongo("127.0.0.1:40202").getDB("admin")
    cfg2.runCommand({addshard: "rs2/127.0.0.1:40021"})
}

function countHealthy(rs) {
    var status = rs.runCommand({replSetGetStatus: 1})
    var count = 0
    if (typeof status.members != "undefined") {
        for (var i = 0; i != status.members.length; i++) {
            var m = status.members[i]
            if (m.health == 1 && (m.state == 1 || m.state == 2)) {
                count += 1
            }
        }
    }
    return count
}

var totalRSMembers = rs1cfg.members.length + rs2cfg.members.length

for (var i = 0; i != 60; i++) {
    var count = countHealthy(rs1a) + countHealthy(rs2a)
    print("Replica sets have", count, "healthy nodes.")
    if (count == totalRSMembers) {
        sleep(2000)
        configShards()
        quit(0)
    }
    sleep(1000)
}

print("Replica sets didn't sync up properly.")
quit(12)

// vim:ts=4:sw=4:et
