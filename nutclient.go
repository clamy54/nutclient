// Package nutclient provides a simple client for connecting to UPS NUT daemon
package nutclient

import (
	"crypto/tls"
	"errors"
	"net"
	"net/textproto"
	"strconv"
	"strings"
)

// A Client represents a client connection to a nut server.
type Client struct {
	Text       *textproto.Conn
	conn       net.Conn
	tls        bool
	serverName string
	upsName    string
}

// The addr must include a port, as in "nutsrv.example.com:3493".
func Dial(address string) (*Client, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, err
	}
	host, _, _ := net.SplitHostPort(address)
	return NewClient(conn, host)
}

// NewClient returns a new Client instance
func NewClient(conn net.Conn, host string) (*Client, error) {
	text := textproto.NewConn(conn)
	c := &Client{Text: text, conn: conn, serverName: host, tls: false, upsName: ""}
	_, c.tls = conn.(*tls.Conn)
	return c, nil
}

// Close closes the connection.
func (c *Client) Close() error {
	return c.Text.Close()
}

// Sends a command and returns the response
func (c *Client) cmd(format string) (string, error) {

	err := c.Text.PrintfLine("%s", format)
	if err != nil {
		return "", err
	}
	response, err := c.Text.ReadLine()

	if err != nil {
		return "", err
	}
	retcode, retarg, _ := strings.Cut(response, " ")

	if strings.EqualFold(retcode, "OK") {
		return retarg, nil
	} else {
		return "", errors.New(response)
	}
}

// Get a specific data from current ups
func (c *Client) GetData(format string) (string, error) {

	if len([]rune(c.upsName)) == 0 {
		return "", errors.New("No UPS defined, use LOGIN first")
	}

	if len(format) == 0 {
		return "", errors.New("Variable cannot be empty")
	}

	command := "GET VAR " + c.upsName + " " + format

	err := c.Text.PrintfLine("%s", command)

	if err != nil {
		return "", err
	}
	response, err := c.Text.ReadLine()

	if err != nil {
		return "", err
	}
	retcode, retarg, _ := strings.Cut(response, " ")

	if strings.EqualFold(retcode, "VAR") {
		_, retarg, _ = strings.Cut(response, "\"")
		retarg = strings.ReplaceAll(retarg, "\"", "")
		return retarg, nil
	} else {
		return "", errors.New(response)
	}
}

// Get a multiline data response
func (c *Client) getmultilinesdata(command string) ([]string, error) {

	var retslice []string
	if len(command) == 0 {
		return nil, errors.New("Variable cannot be empty")
	}

	err := c.Text.PrintfLine("%s", command)

	if err != nil {
		return nil, err
	}
	response, err := c.Text.ReadLine()

	if err != nil {
		return nil, err
	}
	retcode, _, _ := strings.Cut(response, " ")

	if strings.EqualFold(retcode, "BEGIN") {
		exitloop := false
		for !exitloop {
			response, err := c.Text.ReadLine()

			if err != nil {
				return nil, err
			}
			retcode, _, _ := strings.Cut(response, " ")
			if strings.EqualFold(retcode, "END") {
				exitloop = true
			} else {
				retslice = append(retslice, response)
			}
		}
	} else {
		return nil, errors.New(response)
	}
	return retslice, nil
}

// StartTLS sends the STARTTLS command and encrypts all further communication.
func (c *Client) StartTLS(configtls *tls.Config) error {

	_, err := c.cmd("STARTTLS")
	if err != nil {
		return err
	}
	c.conn = tls.Client(c.conn, configtls)
	c.Text = textproto.NewConn(c.conn)
	c.tls = true
	return err
}

// Perform Auth on nut server
func (c *Client) Auth(login string, password string) error {

	_, err := c.cmd("USERNAME " + login)
	if err != nil {
		return errors.New("ERROR : Bad Login " + err.Error())
	}

	_, err = c.cmd("PASSWORD " + password)
	if err != nil {
		return errors.New("ERROR : Bad Password " + err.Error())
	}

	return nil
}

