# apns

A Go package to interface with the Apple Push Notification Service

## Install

```
go get github.com/timehop/apns
```

## Usage

### Sending a push notification (basic)

```go
c, _ := apns.NewClient(apns.ProductionGateway, apnsCert, apnsKey)

p := apns.NewPayload()
p.APS.Alert.Body = "I am a push notification!"
p.APS.Badge = 5
p.APS.Sound = "turn_down_for_what.aiff"

m := apns.NewNotification()
m.Payload = p
m.DeviceToken = "A_DEVICE_TOKEN"
m.Priority = apns.PriorityImmediate

c.Send(m)
```

### Sending a push notification with error handling

```go
c, err := apns.NewClient(apns.ProductionGateway, apnsCert, apnsKey)
if err != nil {
	log.Fatal("could not create new client", err.Error()
}

go func() {
	for f := range c.FailedNotifs {
		fmt.Println("Notif", f.Notif.ID, "failed with", f.Err.Error())
	}
}()

p := apns.NewPayload()
p.APS.Alert.Body = "I am a push notification!"
p.APS.Badge = 5
p.APS.Sound = "turn_down_for_what.aiff"
p.APS.ContentAvailable = 1

p.SetCustomValue("link", "zombo://dot/com")
p.SetCustomValue("game", map[string]int{"score": 234})

m := apns.NewNotification()
m.Payload = p
m.DeviceToken = "A_DEVICE_TOKEN"
m.Priority = apns.PriorityImmediate
m.Identifier = 12312, // Integer for APNS
m.ID = "user_id:timestamp", // ID not sent to Apple – to identify error notifications

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

## Running the tests

We use [Ginkgo](https://onsi.github.io/ginkgo) for our testing framework and
[Gomega](http://onsi.github.io/gomega/) for our matchers. To run the tests:

```
go get github.com/onsi/ginkgo
go get github.com/onsi/gomega
ginkgo -randomizeAllSpecs
```

## Contributing

- Fork the repo
- Make your changes
- [Run the tests](https://github.com/timehop/apns#running-the-tests)
- Submit a pull request

If you need any ideas on what to work on, check out the
[TODO](https://github.com/timehop/apns/blob/master/TODO.md)

## License

[MIT License](https://github.com/timehop/apns/blob/master/LICENSE)
