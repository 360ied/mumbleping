package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	mathrand "math/rand"
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

func sendNotif(title, description string) {
	log.Printf("[STATUS]\n%s\n%s", title, description)

	ntf := notify.NewNotification(title, description)
	if _, err := ntf.Show(); err != nil {
		log.Printf("[ERROR] Could not show notification: %s", err.Error())
	}
}

func main() {
	var timeout int64
	var outJson bool
	var watch bool
	var faultChance float64
	flag.Int64Var(&timeout, "timeout", 1000, "Ping timeout in milliseconds")
	flag.BoolVar(&outJson, "json", false, "Output in JSON")
	flag.BoolVar(&watch, "watch", false, "Watch the Mumble server and show notifications")
	flag.Float64Var(&faultChance, "fault-chance", 0,
		"Chance for randomly injecting a fault into the ping mechanism (valid values range from 0.0 to 1.0)")

	flag.Parse()

	if faultChance < 0 || faultChance > 1 {
		log.Fatalf("Args Error: -fault-chance's value isn't between 0.0 and 1.0")
	}

	ip := flag.Arg(0)
	if ip == "" {
		log.Fatalf("Args Error: IP Address not provided.")
	}

	// previous result
	pRes := mping.PingResult{}
	pUp := false
	retry := false

	up := false

start:

	res, err := mping.Ping(log.Default(), ip, timeout)
	if faultChance != 0 && faultChance < mathrand.Float64() {
		err = errors.New("fault injection")
	}
	if err != nil {
		up = false
		res = mping.PingResult{}
	} else {
		up = true
	}

	if watch {
		goto watchSkip
	}

	if !up {
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

	os.Exit(0)

watchSkip:

	if retry {
		if up {
			log.Printf("[INFO] Retry succeeded. The server is not actually down.")
			retry = false
			goto retrySkip
		} else {
			log.Printf("[INFO] Retry failed. The server is most likely actually down.")
			retry = false
		}
	}

	// up status has changed
	if up != pUp {
		if !up && pUp {
			// the ping packet might've just been lost in transit
			// try again to be more sure
			log.Printf("[INFO] The last ping was not responded to but the last last ping was. The ping packet might've been lost in transit. Retrying to make sure the server is actually down...")
			retry = true
			goto retrySkip
		}

		statusS := ""
		if up {
			statusS = "up"
		} else {
			statusS = "down"
		}

		sendNotif(
			fmt.Sprintf("%s is now %s", ip, statusS),
			fmt.Sprintf("%d/%d users connected", res.UserC, res.MaxUserC))
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

		sendNotif(
			fmt.Sprintf("%d user%s %s %s", diffAbs, pluralS, joinS, ip),
			fmt.Sprintf("%d/%d users connected", res.UserC, res.MaxUserC))
	}

retrySkip:
	time.Sleep(2 * time.Second)
	pRes = res
	pUp = up
	goto start
}
