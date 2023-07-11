package main

import (
	"abobus/gopkg/ipc"
	"abobus/gopkg/vpn"
	"fmt"
	"log"
	"net"
	"os"
)

const DefaultSockAddr = "/tmp/vpnchains.socket"
const InjectedLibPath = "/usr/lib/libvpnchains_inject.so"
const BufSize = 100500

func errorMsg(path string) string {
	return "Usage: " + path +
		" <command> [command args...]"
}

func handleIpc(ready chan struct{}, config *vpn.WireguardConfig) {
	err := os.Remove(DefaultSockAddr)
	if err != nil {
		log.Fatalln(err)
	}

	var buf = make([]byte, BufSize)

	conn := ipc.NewConnection(DefaultSockAddr)
	handler := func(conn net.Conn) {
		n, err := conn.Read(buf)
		requestBuf := buf[:n]

		if err != nil {
			log.Fatalln(err)
		}

		responseBuf, err := ipc.HandleRequest(requestBuf)
		if responseBuf == nil && err != nil {
			log.Fatalln(err) // вроде как невозможно
		} else if err != nil {
			log.Println(err, ". Returning error response.")
		}

		_, err = conn.Write(responseBuf)
		if err != nil {
			log.Println(err)
		}
	}

	ready <- struct{}{}
	err = conn.Listen(handler)
	if err != nil {
		log.Println("sldfadsf")
		log.Fatalln(err)
	}
}

func main() {
	args := os.Args
	if len(args) < 3 {
		fmt.Println(errorMsg(args[0]))
		os.Exit(0)
	}

	wireguardConfigPath := args[1]
	commandPath := args[2]
	commandArgs := args[3:]

	config, err := vpn.WireguardConfigFromFile(wireguardConfigPath)
	if err != nil {
		log.Fatalln(err)
	}

	cmd := ipc.CreateCommandWithInjectedLibrary(InjectedLibPath, commandPath, commandArgs)

	ready := make(chan struct{})
	go handleIpc(ready, config)

	<-ready
	err = cmd.Start()
	if err != nil {
		log.Fatalln(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatalln("subprocess says, ", err)
	}

	err = os.Remove(DefaultSockAddr)
	if err != nil {
		log.Println(err)
	}
}
