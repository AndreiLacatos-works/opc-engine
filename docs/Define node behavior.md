# Defining the behavior of a node

To define how a value node evolves over time, you must configure the waveform property of the OPC node. This property is composed of the following components:

- **Duration**: Specifies the total length of time (in milliseconds) over which the value will change
- **Tick Frequency**: Defines the frequency (in milliseconds) at which the value changes within the given duration
- **Transitions**: Describes the specific changes or shifts in the value at designated intervals

Once the duration period has elapsed, the waveform process is replayed from the beginning, creating an endless loop. The tick frequency determines the "heartbeat" rate at which values are updated, while the transitions specify the exact values the node holds at any given time.

For simple boolean transitions, only the timestamp of the transition needs to be defined. However, for numeric types, the value at each transition is also required. When a node's value is queried between two transitions, the value of the last transition is returned, effectively acting as a holding register.

By adjusting these components, you can simulate a dynamic value that changes according to your desired behavior, allowing for realistic time-based data modeling in your OPC UA server simulation.
