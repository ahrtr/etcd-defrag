// etcd-defrag - A smarter etcd defragmentation tool
package main

import (
	"log"
	"os"

	"github.com/ahrtr/etcd-defrag/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
}
