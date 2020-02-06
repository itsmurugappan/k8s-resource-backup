package main

import (
	"github.com/itsmurugappan/k8s-resource-backup/pkg/backup"
	"os"
)

func main() {

	action, _ := os.LookupEnv("action")

	if action == "backup" {
		backup.BackUpKnative()
	} else if action == "restore" {
		backup.RestoreKnative()
	}
}
