package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net"
	"time"
)

type Content struct {
	Act    string
	Detail map[string]string
}

func main() {
	conn, err := net.DialTimeout("tcp", "127.0.0.1:6010", 5*time.Second)
	if err != nil {
		panic(err)
	}

	create := Content{"create", map[string]string{"cpu": "2", "mem": "4"}}
	//	tmp := map[string]string{"act": "create", "cpu": "2", "mem": "4"}
	operate, err := json.Marshal(create)
	if err != nil {
		fmt.Println(err)
		return
	}
	t := string(operate)
	fmt.Println(t)
	fmt.Fprintf(conn, t+"\n")
	//	fmt.Fprintf(conn, "cc945cc\n")

	data, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		panic(err)
	}

	fmt.Printf("%#v\n", data)
}
