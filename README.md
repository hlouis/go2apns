# Go2APNS

Current state is in developing, not production ready!

## Introduction
Go2APNS use http/2 protocol to send notification to Apple push notification server for iOS remote notification.

## Build & Install

## Push API
POST value to http://\<host\>/push, support keys:

- token: device token
- expiration: apns-expiration
- priority: apns-priority
- topic: apns-topic
- payload: body content of the message is the JSON string please see [The Remote Notification Payload][1]

All apns-\* value please see the [APNs Provider API][2] document.

Please check the HTTP status:
 - 200: indicate this push success, and there is no other message
- 4xx: wrong request, will follow a json object to show you the reason
- 5xx: server internal error, will follow a json object to show you the reason

json object and error status description please see apple [APNs Provider document][3]

[1]:	https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/TheNotificationPayload.html#//apple_ref/doc/uid/TP40008194-CH107-SW1
[2]:	https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/APNsProviderAPI.html#//apple_ref/doc/uid/TP40008194-CH101-SW1
[3]:	https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/APNsProviderAPI.html#//apple_ref/doc/uid/TP40008194-CH101-SW1