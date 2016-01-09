// Package main provides whole program entry point
package go2apns

// Notification define the data structure for one
// notification
type Notification struct {
	Token      string // device token for this apn
	Id         string // apn-id
	Expiration int    // apn-expiration
	Priority   int    // apn-priority
	BundleID   string // bundle id like com.xxx.app.haha
	Alert      string // alert text conetent

	Result chan string // push result channel
}
