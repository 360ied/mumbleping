package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"mumbleping/mping"
	"os"
	"time"

	"github.com/TheCreeper/go-notify"
)

// taken from
// http://cavaliercoder.com/blog/optimized-abs-for-int64-in-go.html
func Int64Abs(n int64) int64 {
	y := n >> 63
	return (n ^ y) - y
}

func main() {
	var timeout int64
	var outJson bool
	var watch bool
	flag.Int64Var(&timeout, "timeout", 1000, "Ping timeout in milliseconds")
	flag.BoolVar(&outJson, "json", false, "Output in JSON")
	flag.BoolVar(&watch, "watch", false, "Watch the Mumble server and show notifications")

	flag.Parse()

	ip := flag.Arg(0)
	if ip == "" {
		log.Fatalf("Args Error: IP Address not provided.")
	}

	// previous result
	pRes := mping.PingResult{}
	pUp := false

	up := false

start:

	res, err := mping.Ping(log.Default(), ip, timeout)
	if err != nil {
		up = false
		res = mping.PingResult{}
	} else {
		up = true
	}

	if !up && !watch {
		log.Printf(
			"[ERROR] Ping returned an error, this probably means the Mumble server is not up: %s",
			err.Error())
		os.Exit(1)
	}
	if outJson {
		outJ, err := json.MarshalIndent(res, "", "\t")
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s\n", outJ)
	} else {
		fmt.Printf(
			"-----Ping-----\nVersion: %s\nUsers: %d/%d\nAllowed Bandwidth: %d\nLatency: %s\n--------------\n",
			res.Version, res.UserC, res.MaxUserC, res.AllowedBandwidth, res.Latency.String())
	}

	if !watch {
		return
	}

	// up status has changed
	if up != pUp {
		statusS := ""
		if up {
			statusS = "up"
		} else {
			statusS = "down"
		}

		ntf := notify.NewNotification(
			fmt.Sprintf("mping: %s is now %s.", ip, statusS),
			fmt.Sprintf("%d/%d users connected.", res.UserC, res.MaxUserC))

		if _, err := ntf.Show(); err != nil {
			panic(err)
		}
	} else if res.UserC != pRes.UserC {
		joinS := ""
		if res.UserC > pRes.UserC {
			joinS = "joined"
		} else {
			joinS = "left"
		}

		userCDiff := int64(res.UserC) - int64(pRes.UserC)
		diffAbs := Int64Abs(userCDiff)

		pluralS := ""
		if diffAbs > 1 {
			pluralS = "s"
		}

		ntf := notify.NewNotification(
			fmt.Sprintf("mping: %d user%s %s %s.", diffAbs, pluralS, joinS, ip),
			fmt.Sprintf("%d/%d users connected.", res.UserC, res.MaxUserC))

		if _, err := ntf.Show(); err != nil {
			panic(err)
		}
	}

	time.Sleep(2 * time.Second)
	pRes = res
	pUp = up
	goto start
}
