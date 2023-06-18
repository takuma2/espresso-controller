# Espresso Controller

PID temperature control and monitoring for a Rancilio Silvia or comparable espresso machine.
![dashboard](images/dashboard.png)
![configuration](images/configuration.png)

## Table of Contents

- [Tech](#tech-stack)
- [Requirements](#requirements)
- [Installation](#installation)
  - [Requirements](#requirements)
  - [Wiring](#wiring)
  - [Raspi Setup](#raspi-setup)
  - [Control and monitor](#control-and-monitor)
  - [Finished](#finished)
- [Contributing](#contributing)
- [Credits](#credits)

## Tech Stack

The application is a single Go binary implementing:

- gRPC API as defined in [espresso.proto](https://github.com/gregorychen3/espresso-controller/blob/master/pkg/espressopb/espresso.proto),
- dashboard web app using React and [Material-UI](https://material-ui.com/), and
- `/metrics` web endpoint for Prometheus scraping

serving on a single port (default 8080) of a Raspberry Pi.

## Installation

### Requirements

- Espresso machine, e.g., [Rancilio Silvia](https://www.ranciliogroupna.com/silvia)
- Raspberry Pi 4. Will not work well on Raspberry Pi Zero W. The Pi Zero W, having a single processor, will experience performance issues because it will not handle the necessary concurrency well (needs to measure temperature, toggle heating elements, and serve the UI static files / requests).
- [Solid state relay](https://www.amazon.com/dp/B00HV974KC/ref=cm_sw_em_r_mt_dp_U_9WTYEbEA0TNGG)
- [Type K thermocouple](https://www.amazon.com/gp/product/B01NBM7SBK)
- [MAX31855 thermocouple amplifier](https://www.adafruit.com/product/269)
- [Male blade connectors](https://en.wikipedia.org/wiki/FASTON_terminal#/media/File:Faston_Style_Terminals_Male.jpg)
- Electrical wire

### Wiring

Here is the original circuit diagram from the [manual](https://www.ranciliogroupna.com/filebin/images/Downloadables/User_Manuals/Homeline/Silvia_User_Manual_2017.PDF):
![unmodified](images/circuit_diagram_original.png)
Rewire it like this (default gpio pin numbers shown):
![modified](images/circuit_diagram_modified.png)

#### Note on Thermocouple Placement

The thermocouple should be attached securely to the outer wall of the boiler. On the Rancilio Silvia, a convenient way to accomplish this is to loosen the screw holding down the factory thermostat to the boiler. Then, the thermocouple can be slipped into the gap between the boiler and thermostat.

### Raspi Setup

1. Follow the Raspberry Pi [Getting Started Guide](https://projects.raspberrypi.org/en/projects/raspberry-pi-getting-started). Be sure to connect it to a wifi network.
1. Ensure wifi power saving mode is off.

   ```console
   pi@raspberrypi:~ $ sudo iw wlan0 set power_save off
   pi@raspberrypi:~ $ iw wlan0 get power_save
   Power save: off
   ```

1. Take note of the Raspberry Pi's private ip address.

   ```console
   pi@raspberrypi:~ $ hostname -I
   192.168.1.124
   ```

1. Download the application and start it.

   ```console
   pi@raspberrypi:~ $ curl -L -o espresso-2 https://github.com/takuma2/espresso-controller/releases/download/v0.1.0/espresso-2
   pi@raspberrypi:~ $ chmod +x espresso-2
   pi@raspberrypi:~ $ ./espresso-2 --help
   Control and monitor an espresso machine

   Usage:
     espresso [flags]

   Flags:
         --boiler-therm-clk-pin int    The GPIO pin connected to the boiler thermometer's max31855 clock (default 4)
         --boiler-therm-cs-pin int     The GPIO pin connected to the boiler thermometer's max31855 chip select, aka chip enable (default 3)
         --boiler-therm-miso-pin int   The GPIO pin connected to the boiler thermometer's max31855 data output (default 2)
     -h, --help                        help for espresso
     -p, --port string                 Port on which the espresso server should listen (default "8080")
     -r, --relay-pin int               The GPIO connected to the relay (default 21)
     -v, --verbose                     verbose output
   pi@raspberrypi:~ $ ./espresso -v

        ╓▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
       █▀─╓▄         ┌▄▄┌         ▄▄ ╙█
       █ ╫█╠█ ╔▓▓  ▄█▀╟█╙█▄  ╔▓▓ █▌╟█ ██▓▄
       █  ╙╙       █  ╘▀  █▄      ╙╙  █─▐█
       █ ╔█▀█ ╓▄▄  █▌    ▄█  ╓▄▄ ▓██▌ █▄██
       █  ▀▀▀       ╙▀▀▀▀╙       ╙▀▀└ █╨
       ╙▀██▓▓▓▓▓▓██▓▓▓▓▓▓▓▓██▓██▓▓▓██▀╙
    ╔▓▓▓▓██▓▓▓▓▓▓█▌        ▐▌ ▐█▓▓▓██▓▓▓▄
    ╫█▄▄▄▄▄▄▄▄▄▄▄██▄▄▄▄▄▄▄▄██      ▐█ ▄██▄
         ╫▌          █▌▄█          ▐█ █▌▐█
         ╫▌           └└─          ▐█ █▄▐█
         ╫▌      ╒▓▓▓▓▓▓▓▓▓▓▓▓▓▄   ▐█ ╙╙╙└
         ╫▌      ▐█▄▄▄▄▄▄▄▄██  █▌  ▐█
         ╫▌      ╘█        █▀▀▀▀   ▐█
         █▌       ╠█▓▄┌ ▄▄█▌       ▐█
       █▀╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙╙█
       █                              █─
       ╙▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀

   For more information, go to https://github.com/gregorychen3/espresso-controller

   2020-05-24T16:45:27.372-0400	INFO	espresso/server.go:115	Initializing gRPC server	{"port": 8080}
   2020-05-24T16:45:27.372-0400	INFO	espresso/server.go:123	Initializing gRPC web server	{"port": 8080}
   ```

   Or, start as a background process and leave it running.

   ```console
   pi@raspberrypi:~ $ ./espresso-2 &
   pi@raspberrypi:~ $ exit
   ```

### Control and monitor

Finally, visit `http://<ip_addr_from_step_2>:8080` using a web browser. Control and monitor the system from the dashboard there.

### Finished

![finished installation](images/finished_installation.jpg)

## Contributing

See the development [README](README_DEVELOPMENT.md).

## Credits

Logo icon made by [catkuro](https://www.flaticon.com/authors/catkuro) from [flaticon.com](https://www.flaticon.com) and converted to ASCII art using [asciiart.club](https://asciiart.club).
