# ECU-R 

ECU-R is a library to extract PV array and inverter information from the AP systems ECU-R energy monitor. You can use the library to connect and use the information directly in your own programs, or use the provided binary to print the information in a table of JSON format.

## Supported information

* ECU level information
    * Number of inverters registered / online
    * Current energy
    * Lifetime production
    * Today's total production
* Inverter level information
    * Inverter status (online/offline)
    * Inverter signal strength
    * Grid frequency
    * Grid voltage
    * per MMPT power information

## Ongoing work

* Historic production information by week, month and year

## Usage

### Via the provided cli tool

`go run github.com/hectormalot/ecur/cmd get --host $WIFI_IP_OF_ECUR --json`

### Using the library

````golang
package main

func main() {
    // Error handling omitted for clarity
    c, _ := ecur.NewClient(host, port)
    _ = c.Connect()
    defer c.Close()

    // ECU Level information
    EcuInfo, _ := c.GetECUInfo()
    c.ecuID = ecuInfo.EcuID

    // Inverter statistics
    InverterInfo, _ := c.GetInverterInfo()

    // Inverter signal strength
    InverterSignal, _ := c.GetInverterSignal()

    fmt.Println(EcuInfo, InverterInfo, InverterSignal)
}
````

## Contribution

I only have a QS1 system at home. For YC600/YC1000 inverter models, I've used the reverse engineering work documented at the [home assistant forums](https://community.home-assistant.io/t/apsystems-aps-ecu-r-local-inverters-data-pull/260835/234) and [tweakers.net](https://gathering.tweakers.net/forum/list_messages/2032302?data%5Bfilter_keywords%5D=aps1100280030). I'm happy to take pull requests for additional functionality and/or bug fixes.

## License

All code made available under the MIT license