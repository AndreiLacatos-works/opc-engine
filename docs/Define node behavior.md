# Defining the behavior of a node

To define how a value node evolves over time, you must configure the waveform property of the OPC node. This property is composed of the following components:

- **Duration**: Specifies the total length of time (in milliseconds) over which the value will change
- **Tick Frequency**: Defines the frequency (in milliseconds) at which the value changes within the given duration
- **Transitions**: Describes the specific changes or shifts in the value at designated intervals

Once the duration period has elapsed, the waveform process is replayed from the beginning, creating an endless loop. The tick frequency determines the "heartbeat" rate at which values are updated, while the transitions specify the exact values the node holds at any given time.

For simple boolean transitions, only the timestamp of the transition needs to be defined. However, for numeric types, the value at each transition is also required. When two data points are separated by at least one tick, value of the tick(s) between the data points are defined by the smoothing behavior. 3 strategies are supported:

- steps ("step")
- linear interpolation ("linear")
- cubic spline ("cubic")

Step strategy does not apply any smoothing, it acts as a holding register. Linear strategy takes into account the number of intermediary ticks and the delta between the two values; it simulates a linear stransition. The cubic spline strategy uses a cubic spline polynomial expression to provide smooth transitions. When a node's value is queried between two ticks, the value associated with the last tick is returned.

By adjusting these components, you can simulate a dynamic value that changes according to your desired behavior, allowing for realistic time-based data modeling in your OPC UA server simulation.
