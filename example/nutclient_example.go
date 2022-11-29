package main

import (
	"crypto/tls"
	"fmt"
	"os"

	"github.com/clamy54/nutclient"
)

func main() {

	// Connect to the nut server
	c, err := nutclient.Dial("ups.mydomain.com:3493")
	if err != nil {
		println("Error: ", err.Error())
		os.Exit(1)
	}

	defer c.Close()

	// This server support TLS, so we are starting a TLS session
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         "localhost",
	}

	err = c.StartTLS(tlsconfig)
	if err != nil {
		println("TLS Error: ", err.Error())
		os.Exit(1)
	}

	// Authenticate against nut server
	err = c.Auth("mylogin", "mypassword")
	if err != nil {
		println("Auth Error: ", err.Error())
		os.Exit(1)
	}

	// Select default ups
	err = c.Login("my-ups-name")
	if err != nil {
		println("Login Error: ", err.Error())
		os.Exit(1)
	}

	// Get Ups model name
	model, err := c.GetUpsModel()
	if err != nil {
		println("Cannot get UPS Model: ", err.Error())
		os.Exit(1)
	}
	fmt.Printf("UPS Name : %s\n", model)

	// Check if ups is online
	online, _ := c.IsOnline()

	if online {
		fmt.Printf("ups is online \n")
	}

	// Check if ups is on battery.
	onbattery, _ := c.IsOnBattery()

	if onbattery {
		fmt.Printf("ups is on batteries \n")
		charge, err := c.BatteryCharge()
		if err == nil {
			// Display ups charge
			fmt.Printf("Charge : %v %% \n", charge)
		} else {
			fmt.Println("Cannot get ups charge")
		}
	}

}
