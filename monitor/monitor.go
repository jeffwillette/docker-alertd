package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/smtp"
	"time"

	"github.com/docker/docker/client"
)

func sendEmail(email *EmailSettings, subject, message string) {
	// Set up authentication information.
	auth := smtp.PlainAuth("", email.From, email.Password, email.SMTP)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{email.To}
	msg := []byte(
		"To: " + email.To + "\r\n" +
			"Subject: " + subject + "\r\n" + "\r\n" +
			message + "\r\n")

	port := fmt.Sprintf("%d", email.Port)
	err := smtp.SendMail(
		email.SMTP+":"+port,
		auth,
		email.From,
		to,
		msg)

	if err != nil {
		log.Fatal(err)
	}
}


// TODO: write doc here
func checkContainer(container *Container, cli *client.Client) ([]string) {
	// string of alerts to be returned
	var alerts []string
	rawStats, err := cli.ContainerStats(
		context.Background(), container.Name, false)
	// TODO: make this an alert that the container does no exist and skip it
	// servie will need to be restarted when the container is present
	if err != nil {
		log.Fatal(err)
	}

	// Gets the stats from a Reader interface and appends them to a
	// byte slice
	var stats []byte
	buf := make([]byte, 1024)

	for {
		n, err := rawStats.Body.Read(buf)
		stats = append(stats, buf[:n]...)
		if err != nil {
			// EOF will register as an error so the inner loop only
			// executes if there is a non EOF Error
			if err != io.EOF {
				log.Fatalf("read error while reading stats: ", err)
			}
			break
		}
	}
	rawStats.Body.Close()

	// Unmarshaling the JSON blob into an AlertdStats object
	var alertdStats AlertdStats
	err = json.Unmarshal(stats, &alertdStats)
	if err != nil {
		log.Fatalf("Error unmarshaling JSON: %s", err)
	}

	// Checking different stats and adding them to alert slices to be
	// emailed after all containers have been checked
	if alert, usage := checkCpuUsage(container, &alertdStats); alert {
		alertMessage := fmt.Sprintf("CPU ALERT: %s's CPU usage "+
			"exceeded %d, it "+"is currently using %f\n",
			container.Name, container.MaxCpu, usage)

		alerts = append(alerts, alertMessage)
		log.Printf(alertMessage)
	}

	if alert, pids := checkMinPids(container, &alertdStats); alert {
		alertMessage := fmt.Sprintf("PID ALERT: %s's running processes "+
			"went below %d, there are currently %d PID's\n",
			container.Name, container.MinProcs, pids)

		alerts = append(alerts, alertMessage)
		log.Printf(alertMessage)
	}
	return alerts
}


// The main loop should continously run forever
func Start(conf *Conf) {
	log.Printf("started docker-alertd process\n------------------------------")

	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}

	for {
		var emailMessage string
		for _, container := range conf.Containers {
			a := checkContainer(&container, cli)
			// Concatenating all alert messages into one string
			for _, alert := range a {
				emailMessage += fmt.Sprintf("%s", alert)
			}
		}

		if len(emailMessage) > 0 {
			sendEmail(&conf.EmailSettings, "docker ALERT", emailMessage)
			log.Println("alert email sent")
		}

		time.Sleep(1 * time.Second)
	}
}
