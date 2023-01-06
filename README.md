# midimux

midimux is a MIDI multiplexer and proxy.


## What does it do ?

It takes MIDI input from various sources and forwards it to various destinations,
with both input and output being MIDI devices, virtual devices (software) or even remote midimux instances over UDP.


## SHOW ME.

First,
you can list all recognized devices:
```
% midimux -list
inputs:
0 - Scarlett 4i4 USB
1 - MPK Mini Mk II
2 - iRig Keys IO 49

outputs:
0 - Scarlett 4i4 USB
1 - MPK Mini Mk II
2 - iRig Keys IO 49
3 - FluidSynth virtual port (79151)
```

From this point,
you can run `midimux` and instruct it which inputs are to be listened to and where they should be forwarded.

The following example will read input from both `iRig` and `MPK` keyboards and write to `FluidSynth` and `Scarlett`:

```
% midimux -input iRig -input MPK -output FluidSynth -output Scarlett
```

Adding `-verbose` provides insight on what was received:
```
% midimux -input iRig -input MPK -output FluidSynth -output Scarlett -verbose
input: iRig Keys IO 49
input: MPK Mini Mk II
output: FluidSynth virtual port (79151)
output: Scarlett 4i4 USB

NoteOn channel: 0 key: 43 velocity: 75
NoteOn channel: 0 key: 46 velocity: 77
NoteOn channel: 0 key: 53 velocity: 83
NoteOn channel: 0 key: 50 velocity: 81
NoteOn channel: 0 key: 50 velocity: 0
NoteOn channel: 0 key: 46 velocity: 0
NoteOn channel: 0 key: 43 velocity: 0
NoteOn channel: 0 key: 53 velocity: 0
[...]
```

When `-input` is expressed as a `:port` or `address:port`,
then `midimux` will start a server and wait for MIDI packets over UDP.

When `-output` is expressed as a `:port` or `address:port`,
then `midimux` will emit MIDI packets to that destination over UDP.

In the following example,
`midimux` starts a server to receive input from the network and uses `FluidSynth` as its output:

```
% midimux -input :1053 -output FluidSynth -verbose
input: udp://[::]:1053
output: FluidSynth virtual port (79151)

NoteOn channel: 0 key: 43 velocity: 75
NoteOn channel: 0 key: 46 velocity: 77
NoteOn channel: 0 key: 53 velocity: 83
NoteOn channel: 0 key: 50 velocity: 81
NoteOn channel: 0 key: 50 velocity: 0
NoteOn channel: 0 key: 46 velocity: 0
NoteOn channel: 0 key: 43 velocity: 0
NoteOn channel: 0 key: 53 velocity: 0
[...]
```

While on a different machine,
`midimux` uses `iRig` as its input and both its local `FluidSynth` and a remote `midimux` for its output:

```
% midimux -input iRig -output FluidSynth -output 192.168.0.44:1053 -verbose
input: input: iRig Keys IO 49
output: FluidSynth virtual port (79151)
output: udp://192.168.0.44:1053

NoteOn channel: 0 key: 43 velocity: 75
NoteOn channel: 0 key: 46 velocity: 77
NoteOn channel: 0 key: 53 velocity: 83
NoteOn channel: 0 key: 50 velocity: 81
NoteOn channel: 0 key: 50 velocity: 0
NoteOn channel: 0 key: 46 velocity: 0
NoteOn channel: 0 key: 43 velocity: 0
NoteOn channel: 0 key: 53 velocity: 0
[...]
```

If no `-output` is provided,
then `midimux` will consume input and discard them unless `-verbose` is provided in which case it'll print midi messages to stdout.

It is possible to chain multiple `midimux` instances,
for example by having machine 1 take input from the network and output to a synthesizer,
machine 2 take input from the network and output to machine 1,
and machine 3 take input from `iRig` and output to machine 2:
the midi stream will cascade from the keyboard to the synthesizer (with a little latency :p).