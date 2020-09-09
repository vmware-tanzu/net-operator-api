# net-operator-api

## Overview

The net-operator-api project provides the object model and generated client libraries
for the Net Operator project, which is a part of vSphere's [Kubernetes](https://kubernetes.io)
support.

Net Operator allows users to manage the lifecycle of vSphere Networks and Virtual Machine NetworkInterfaces
using a declarative Kubernetes consumption model and is an integral part of Project Pacific
in vSphere 7 with Kubernetes.

The NetworkInterface object can be used as the definition for a client-managed NetworkInterface in a
VirtualMachine Spec.

## Use cases

The use cases of net-operator-api are currently limited to 3rd party integrations with vSphere with Kubernetes.

In vSphere with Kubernetes it is not currently possible to create new NetworkInterface using this API, but we
hope to expand on this functionality over time.

## Contributing

We welcome new contributors who are interested in collaborating on our Kubernetes support for vSphere.

More information [here](CONTRIBUTING.md)

## Getting Started

Check out how to get started with net-operator-api [here](GETTING-STARTED.md)

## Roadmap

Use of the net-operator-api is currently only supported in vSphere with Kubernetes.
However the intention is to make it more widely available as a Kubernetes-native
way of interacting with vSphere VMs.

## Maintainers

Current list of project maintainers [here](MAINTAINERS.md)

## License

Net Operator API is licensed under the [Apache License, version 2.0](LICENSE)
