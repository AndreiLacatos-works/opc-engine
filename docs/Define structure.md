# Defining custom OPC node structure

The project file is structured as a JSON that primarily defines the layout of OPC nodes. It supports two types of OPC nodes:

- **Container Nodes** (used for organizational purposes)
- **Value Nodes** (nodes with values defined by a specific data type)

A **container node** is characterized by an ID, a label, and a list of child nodes, which can be either value nodes or other container nodes. A **value node** is defined by an ID, a label, and a waveform. The waveform specifies how the value of the node evolves over time. For more information on configuring the waveform, refer to [Define node behavior](Define%20node%20behavior.md). Currently, only boolean and float data types are supported.
