package mothership

import (
	"time"

	"github.com/sirupsen/logrus"
)

type Starbase struct {
	ID string
	Ip string
}

func waitUntilStarbaseConnects(star *Starbase) error {
	time.Sleep(200 * time.Millisecond)
	logrus.Debugf("Starbase connected: %+v", star)
	return nil
}

func deployService(service string, star *Starbase) error {
	time.Sleep(100 * time.Millisecond)
	logrus.Debugf("Deployed service %s to starbase: %+v", service, star)
	return nil
}

func CreateStarbase() (*Starbase, error) {
	star := &Starbase{ID: "starbae", Ip: "10.0.0.42"}
	logrus.Debugf("Created starbase: %+v", star)
	waitUntilStarbaseConnects(star)
	return star, nil
}

func LaunchService(star *Starbase, nacl, uplink string) error {
	logrus.Debugf("Built service with nacl: %s with uplink: %s", nacl, uplink)
	serviceShasum := "<service-shasum>"
	deployService(serviceShasum, star)
	return nil
}
