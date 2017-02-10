package monitor

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/smtp"

	"github.com/docker/docker/client"
	"github.com/antonholmquist/jason"
)

func sendEmail(email Email, subject, message string) {
	// Set up authentication information.
	auth := smtp.PlainAuth("", email.from, email.password, email.smtp)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{email.to}
	msg := []byte(
		"To: " + email.to + "\r\n" +
		"Subject: " + subject + "\r\n" + "\r\n" +
		message + "\r\n")

	port := fmt.Sprintf("%d", email.port)
	err := smtp.SendMail(
		email.smtp + ":" + port,
		auth,
		email.from,
		to,
		msg)

	if err != nil {
		log.Fatal(err)
	}
}


func Start(containers *[]Container, email *Email) {
	log.Printf("started docker-alertd process\n-----------------------------------")
	for {
		cli, err := client.NewEnvClient()
		if err != nil {
			log.Fatal(err)
		}

		for _, container := range *containers {

			// TODO: should this be put into channels?
		    rawStats, err := cli.ContainerStats(
		    	context.Background(), container.name, false)
			if err != nil { log.Fatal(err) }

			// Gets the stats from a Reader interface and appends them to a
			// byte slice
			var stats []byte
			buf := make([]byte, 1024)
			for {
				n, err := rawStats.Body.Read(buf)
				stats = append(stats, buf[:n]...)
				if err != nil {
					if err != io.EOF {
						fmt.Println("read error:", err)
					}
					break
				}
			}

			obj, _ := jason.NewObjectFromBytes(stats)

			if alertCPU := checkCpuUsage(obj, container); alertCPU {
				log.Println("CPU USAGE ALERT: ", container.name)
				sendEmail(
					*email,
					"docker ALERT",
					"CPU USAGE ALERT: " + container.name,
				)
			}
			if alertProc := checkMinPids(obj, container); alertProc {
				log.Println("PID ALERT: ", container.name)
			}
		}
	}
}