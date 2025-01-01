# OPC engine

## What does it do?

Create an OPC server, designed to host nodes defined in a structured project file. Supports defining custom behavior for node values, enabling them to change dynamically over time based on user-defined rules or algorithms. Ideal for testing and simulating real-world conditions in a controlled environment.

For detailed documentation on defining OPC nodes and behavior refer to [Define structure](docs/Define%20structure.md) and [Define node behavior](docs/Define%20node%20behavior.md)

Works best with [OPC Node designer](https://github.com/AndreiLacatos-works/opc-node-designer), provides a graphical interface to manage node configuration.

## Example

An example project is included that simulates a node with boolean values that toggle every 500ms, and another node with numeric (float) values that change based on a custom-defined rule.

## Installation

Install the latest docker image from the releases page, then load it:

```sh
docker load -i opc-engine-simulator-v0.1.0.tar
```

Launch the container via

```sh
docker-compose up
```

Note: when updating, make sure to remove the old container before launching the new version
