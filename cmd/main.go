package main

import (
	"context"
	"fmt"
	"math"
	"time"

	todoist "github.com/niskhakov/gotodoist"
)

func main() {
	client, err := todoist.NewClient("<clientID here>", "<clientSecret goes here>")
	accessToken := "<accessToken for specific client here>"
	if err != nil {
		fmt.Printf("Got error while creating client: %v\n", err)
		return
	}

	projects, err := client.GetProjects(context.Background(), accessToken)
	if err != nil {
		fmt.Printf("Couldn't get projects, error = %v\n", err)
	}

	for i, p := range projects {
		fmt.Printf("%d: %s (%d), inbox = %v\n", i, p.Name, p.ID, p.InboxProject)
	}
	fmt.Println()

	// Find inbox project
	inboxID := -1
	for _, p := range projects {
		if p.InboxProject {
			inboxID = p.ID
		}
	}

	if inboxID == -1 {
		fmt.Printf("Inbox project not found")
		return
	}

	tasks, err := client.GetTasksByProject(context.Background(), accessToken, inboxID)
	if err != nil {
		fmt.Printf("Couldn't get tasks, error = %v\n", err)
		return
	}

	for i, t := range tasks {
		if t.Due.Datetime == "" {
			continue
		}
		ddt, err := time.Parse(time.RFC3339, t.Due.Datetime)
		if err != nil {
			fmt.Printf("\tCan't cast to time - %s\n", t.Content)
			continue
		}
		fmt.Printf("%d: %s (%d) - %v, until %v mins\n", i, t.Content, t.ProjectID, ddt, math.Round(time.Until(ddt).Minutes()))
	}
}
