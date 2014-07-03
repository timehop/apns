package main

import (
	"fmt"
	"log"

	"github.com/timehop/apns"
)

func main() {
	c, err := apns.NewClientWithFiles(apns.ProductionGateway, "apns.crt", "apns.key")
	if err != nil {
		log.Fatal("Could not create client", err.Error())
	}

	i := 0
	for {
		fmt.Print("Enter '<token> <badge> <msg>': ")

		var tok, body string
		var badge int

		_, err := fmt.Scanf("%s %d %s", &tok, &badge, &body)
		if err != nil {
			fmt.Printf("Something went wrong: %v\n", err.Error())
			continue
		}

		p := apns.Payload{}
		p.APS.Alert.Body = body
		p.APS.Badge = badge

		m := apns.Notification{
			Payload:     p,
			DeviceToken: tok,
			Priority:    apns.PriorityImmediate,
			Identifier:  uint32(i),
		}
		c.Send(m)

		i++
	}
}
