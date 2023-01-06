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

func main() {
	var opt_input string
	var opt_output string
	var opt_connect string
	var opt_listen string
	var opt_verbose bool
	var opt_list bool

	flag.BoolVar(&opt_list, "list", false, "list input and output ports")
	flag.StringVar(&opt_input, "input", "", "input device port")
	flag.StringVar(&opt_output, "output", "", "output device port")
	flag.BoolVar(&opt_verbose, "verbose", false, "verbose")
	flag.StringVar(&opt_connect, "connect", "", "endpoint to a midimux server (localhost:1057)")
	flag.StringVar(&opt_listen, "listen", "", "interface for accepting midimux clients (:1057)")
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

	var inPort drivers.In
	var outPort drivers.Out
	var err error

	if opt_input != "" {
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
		if opt_verbose {
			fmt.Println("input:", inPort)
		}
	}

	if opt_output != "" {
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
	}

	var udpEndpoint *net.UDPAddr
	var udpConnection *net.UDPConn
	if opt_input != "" && opt_connect != "" {
		udpEndpoint, err = net.ResolveUDPAddr("udp", opt_connect)
		if err != nil {
			log.Fatal(err)
		}
		udpConnection, err = net.DialUDP("udp", nil, udpEndpoint)
		if err != nil {
			log.Fatal(err)
		}
	}

	var packetConn net.PacketConn
	if opt_listen != "" {
		packetConn, err = net.ListenPacket("udp", opt_listen)
		if err != nil {
			log.Fatal(err)
		}
		defer packetConn.Close()
	}

	if opt_input == "" && opt_listen == "" {
		log.Fatal("no input, no output, what am I expected to do ?")
	}

	done := make(chan bool, 0)

	if opt_input != "" {
		go func() {
			_, err = midi.ListenTo(inPort, func(msg midi.Message, absms int32) {
				if opt_output != "" {
					go outPort.Send(msg.Bytes())
				}
				if udpConnection != nil {
					go udpConnection.Write(msg.Bytes())
				}
				if opt_verbose {
					fmt.Println(msg)
				}
			})
			if err != nil {
				log.Fatal(err)
			}
		}()
	}

	if opt_listen != "" {
		for {
			buf := make([]byte, 1024)
			n, _, err := packetConn.ReadFrom(buf)
			if err != nil {
				continue
			}
			var msg midi.Message = buf[:n]
			if opt_output != "" {
				go outPort.Send(msg.Bytes())
			}
			if opt_verbose {
				fmt.Println(msg)
			}
		}
	}

	<-done
}
