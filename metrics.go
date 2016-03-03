package main

import (
	//	"encoding/json"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"reflect"
	"strconv"
	"strings"
)

func getAllContainerUuid() []string {
	var files []string
	//	dirPath := "/sys/fs/cgroup/cpuacct/system.slice/" // centos7+
	dirPath := "/cgroup/cpuacct/docker/" // centos6...
	filesTmp, err := ioutil.ReadDir(dirPath)
	if err != nil {
		fmt.Println("read file list failed!")
		fmt.Println(err)
	}
	for _, fi := range filesTmp {
		if !fi.IsDir() {
			continue
		}
		//		fmt.Println(fi)
		//		fmt.Println(reflect.TypeOf(fi))
		files = append(files, fi.Name())
	}
	//	fmt.Println(files)
	return files
}
func getContainerInfo(ID string) []string {
	//	var dat map[string]interface{}
	//	cmd := exec.Command("/usr/bin/docker", "inspect ", ID)
	//fmt.Println(ID)
	cmd := exec.Command("docker", "inspect", "-f", "'{{.State.Pid}}'", ID)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(reflect.TypeOf(out))
	println("---------------")
	fmt.Println(out)
	outString := strings.Replace(string(out), "\n", " ", -1)
	//	fmt.Println(out["Id"])
	println("---------------")
	fmt.Println(string(outString))
	outString = strings.Replace(string(outString), "'", "", -1)
	var res []string
	res = append(res, outString)
	fmt.Println(res)
	return res
	//	if err := json.Unmarshal(out, &dat); err != nil {
	//		//		panic(err)
	//		fmt.Println(err)
	//	}
	//	fmt.Println(dat)
}
func getAllContainerStat(uuid string) {
	cpu_res := getContainerCpuStat(uuid)
	mem_res := getContainerMemStat(uuid)
	net_res := getContainerNetStat(uuid)
	io_res := getContainerIoStat(uuid)
	all_info := map[string]map[string]int{"cpu": cpu_res, "mem": mem_res, "net": net_res, "io": io_res}
	fmt.Println(all_info)
}
func getContainerNetStat(uuid string) map[string]int {
	cmd := exec.Command("docker", "inspect", "-f", "'{{.State.Pid}}'", uuid)
	out, err := cmd.Output()
	if err != nil {
		fmt.Println(err)
	}
	outString := strings.Replace(string(out), "\n", "", -1)
	outString = strings.Replace(outString, "'", "", -1)
	res := map[string]int{}
	file := "/proc/" + outString + "/net/dev"
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		os.Exit(1)
	}
	buf := bufio.NewReader(f)
	for {
		line, lerr := buf.ReadString('\n')
		if lerr != nil || io.EOF == lerr {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Inter-") || strings.HasPrefix(line, "face") || strings.HasPrefix(line, "lo") {
			continue
		}
		if strings.HasPrefix(line, "eth0") {
			tmp := strings.Fields(line)
			//			nic := strings.Replace(tmp[0], ":", "", -1)
			rbytes, _ := strconv.Atoi(tmp[1])
			tbytes, _ := strconv.Atoi(tmp[9])
			res = map[string]int{"rbytes": rbytes, "tbytes": tbytes}
		}
	}
	//	fmt.Println(res)
	return res
}
func getContainerMemStat(uuid string) map[string]int {
	res := map[string]int{}
	file := "/cgroup/memory/docker/" + uuid + "/memory.stat"
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		os.Exit(1)
	}
	buf := bufio.NewReader(f)
	for {
		line, lerr := buf.ReadString('\n')
		if lerr != nil || io.EOF == lerr {
			break
		}
		line = strings.TrimSpace(line)
		tmp := strings.Split(line, " ")
		if strings.HasPrefix(line, "hierarchical_memory_limit") {
			num, _ := strconv.Atoi(tmp[1])
			res["mem_limit"] = num / 1024
		} else if strings.HasPrefix(line, "total_cache") {
			num, _ := strconv.Atoi(tmp[1])
			res["total_cache"] = num / 1024
		} else if strings.HasPrefix(line, "total_rss") {
			num, _ := strconv.Atoi(tmp[1])
			res["total_rss"] = num / 1024
		} else if strings.HasPrefix(line, "total_mapped_file") {
			num, _ := strconv.Atoi(tmp[1])
			res["total_mapped_file"] = num / 1024
		}
	}
	//	fmt.Println(res)
	return res
}

func getContainerCpuStat(uuid string) map[string]int {
	res := map[string]int{}
	//	res["uuid"] = uuid

	cpuAcctFile := "/cgroup/cpuacct/docker/" + uuid + "/cpuacct.stat"
	facct, acctErr := os.Open(cpuAcctFile)
	defer facct.Close()
	if acctErr != nil {
		fmt.Println(acctErr)
		os.Exit(1)
	}
	bufAcct := bufio.NewReader(facct)
	i := 1
	for {
		line, lerr := bufAcct.ReadString('\n')
		line = strings.TrimSpace(line)
		if lerr != nil || io.EOF == lerr {
			break
		}
		if i == 1 {
			info := strings.Split(line, " ")
			res["user"], _ = strconv.Atoi(info[1])
		} else if i == 2 {
			info := strings.Split(line, " ")
			res["system"], _ = strconv.Atoi(info[1])
		}
		i++
	}

	cpuSetFile := "/search/cpunum"
	//	setFile := "/cgroup/cpuset/docker/" + uuid + "/cpuset.cpus"
	fset, setErr := os.Open(cpuSetFile)
	defer fset.Close()
	if setErr != nil {
		fmt.Println(setErr)
		os.Exit(1)
	}
	bufSet := bufio.NewReader(fset)
	j := 1
	for {
		jine, jerr := bufSet.ReadString('\n')
		jine = strings.TrimSpace(jine)
		if jerr != nil || io.EOF == jerr {
			break
		}
		tmp := strings.Split(jine, "-")
		startCpu, _ := strconv.Atoi(tmp[0])
		endCpu, _ := strconv.Atoi(tmp[1])
		cpuNum := endCpu - startCpu + 1
		res["cpunum"] = cpuNum
		j++
	}
	//	fmt.Println(res)
	return res
}

func getContainerIoStat(uuid string) map[string]int {
	//	cmd := exec.Command("lsblk | grep 2518e0f3e5fe287b74279f | awk '{print $(NF-4)}'| sort -u")
	res := map[string]int{}
	dm := "253:2"
	file := "/cgroup/blkio/docker/" + uuid + "/blkio.throttle.io_service_bytes"
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
	}
	buf := bufio.NewReader(f)
	for {
		line, lerr := buf.ReadString('\n')
		if lerr != nil || io.EOF == lerr {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, dm) {
			tmp := strings.Split(line, " ")
			if tmp[1] == "Read" {
				readtmp, _ := strconv.Atoi(tmp[2])
				read := readtmp / 1024
				res["read"] = read
			} else if tmp[1] == "Write" {
				writetmp, _ := strconv.Atoi(tmp[2])
				write := writetmp / 1024
				res["write"] = write
			}
			//			fmt.Println(res)
		}
	}
	//	fmt.Println(res)
	return res
}
func main() {
	ID := getAllContainerUuid()
	for _, each_uuid := range ID {
		getAllContainerStat(each_uuid)
	}

}
