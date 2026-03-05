package ukaz

import "time"

const governorNodeID = "GOVERNOR"

// statusNodes are the nodes queried for system status (GET:UPTIME).
var statusNodes = []string{"LUCH", "VERTEX", "ACHTUNG", "GOVERNOR"}

const statusCollectTimeout = 2 * time.Second
