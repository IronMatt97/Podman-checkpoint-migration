# Podman checkpoint/restore migration process
This is archive is a demo aimed to test CRIU's checkpoint and restore procedure. In particular, the tool is used by Podman's API.
## Functionality
What happens in this demo is pretty simple. A client sends an http request to an external host, asking to add +1 to a number.
At that point the client has 3 seconds to request the container migration (before the result is ready to query).
When the migration is requested, the first host checkpoints its container, and sends it to the other one. The latter will restore the execution and communicate the result back to the new node.
The client can then query the result to the second host.
## How to use:
- Step 1: prepare two pre-configured VMs in order to simulate external hosts. For this demo, VMs require:
  - Podman -> Getting started https://podman.io/getting-started/
  - CRIU last build -> https://github.com/checkpoint-restore/criu
  - Python and Go as distribution packages
  - Make sure that you are running cgroups v2 and using runc instead of crun as ociRuntime.
  - Remember that Criu is a kernel level tool, and requires root permissions in order to checkpoint and restore. 
- Step 2: go to the main folder, set the config.ini file to have correct VMs addresses and run:
> make
- Step 3: move the main folder onto both VMs, and run the script:
>./initializeDemo.sh
- Step 4: respectively on host A and host B run from the main folder:
> go run node/node.go -executor
> go run node/node.go

Only one the first machine there will be an executor container at the beginning.
At this point, from the client on the base machine you can communicate with the external hosts and proceed with the prompt.
