// Package main provides whole program entry point
package go2apns

// Notification define the data structure for one
// notification
type Notification struct {
	Token      string // device token for this apn
	Id         string // apn-id
	Expiration string // apn-expiration
	Priority   string // apn-priority
	Topic      string // bundle id like com.xxx.app.haha
	Payload    string // alert text conetent

	Result chan NotiResult // push result channel
}

type NotiResult struct {
	Code int    // status code for push result
	Msg  string // message if any error occured
}
