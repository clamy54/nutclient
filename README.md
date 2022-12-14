nutclient for Golang
====================

## About

Network UPS Tools (NUT) client module for Go.

## Examples

View inside example directory.

## Functions (depends on nut server configuration and ups capabilities)

* Dial(address)
* StartTLS(tlsconfig) 
* Auth("login","password")
* Login("upsname")
* Close()
* GetUpsModel()
* Logout()
* IsOnline()
* IsOnBattery()
* IsLowBattery()
* BatteryCharge()
* BatteryChargeLow()
* BatteryChargeWarning()
* BatteryChargeRestart()
* BatteryRuntime()
* BatteryRuntimeLow()
* BatteryRuntimeRestart()
* GetServerInfo()
* GetServerVersion()
* UpsLoad()
* UpsTemperature()
* UpsApparentPower()
* UpsActivePower()
* InputVoltage()
* InputCurrent()
* OutputVoltage()
* OutputCurrent()
* OutputFrequency()
* InputFrequency()
* GetUpsModel()
* GetUpsSerial()
* GetServerUpsList()
* GetUpsVars()
* GetData(varname)



## License

Released under the [MIT license][1]. See `LICENSE.md` file for details.

[1]: http://www.opensource.org/licenses/mit-license.php