// Perform Login command to select current ups
func (c *Client) Login(upsName string) error {

	_, err := c.cmd("LOGIN " + upsName)
	if err != nil {
		return err
	}

	c.upsName = upsName
	return nil
}

// Perform Logout command.
func (c *Client) Logout() error {

	_, err := c.cmd("LOGOUT")
	if err != nil {
		return err
	}
	return nil
}

// Return true if current ups is online
func (c *Client) IsOnline() (bool, error) {
	online := false
	if len([]rune(c.upsName)) == 0 {
		return false, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.status")

	if err != nil {
		return false, errors.New("Error getting current ups status")
	}
	firstarg, _, _ := strings.Cut(result, " ")

	if len([]rune(firstarg)) >= 2 {

		if (strings.ToUpper(firstarg) == "OL") || (strings.ToUpper(firstarg) == "BYPASS") {
			online = true
		}
	} else {
		return false, errors.New("Cannot identify ups response")
	}
	return online, nil
}

// Return true if current ups is on battery
func (c *Client) IsOnBattery() (bool, error) {
	onbattery := false
	if len([]rune(c.upsName)) == 0 {
		return false, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.status")
	if err != nil {
		return false, errors.New("Error getting current ups status")
	}

	if len([]rune(result)) >= 2 {
		if (strings.ToUpper(result[0:2]) == "OB") || (strings.ToUpper(result[0:2]) == "LB") {
			onbattery = true
		}
	} else {
		return false, errors.New("Cannot identify ups response")
	}
	return onbattery, nil
}

// Return true if current ups status is low battery
func (c *Client) IsLowBattery() (bool, error) {
	lowbattery := false
	if len([]rune(c.upsName)) == 0 {
		return false, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.status")
	if err != nil {
		return false, errors.New("Error getting current ups status")
	}

	if len([]rune(result)) >= 2 {
		if strings.ToUpper(result[0:2]) == "LB" {
			lowbattery = true
		}
	} else {
		return false, errors.New("Cannot identify ups response")
	}
	return lowbattery, nil
}

// Return Battery Charge
func (c *Client) BatteryCharge() (int, error) {
	charge := -1
	if len([]rune(c.upsName)) == 0 {
		return charge, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.charge")
	if err != nil {
		return charge, errors.New("Error getting current battery charge")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return charge, errors.New("Cannot convert battery charge to numerical value")
	} else {
		charge = value
	}
	return charge, nil
}

// Return Battery Charge Low value
func (c *Client) BatteryChargeLow() (int, error) {
	charge := -1
	if len([]rune(c.upsName)) == 0 {
		return charge, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.charge.low")
	if err != nil {
		return charge, errors.New("Error getting current battery charge low")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return charge, errors.New("Cannot convert battery charge low to numerical value")
	} else {
		charge = value
	}
	return charge, nil
}

// Return Battery Charge Warning value
func (c *Client) BatteryChargeWarning() (int, error) {
	charge := -1
	if len([]rune(c.upsName)) == 0 {
		return charge, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.charge.warning")
	if err != nil {
		return charge, errors.New("Error getting current battery charge warning")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return charge, errors.New("Cannot convert battery charge warning to numerical value")
	} else {
		charge = value
	}
	return charge, nil
}

// Return Battery Charge Restart value
func (c *Client) BatteryChargeRestart() (int, error) {
	charge := -1
	if len([]rune(c.upsName)) == 0 {
		return charge, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.charge.restart")
	if err != nil {
		return charge, errors.New("Error getting current battery charge restart")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return charge, errors.New("Cannot convert battery charge restart to numerical value")
	} else {
		charge = value
	}
	return charge, nil
}

// Return Battery runtime (seconds)
func (c *Client) BatteryRuntime() (int, error) {
	runtime := -1
	if len([]rune(c.upsName)) == 0 {
		return runtime, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.runtime")
	if err != nil {
		return runtime, errors.New("Error getting current battery runtime")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return runtime, errors.New("Cannot convert battery runtime to numerical value")
	} else {
		runtime = value
	}
	return runtime, nil
}

// Return Battery runtime  low (seconds)
func (c *Client) BatteryRuntimeLow() (int, error) {
	runtime := -1
	if len([]rune(c.upsName)) == 0 {
		return runtime, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.runtime.low")
	if err != nil {
		return runtime, errors.New("Error getting current battery runtime low")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return runtime, errors.New("Cannot convert battery runtime low to numerical value")
	} else {
		runtime = value
	}
	return runtime, nil
}

// Return Battery runtime restart (seconds)
func (c *Client) BatteryRuntimeRestart() (int, error) {
	runtime := -1
	if len([]rune(c.upsName)) == 0 {
		return runtime, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("battery.runtime.restart")
	if err != nil {
		return runtime, errors.New("Error getting current battery runtime restart")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return runtime, errors.New("Cannot convert battery runtime restart to numerical value")
	} else {
		runtime = value
	}
	return runtime, nil
}

// Return Server Info
func (c *Client) GetServerInfo() (string, error) {
	info := ""

	result, err := c.GetData("server.info")
	if err != nil {
		return info, errors.New("Error getting server.info")
	}

	info = result
	return info, nil
}

// Return Server Version
func (c *Client) GetServerVersion() (string, error) {
	info := ""

	result, err := c.GetData("server.version")
	if err != nil {
		return info, errors.New("Error getting server.version")
	}

	info = result
	return info, nil
}

// Return configured UPS list
func (c *Client) GetServerUpsList() ([]string, error) {
	var retslice []string

	result, err := c.getmultilinesdata("LIST UPS")

	if (err != nil) || (len(result) == 0) {
		return nil, errors.New("Error getting ups list")
	}

	for _, value := range result {
		retcode, _, _ := strings.Cut(value, " ")

		if strings.EqualFold(retcode, "UPS") {
			argsstr := strings.Fields(value)
			if len(argsstr) > 0 {
				retslice = append(retslice, argsstr[1])
			}
		}
	}
	return retslice, nil
}

// Return ups vars avaible
func (c *Client) GetUpsVars() ([]string, error) {
	var retslice []string

	if len([]rune(c.upsName)) == 0 {
		return nil, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.getmultilinesdata("LIST VAR " + c.upsName)

	if (err != nil) || (len(result) == 0) {
		return nil, errors.New("Error getting ups list")
	}

	for _, value := range result {
		retcode, _, _ := strings.Cut(value, " ")

		if strings.EqualFold(retcode, "VAR") {
			argsstr := strings.Fields(value)
			if len(argsstr) > 3 {
				retslice = append(retslice, argsstr[2])
			}
		}
	}
	return retslice, nil
}

// Return ups load (percent)
func (c *Client) UpsLoad() (int, error) {
	upsload := -1
	if len([]rune(c.upsName)) == 0 {
		return upsload, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.load")
	if err != nil {
		return upsload, errors.New("Error getting current ups load")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return upsload, errors.New("Cannot convert ups load to numerical value")
	} else {
		upsload = value
	}
	return upsload, nil
}

// Return ups load (degrees C)
func (c *Client) UpsTemperature() (int, error) {
	upstemperature := -1
	if len([]rune(c.upsName)) == 0 {
		return upstemperature, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.temperature")
	if err != nil {
		return upstemperature, errors.New("Error getting current ups temperature")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return upstemperature, errors.New("Cannot convert ups temperature to numerical value")
	} else {
		upstemperature = value
	}
	return upstemperature, nil
}

// Return current apparent ups power (VA)
func (c *Client) UpsApparentPower() (int, error) {
	upspower := -1
	if len([]rune(c.upsName)) == 0 {
		return upspower, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.power")
	if err != nil {
		return upspower, errors.New("Error getting current ups apparent power")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return upspower, errors.New("Cannot convert ups apparent power to numerical value")
	} else {
		upspower = value
	}
	return upspower, nil
}

// Return current active ups power (W)
func (c *Client) UpsActivePower() (int, error) {
	upspower := -1
	if len([]rune(c.upsName)) == 0 {
		return upspower, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("ups.realpower")
	if err != nil {
		return upspower, errors.New("Error getting current ups active power")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return upspower, errors.New("Cannot convert ups active power to numerical value")
	} else {
		upspower = value
	}
	return upspower, nil
}

// Return Input Voltage (V)
func (c *Client) InputVoltage() (int, error) {
	voltage := -1
	if len([]rune(c.upsName)) == 0 {
		return voltage, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("input.voltage")
	if err != nil {
		return voltage, errors.New("Error getting current input voltage")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return voltage, errors.New("Cannot convert input voltage to numerical value")
	} else {
		voltage = value
	}
	return voltage, nil
}

// Return Input Current (A)
func (c *Client) InputCurrent() (int, error) {
	courant := -1
	if len([]rune(c.upsName)) == 0 {
		return courant, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("input.current")
	if err != nil {
		return courant, errors.New("Error getting current input current")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return courant, errors.New("Cannot convert input current to numerical value")
	} else {
		courant = value
	}
	return courant, nil
}

// Return Output Voltage (V)
func (c *Client) OutputVoltage() (int, error) {
	voltage := -1
	if len([]rune(c.upsName)) == 0 {
		return voltage, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("output.voltage")
	if err != nil {
		return voltage, errors.New("Error getting current output voltage")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return voltage, errors.New("Cannot convert output voltage to numerical value")
	} else {
		voltage = value
	}
	return voltage, nil
}

// Return Output Current (A)
func (c *Client) OutputCurrent() (int, error) {
	courant := -1
	if len([]rune(c.upsName)) == 0 {
		return courant, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("output.current")
	if err != nil {
		return courant, errors.New("Error getting current output current")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return courant, errors.New("Cannot convert output current to numerical value")
	} else {
		courant = value
	}
	return courant, nil
}

// Return Output Frequency (Hz)
func (c *Client) OutputFrequency() (int, error) {
	frequency := -1
	if len([]rune(c.upsName)) == 0 {
		return frequency, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("output.frequency")
	if err != nil {
		return frequency, errors.New("Error getting current output frequency")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return frequency, errors.New("Cannot convert output frequency to numerical value")
	} else {
		frequency = value
	}
	return frequency, nil
}

// Return Input Frequency (Hz)
func (c *Client) InputFrequency() (int, error) {
	frequency := -1
	if len([]rune(c.upsName)) == 0 {
		return frequency, errors.New("No UPS defined, use LOGIN first")
	}

	result, err := c.GetData("input.frequency")
	if err != nil {
		return frequency, errors.New("Error getting current input frequency")
	}

	value, err := strconv.Atoi(result)
	if err != nil {
		return frequency, errors.New("Cannot convert input frequency to numerical value")
	} else {
		frequency = value
	}
	return frequency, nil
}

// Return Ups Model
func (c *Client) GetUpsModel() (string, error) {
	info := ""
	if len([]rune(c.upsName)) == 0 {
		return info, errors.New("No UPS defined, use LOGIN first")
	}

	info, err := c.GetData("ups.model")

	if err != nil {
		return info, errors.New("Error getting server.version")
	}

	return info, nil
}

// Return Ups Serial Number
func (c *Client) GetUpsSerial() (string, error) {
	info := ""
	if len([]rune(c.upsName)) == 0 {
		return info, errors.New("No UPS defined, use LOGIN first")
	}

	info, err := c.GetData("ups.serial")

	if err != nil {
		return info, errors.New("Error getting server.version")
	}

	return info, nil
}
