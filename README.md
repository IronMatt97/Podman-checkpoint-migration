# Podman checkpoint/restore migration process
This archive contains a Demo program aimed to test CRIU's checkpoint and restore procedure. In particular, the tool is used by Podman's API.
## Scope
What happens in this Demo is pretty simple. A client sends an http request to an external host, asking to add +1 to a number.
At that point the client has 3 seconds to request the container migration (before the result is ready to query).
When the migration is requested, the first host checkpoints its container, and sends it to the other one. The latter will restore the execution and communicate the result back to the new node.
The client can then query the result to the second host.
## How to use:
### Step 1: Configure VMs
First things first: prepare a virtual machine (Ubuntu is preferred) and **clone it at the end of the procedure**. The two pre-configured VMs will be our migration hosts. For this demo, VMs require:
- Podman: getting started https://podman.io/getting-started/
- CRIU (last build): https://github.com/checkpoint-restore/criu
- Python distribution packages: <code>sudo apt install python3</code>
- Golang (last version): https://go.dev/doc/install
- Make sure that you are running cgroups v2 and using runc instead of crun as ociRuntime. Check this issue in case https://github.com/checkpoint-restore/criu/issues/2000

Remember that Criu is a kernel level tool, and **requires root permissions** in order to checkpoint and restore. 
### Step 2: Set the config.ini
Go to the main folder, set the config.ini file content to have correct VMs addresses and run:
> make

This will create the image of the container that will be migrated.
### Step 3: Initialize the demo
Move the main folder onto both VMs, and run the script:
>./initializeDemo.sh

The command will load the image created in the previous step on the VMs (remember to use root privileges).
### Step 4: Start
Respectively on host A and host B run from the main folder:
> go run node/node.go -executor

> go run node/node.go

Doing this, on machine A a container will be spawned, differently from the machine B.
At this point, from the client on the base machine you can communicate with the external hosts and proceed with the prompt.
> go run client/client.go

Basically:
- Insert a number
- Request migration (before 3 seconds)
- Query the final result on the other node
