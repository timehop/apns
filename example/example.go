package main

import (
	"fmt"
	"log"

	"github.com/timehop/apns"
)

func main() {
	c, err := apns.NewClientWithFiles(apns.ProductionGateway, "apns.crt", "apns.key")
	if err != nil {
		log.Fatal("Could not create client: ", err.Error())
	}

	i := 1
	for {
		fmt.Print("Enter '<token> <badge> <msg>': ")

		var tok, body string
		var badge uint

		_, err := fmt.Scanf("%s %d %s", &tok, &badge, &body)
		if err != nil {
			fmt.Printf("Something went wrong: %v\n", err.Error())
			continue
		}

		p := apns.NewPayload()
		p.APS.Alert.Body = body
		p.APS.Badge.Set(badge)

		p.SetCustomValue("link", "yourapp://precache/20140718")

		m := apns.NewNotification()
		m.Payload = p
		m.DeviceToken = tok
		m.Priority = apns.PriorityImmediate
		m.Identifier = uint32(i)

		c.Send(m)

		i++
	}
}
