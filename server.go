package main

import (
	"bufio"
	"encoding/json"
	//	"flag" //for glog
	"fmt"
	//	"glog"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"
)

type Content struct {
	Act    string
	Detail map[string]string
}

func handleConnection(conn net.Conn) {
	//	flag.Parse()
	//	defer glog.Flush()

	time := time.Now().String()
	timeSlice := strings.Split(time, " ")
	day := timeSlice[0]
	dayTime := strings.Split(day, ".")
	//	curtime := dayTime[0]
	tmpfile := "/search/go/server.go_" + dayTime[0]
	logfile, logfileErr := os.OpenFile(tmpfile, os.O_RDWR|os.O_CREATE, 0)
	if logfileErr != nil {
		fmt.Printf("%s\r\n", logfileErr.Error())
		os.Exit(-1)
	}
	defer logfile.Close()
	logger := log.New(logfile, "\r\n", log.Ldate|log.Ltime|log.Llongfile)
	//	glog.Fatal("get client data error: 111")

	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		logger.Fatal("get client data error: ", err)
		//		glog.Fatal("get client data error: 222")
	}

	var info Content
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		fmt.Println(err)
	}
	fmt.Println(info)
	if info.Act == "create" {
		println("do sth0")
		cmd := exec.Command("/bin/bash", "/search/go/create_docker_vm.sh")
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println("operate failed!")
		}
		logger.Println("operate succeed!")
	} else if info.Act == "delete" {
		println("do sth1")
		cmd := exec.Command("/bin/bash", "/search/delete_docker_vm.sh")
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			fmt.Println("operate failed!")
		}
	} else {
		println("do sth2")
	}
	//	fmt.Printf("%#v\n", data)

	fmt.Fprintf(conn, "received\n")
	conn.Close()
}

func main() {
	ln, err := net.Listen("tcp", ":6010")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal("get client connection error: ", err)
		}

		go handleConnection(conn)
	}
}
