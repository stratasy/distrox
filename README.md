# Distrox

## Instructions

Ensure that [golang](https://golang.org/dl/) is installed and setup correctly.

Run ```make``` to build. A binary (`distrox`) will be generated.

Usage:
``` ./distrox [host] [port] [is_leader] ```

Example:
``` ./distrox localhost 8081 true ```

To connect one node to another node, type:
``` connect [host]:[port] ``` in the console while the node is running.
