# Model Based Arrowhead framework Systems

The mbaigo systems repository contains a collection of Arrowhead Compliant systems that rely on the [mbaigo library](https://github.com/vanDeventer/mbaigo) (A model based Arrowhead implementation in Go).

The structure of each system is the a husk (or main.go) that enwraps the unit asset or thing (thing.go). The thing.go file promotes the assets's resources, which the capsule exposes as services.

The main function instantiates the software model (or structs) of the system, the capsule, the thing with its resources, the hosting device the capsule is running on, and all the services it exposes.

The thing holds the functions to communicate with hardware, databases, or algorithm (which at that point is the thing itself and holds the intellectual property of the developer).

All systems/capsules rely on a configuration file, which it looks for in the directory it is looking for (systemconfig.json). Since the name is the same, no two capsules should be in the same directory. If the capsule does not find the systemconfig.json, it will create it and shutdown to enable the deployment technician to configure the file. The system will run upon restarting the capsule.

## Documentation
There are three types of documentations, of which two are computer generated and therefore always up to date.

	- GitHub README.md files
	- system web server that provide a black box description of the system, its resouces and all their services. It is accessible using a stanadard web browser at the address provided by the system at start up.
	- White box documentation available through a stanard web browser using Go Doc webserve on one's computer or the one at [https://pkg.go.dev](https://pkg.go.dev).