# Healthchecker

Simple network health checker service. Great to monitor the health of your website or server. Written as way to exerise Go. This is NOT a production service :).

Checks available:
- ICMP check
- HTTP request check

Data outputs available (sinks):
- terminal/file
- influxdb udp

To use:

- Build the binary.
- Modify the example configuration file.
- Run the binary with the config file as a flag.
