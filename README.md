# apns

[![GoDoc](https://godoc.org/github.com/timehop/apns?status.svg)](https://godoc.org/github.com/timehop/apns)
[![Build Status](https://travis-ci.org/timehop/apns.svg?branch=master)](https://travis-ci.org/timehop/apns)

A Go package to interface with the Apple Push Notification Service

## Features

This library implements a few features that we couldn't find in any one library elsewhere:

* **Long Lived Clients** - Apple's documentation say that you should hold [a persistent connection open](https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW6) and not create new connections for every payload
* **Use of New Protocol** - Apple came out with v2 of their API with support for variable length payloads. This library uses that protocol.
* **Robust Send Guarantees** - APNS has asynchronous feedback on whether a push sent. That means that if you send pushes after a bad send, those pushes will be lost forever. Our library records the last N pushes, detects errors, and is able to resend the pushes that could have been lost. [More reading](http://redth.codes/the-problem-with-apples-push-notification-ser/)

## API Compatibility

The apns package may undergo breaking changes. A tool like [godep](https://github.com/tools/godep) is recommended to vendor the current release.

## Install

```
go get github.com/timehop/apns
```

Checkout the `develop` branch for the current work in progress.

## Usage

### Sending a push notification (basic)

```go
c, _ := apns.NewClient(apns.ProductionGateway, apnsCert, apnsKey)

p := apns.NewPayload()
p.APS.Alert.Body = "I am a push notification!"
p.APS.Badge.Set(5)
p.APS.Sound = "turn_down_for_what.aiff"

m := apns.NewNotification()
m.Payload = p
m.DeviceToken = "A_DEVICE_TOKEN"
m.Priority = apns.PriorityImmediate

c.Send(m)
```

### Sending a push notification with error handling

```go
    c, err := apns.NewClientWithFiles(apns.ProductionGateway, "cert.pem", "key.pem")
    if err != nil {
        log.Fatal("could not create new client", err.Error())
    }

    go func() {
        for f := range c.FailedNotifs {
            fmt.Println("Notif", f.Notif.ID, "failed with", f.Err.Error())
        }
    }()

    p := apns.NewPayload()
    p.APS.Alert.Body = "I am a push notification!"
    p.APS.Badge.Set(5)
    p.APS.Sound = "turn_down_for_what.aiff"
    p.APS.ContentAvailable = 1

    p.SetCustomValue("link", "zombo://dot/com")
    p.SetCustomValue("game", map[string]int{"score": 234})

    m := apns.NewNotification()
    m.Payload = p
    m.DeviceToken = "A_DEVICE_TOKEN"
    m.Priority = apns.PriorityImmediate
    m.Identifier = 12312       // Integer for APNS
    m.ID = "user_id:timestamp" // ID not sent to Apple – to identify error notifications

    c.Send(m)
```

### Retrieving feedback

```go
f, err := apns.NewFeedback(s.Address(), DummyCert, DummyKey)
if err != nil {
	log.Fatal("Could not create feedback", err.Error())
}

for ft := range f.Receive() {
	fmt.Println("Feedback for token:", ft.DeviceToken)
}
```

Note that the channel returned from `Receive` will close after the
[feedback service](https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW3)
has no more data to send.

## Running the tests

We use [Ginkgo](https://onsi.github.io/ginkgo) for our testing framework and
[Gomega](http://onsi.github.io/gomega/) for our matchers. To run the tests:

```
go get github.com/onsi/ginkgo/ginkgo
go get github.com/onsi/gomega
ginkgo -randomizeAllSpecs
```

## Contributing

- Fork the repo ([Recommended process](https://splice.com/blog/contributing-open-source-git-repositories-go/))
- Make your changes
- [Run the tests](https://github.com/timehop/apns#running-the-tests)
- Submit a pull request

## License

[MIT License](https://github.com/timehop/apns/blob/master/LICENSE)
