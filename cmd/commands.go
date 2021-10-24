package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hectormalot/ecur"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "aps",
	Short: "APS reads data from the AP Systems ECU-R energy monitor for solar panels",
	Long: `APS connects (wifi only) with the AP Systems ECU-R and extracts
useful information such as inverter status, production statistics,
and zigbee signal strength towards connected inverters`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get data from APS ECU-R",
	Run:   GetData,
}

func GetData(cmd *cobra.Command, args []string) {
	c, err := ecur.NewClient(host, port)
	if err != nil {
		log.Fatal("Error:", err)
	}

	EcuData, err := c.GetData()
	if err != nil {
		log.Fatal("Error: ", err)
	}

	if outputJson {
		PrintJSON(EcuData)
		return
	}

	PrintTable(EcuData)
}

func PrintJSON(data ecur.ECUResponse) {
	output, err := json.Marshal(data)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	fmt.Println(string(output))
}

func PrintTable(data ecur.ECUResponse) {
	// ECU information
	pterm.DefaultSection.Println("ECU information:")
	pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
		{"Parameter", "Value", "Unit"},
		{"ECU ID", data.ECUInfo.EcuID, ""},
		{"Software version", data.ECUInfo.Version, ""},
		{"Inverters", fmt.Sprintf("%d/%d", data.ECUInfo.InvertersOnline, data.ECUInfo.InvertersRegistered), "Online/Registered"},
		{"Lifetime Production", fmt.Sprintf("%.1f", float64(data.ECUInfo.LifetimeEnergy)/1000), "kWh"},
		{"Today's Production", fmt.Sprintf("%.3f", float64(data.ECUInfo.TodayEnergy)/1000), "kWh"},
		{"Current Power", fmt.Sprintf("%d", data.ECUInfo.LastPower), "W"},
		{"Ethernet MAC", data.ECUInfo.EthernetMac, ""},
		{"WiFi MAC", data.ECUInfo.WirelessMac, ""},
		{"Last update", data.ArrayInfo.Timestamp.Format("2006-01-02 15:04:05"), ""},
	}).Render()

	// Array information
	for n, i := range data.ArrayInfo.Inverters {
		pterm.DefaultSection.WithLevel(2).Printf("Inverter %s", i.ID)

		pterm.DefaultTable.WithHasHeader().WithData(pterm.TableData{
			{"Parameter", "Value", "Unit"},
			{"Model", i.Model, ""},
			{"Signal", fmt.Sprintf("%.1f", float64(data.InverterSignalInfo.Inverters[n].Signal)/2.56), "%"},
			{"Frequency", fmt.Sprintf("%.2f", i.Frequency), "Hz"},
			{"Voltage", fmt.Sprint(i.VoltageA), "V"},
			{"Temperature", fmt.Sprint(i.Temperature), "Celsius"},
			{"PowerA", fmt.Sprint(i.PowerA), "W"},
			{"PowerB", fmt.Sprint(i.PowerB), "W"},
			{"PowerC", fmt.Sprint(i.PowerC), "W"},
			{"PowerD", fmt.Sprint(i.PowerD), "W"},
		}).Render()
	}
}
