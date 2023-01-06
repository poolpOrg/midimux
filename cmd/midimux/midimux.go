/*
 * Copyright (c) 2023 Gilles Chehade <gilles@poolp.org>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"gitlab.com/gomidi/midi/v2"

	"gitlab.com/gomidi/midi/v2/drivers"
	_ "gitlab.com/gomidi/midi/v2/drivers/rtmididrv"
)

type multiValue struct {
	options []string
}

func main() {

	//var opt_connect string
	//var opt_listen string
	var opt_verbose bool
	var opt_list bool

	var opt_inputs []string
	var opt_outputs []string

	flag.BoolVar(&opt_list, "list", false, "list input and output ports")
	flag.BoolVar(&opt_verbose, "verbose", false, "verbose")
	//flag.StringVar(&opt_connect, "connect", "", "endpoint to a midimux server (localhost:1057)")
	//flag.StringVar(&opt_listen, "listen", "", "interface for accepting midimux clients (:1057)")

	flag.Func("input", "add input device", func(value string) error {
		opt_inputs = append(opt_inputs, value)
		return nil
	})
	flag.Func("output", "add output device", func(value string) error {
		opt_outputs = append(opt_outputs, value)
		return nil
	})

	flag.Parse()

	if opt_list {
		fmt.Println("inputs:")
		for i, name := range midi.GetInPorts() {
			fmt.Println(i, "-", name)
		}
		fmt.Println()

		fmt.Println("outputs:")
		for i, name := range midi.GetOutPorts() {
			fmt.Println(i, "-", name)
		}
		fmt.Println()
	}

	var udpEndpoint *net.UDPAddr

	var inPort drivers.In
	var inPorts []drivers.In
	var inConn net.PacketConn
	var inConns []net.PacketConn

	var outPort drivers.Out
	var outPorts []drivers.Out
	var outConn *net.UDPConn
	var outConns []*net.UDPConn

	var err error

	for _, opt_input := range opt_inputs {
		_, err = net.ResolveUDPAddr("udp", opt_input)
		if err == nil {
			inConn, err = net.ListenPacket("udp", opt_input)
			if err != nil {
				log.Fatal(err)
			}
			defer inConn.Close()
			inConns = append(inConns, inConn)
			if opt_verbose {
				fmt.Println("input:", fmt.Sprintf("udp://%s", inConn.LocalAddr().String()))
			}
		} else {
			inputPortID, err := strconv.Atoi(opt_input)
			if err != nil {
				inPort, err = midi.FindInPort(opt_input)
				if err == nil {
					if !inPort.IsOpen() {
						err = inPort.Open()
					}
				}
			} else {
				inPort, err = midi.InPort(inputPortID)
			}
			if err != nil {
				log.Fatal(err)
			}
			inPorts = append(inPorts, inPort)
			if opt_verbose {
				fmt.Println("input:", inPort)
			}
		}
	}

	for _, opt_output := range opt_outputs {
		udpEndpoint, err = net.ResolveUDPAddr("udp", opt_output)
		if err == nil {
			outConn, err = net.DialUDP("udp", nil, udpEndpoint)
			if err != nil {
				log.Fatal(err)
			}
			outConns = append(outConns, outConn)
		} else {

			outputPortID, err := strconv.Atoi(opt_output)
			if err != nil {
				outPort, err = midi.FindOutPort(opt_output)
				if err == nil {
					if !outPort.IsOpen() {
						err = outPort.Open()
					}
				}
			} else {
				outPort, err = midi.OutPort(outputPortID)
			}
			if err != nil {
				log.Fatal(err)
			}
			if opt_verbose {
				fmt.Println("output:", outPort)
			}
			outPorts = append(outPorts, outPort)
		}
	}

	done := make(chan bool)

	for _, inPort := range inPorts {
		go func(_in drivers.In) {
			_, err := midi.ListenTo(_in, func(msg midi.Message, absms int32) {
				for _, outPort := range outPorts {
					go outPort.Send(msg.Bytes())
				}
				for _, outConn := range outConns {
					go outConn.Write(msg.Bytes())
				}
				if opt_verbose {
					fmt.Println(msg)
				}
			})
			if err != nil {
				log.Fatal(err)
			}
		}(inPort)
	}

	for _, inConn := range inConns {
		go func(_in net.PacketConn) {
			for {
				buf := make([]byte, 1024)
				n, _, err := _in.ReadFrom(buf)
				if err != nil {
					continue
				}
				var msg midi.Message = buf[:n]
				for _, outPort := range outPorts {
					go outPort.Send(msg.Bytes())
				}
				for _, outConn := range outConns {
					go outConn.Write(msg.Bytes())
				}
				if opt_verbose {
					fmt.Println(msg)
				}
			}
		}(inConn)
	}

	<-done
}
