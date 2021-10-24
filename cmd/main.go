package main

import (
	"github.com/hectormalot/ecur"
)

// used for flags
var (
	outputJson bool
	host       string
	port       int
)

func main() {
	Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&host, "host", "a", "localhost", "ECU-R address")
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", ecur.DefaultPort, "Port on which to connect with ECU-R")
	rootCmd.PersistentFlags().BoolVarP(&outputJson, "json", "j", false, "Output results as JSON")
	rootCmd.AddCommand(getCmd)
}
