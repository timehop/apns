/*
A Go package to interface with the Apple Push
Notification Service

Features

This library implements a few features that we couldn't find in any one
library elsewhere:

  Long Lived Clients     - Apple's documentation say that you should hold a
                           persistent connection open and not create new
                           connections for every payload
                           See: https://developer.apple.com/library/ios/documentation/NetworkingInternet/Conceptual/RemoteNotificationsPG/Chapters/CommunicatingWIthAPS.html#//apple_ref/doc/uid/TP40008194-CH101-SW6)

  Use of New Protocol    - Apple came out with v2 of their API with support for
                           variable length payloads. This library uses that
                           protocol.

  Robust Send Guarantees - APNS has asynchronous feedback on whether a push
                           sent. That means that if you send pushes after a
                           bad send, those pushes will be lost forever. Our
                           library records the last N pushes, detects errors,
                           and is able to resend the pushes that could have
                           been lost.
                           See: http://redth.codes/the-problem-with-apples-push-notification-ser/
*/
package apns
