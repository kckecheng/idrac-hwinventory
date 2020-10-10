About
=======

Detect hardware basic inforamtion by using iDRAC without touching OS.

Prerequisites
--------------

- SSH access for iDRAC

Usage
-------

Download the binary from the release page or build it as below:

::

  go build .
  ./idrac-hwinventory --help
  # Show active NIC and FC speed
  ./idrac-hwinventory -i 192.168.100.100 -f LinkSpeed -f PortSpeed
  # Show all information
  ./idrac-hwinventory -i 192.168.100.100 -t all
  # Show all device types
  ./idrac-hwinventory -i 192.168.100.100 -t all -f "Device Type" | grep 'Device Type' | sort | uniq
