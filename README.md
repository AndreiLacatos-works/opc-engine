# OPC engine

## What does it do?

Create an OPC server, designed to host nodes defined in a structured project file. Supports defining custom behavior for node values, enabling them to change dynamically over time based on user-defined rules or algorithms. Ideal for testing and simulating real-world conditions in a controlled environment.

Works best with [OPC Node designer](https://github.com/AndreiLacatos-works/opc-node-designer)

## Example

An example project is included that simulates a node with boolean values that toggle every 500ms, and another node with numeric (float) values that change based on a custom-defined rule.
