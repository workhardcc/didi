package main

import (
	"bufio"
	"encoding/json"
	//	"reflect"
	//	"flag" //for glog
	"fmt"
	//	"glog"
	"log"
	"net"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

type Content struct {
	Act    string
	Detail map[string]string
}

//type Log struct {
//	Loglevel string
//	content  string
//	err      func(err error)
//}

func saveLog(loglevel string, content string) bool {
	time := time.Now().String()
	timeSlice := strings.Split(time, " ")
	day := timeSlice[0]
	dayTime := strings.Split(day, ".")
	tmpfile := "/search/go/server.go_" + dayTime[0]
	logfile, logfileErr := os.OpenFile(tmpfile, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logfileErr != nil {
		fmt.Printf("%s\r\n", logfileErr.Error())
		os.Exit(-1)
	}
	defer logfile.Close()
	logger := log.New(logfile, "[INFO] ", log.LstdFlags)
	if loglevel == "info" {
		logger.Println(content)
	} else if loglevel == "fatal" {
		logger.SetPrefix("[WARNING] ")
		logger.Println(content)
	}
	return true
}

func allowedIp(ip string, orderlist []string) bool {
	target := ip
	i := sort.Search(len(orderlist), func(i int) bool { return orderlist[i] >= target })
	if i < len(orderlist) && orderlist[i] == target {
		return true
	}
	return false
}

func handleConnection(conn net.Conn) {
	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		saveLog("fatal", "get client data error: ")
	}

	var info Content
	if err := json.Unmarshal([]byte(data), &info); err != nil {
		fmt.Println(err)
	}
	//	fmt.Println(data)
	//	fmt.Println(info)
	//	fmt.Println(reflect.TypeOf(data))
	//	fmt.Println(reflect.TypeOf(info))
	if info.Act == "create" {
		cmd := exec.Command("/bin/bash", "/search/go/create_docker_vm.sh")
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			saveLog("fatal", "create vm failed!")
		} else {
			saveLog("info", "create vm succeed: "+data)
		}
	} else if info.Act == "delete" {
		cmd := exec.Command("/bin/bash", "/search/go/delete_docker_vm.sh")
		err := cmd.Run()
		if err != nil {
			fmt.Println(err)
			saveLog("fatal", "delete vm failed!")
		} else {
			saveLog("info", "delete vm successed!"+data)
		}
	} else {
		saveLog("info", "do sth2!")
	}
	//	fmt.Printf("%#v\n", data)

	fmt.Fprintf(conn, "received\n")
	conn.Close()
}

func main() {
	orderlist := []string{"127.0.0.1", "10.x.x.x"}
	sort.Strings(orderlist)
	ln, err := net.Listen("tcp", ":6010")
	if err != nil {
		panic(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			saveLog("fatal", "get client connection error: ")
		}
		remote_ip := conn.RemoteAddr().String()
		tmpIp := strings.Split(remote_ip, ":")
		remote_ip = tmpIp[0]
		if !allowedIp(remote_ip, orderlist) {
			saveLog("fatal", "<"+remote_ip+"> This ip not allowed access!")
			fmt.Fprintf(conn, "This ip not allowed!")
			conn.Close()

			continue
		}
		go handleConnection(conn)
	}
}
