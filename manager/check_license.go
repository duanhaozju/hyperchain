package manager

import "time"

func (self *ProtocolManager) checkExpired() {
	expiredChecker := func(currentTime, expiredTime time.Time) bool {
		return currentTime.Before(expiredTime)
	}
	ticker := time.NewTicker(1 * time.Hour)
	for {
		select {
		case <-ticker.C:
			currentTime := time.Now()
			if !expiredChecker(currentTime, self.expiredTime) {
				log.Error("License Expired")
				self.expired <- true
			}
		}
	}

}