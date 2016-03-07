package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	//  "reflect
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func saveLog(loglevel string, content string) {
	day := time.Now().Format("2006-01-02") // must 2016-01-02 or time.Now().string()[0:10]
	logfileName := "/opt/docker_vm_info_" + day
	logfile, logfileErr := os.OpenFile(logfileName, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if logfileErr != nil {
		fmt.Printf("%s\r\n", logfileErr.Error())
		os.Exit(-1)
	}
	defer logfile.Close() //anytime dont forgot close file
	logger := log.New(logfile, "[INFO] ", log.LstdFlags)
	if loglevel == "info" {
	} else if loglevel == "fatal" {
		logger.SetPrefix("[FATAL] ")
	} else if loglevel == "warning" {
		logger.SetPrefix("[WARNING] ")
	} else {
		logger.SetPrefix("[UNKONW] ")
	}
	logger.Println(content)

}

func getAllContainerUuid() []string {
	var files []string
	dirPath := "/sys/fs/cgroup/cpuacct/system.slice/" // centos7+
	fileList, err := ioutil.ReadDir(dirPath)
	if err != nil {
		saveLog("fatal", "Cannot found vm: "+err.Error())
		os.Exit(-1)
	}

	for _, fi := range fileList {
		if strings.HasPrefix(fi.Name(), "docker-") && strings.HasSuffix(fi.Name(), "scope") {
			files = append(files, fi.Name()) //fi.Name() get file lists
		}
	}
	return files
}

func getContainerPid(ID string) string {

	regA := regexp.MustCompile(`docker-|.scope`)
	duuid := regA.ReplaceAllString(ID, "")

	cmd := exec.Command("docker", "inspect", "-f", "'{{.State.Pid}}'", duuid)
	out, err := cmd.Output()
	if err != nil {
		saveLog("warning", "Cann't get docker's uuid"+err.Error())
	}
	reg := regexp.MustCompile(`\n|'`)
	statePid := reg.ReplaceAllString(string(out), "")
	return statePid // like 7595
}

func getAllContainerStat(uuid string) map[string]map[string]int {
	cpu_res := getContainerCpuStat(uuid)
	//	mem_res := getContainerMemStat(uuid)	// no need twice, collect in main()
	net_res := getContainerNetStat(uuid)
	io_res := getContainerIoStat(uuid)
	//	all_info := map[string]map[string]int{"cpu": cpu_res, "mem": mem_res, "net": net_res, "io": io_res}	//delete mem
	all_info := map[string]map[string]int{"cpu": cpu_res, "net": net_res, "io": io_res}
	//fmt.Println(all_info)
	return all_info
}

func getContainerNetStat(filename string) map[string]int {
	//	regA := regexp.MustCompile(`docker-|.scope`)
	//	duuid := regA.ReplaceAllString(filename, "")

	statePid := getContainerPid(filename)

	res := map[string]int{}
	file := "/proc/" + statePid + "/net/dev" //not in /cgroup but in /proc, so we need docker process id
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		saveLog("warning", "Didn't found nic's file. "+err.Error())
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
		if strings.HasPrefix(line, "eth") {
			tmp := strings.Fields(line)
			//nic := strings.Replace(tmp[0], ":", "", -1)
			rbytes, _ := strconv.Atoi(tmp[1])
			tbytes, _ := strconv.Atoi(tmp[9])
			res = map[string]int{"rbytes": rbytes, "tbytes": tbytes}
		}
	}
	return res
}

func getContainerMemStat(uuid string) map[string]int {
	res := map[string]int{}
	file := "/sys/fs/cgroup/memory/system.slice/" + uuid + "/memory.stat"
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		fmt.Println(err)
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

	cpuAcctFile := "/sys/fs/cgroup/cpuacct/system.slice/" + uuid + "/cpuacct.stat"
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

	setFile := "/sys/fs/cgroup/cpuset/system.slice/" + uuid + "/cpuset.cpus"
	fset, setErr := os.Open(setFile)
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
	return res
}

func getContainerIoStat(uuid string) map[string]int {
	//	cmd := exec.Command("lsblk | grep 2518e0f3e5fe287b74279f | awk '{print $(NF-4)}'| sort -u")
	res := map[string]int{}
	dm := "253:9"
	file := "/sys/fs/cgroup/blkio/system.slice/" + uuid + "/blkio.throttle.io_service_bytes"
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
				read := readtmp / 1024 //#transform to KB
				res["read"] = read
			} else if tmp[1] == "Write" {
				writetmp, _ := strconv.Atoi(tmp[2])
				write := writetmp / 1024 //#transform to KB
				res["write"] = write
			}
			//			fmt.Println(res)
		}
	}
	return res
}

func getContainerFsStat(uuid string) map[string]int {
	statePid := getContainerPid(uuid)
	res := map[string]int{}
	//	nsenter --target 23104 --mount --uts --ipc --net --pid -- /bin/df -ahTP /
	cmd := exec.Command("timeout", "-s", "SIGKILL", "3s", "nsenter", "--target", statePid, "--mount", "--uts", "--ipc", "--net", "--pid", "--", "/bin/df", "-ahTP", "/") // just root directory,"/"
	out, err := cmd.Output()
	if err != nil {
		saveLog("fatal", statePid+"Read df info failed: "+err.Error())
		//		os.Exit(-1)
	}

	// Filesystem    Type  Size  Used Avail Use% Mounted on
	//	/dev/mapper/docker-253:0-9056391-17ea03785e078a621973bd9279f0d4b582a8bce3ba2012a8dded6e62a893637a ext4   99G  268M   94G   1% /

	cap := strings.Fields(string(out))
	reg := regexp.MustCompile(`G|%`)
	res["cap_ratio"], _ = strconv.Atoi(reg.ReplaceAllString(cap[13], ""))

	cmd = exec.Command("timeout", "-s", "SIGKILL", "3s", "nsenter", "--target", statePid, "--mount", "--uts", "--ipc", "--net", "--pid", "--", "/bin/df", "-iaP", "/")
	out_inode, inode_err := cmd.Output()

	// [Filesystem Inodes IUsed IFree IUse% Mounted on /dev/mapper/docker-252:3-1188858-46dcf1dd9c445b969dcb026e86df00cd21e115ba0e9d8ee22fced2a7694aae00 655360 11857 643503 2% /]

	if inode_err != nil {
		saveLog("fatal", "Read inode info failed: "+inode_err.Error())
		//		os.Exit(-1)
	}
	inode := strings.Fields(string(out_inode))
	res["inode_ratio"], _ = strconv.Atoi(reg.ReplaceAllString(inode[11], ""))
	//	fmt.Println(res)
	return res
}

func calculate(each_uuid string, dname string, ID []string, wg *sync.WaitGroup) {
	defer wg.Done()
	info1 := map[string]map[string]map[string]int{}
	info2 := map[string]map[string]map[string]int{}
	for _, each_uuid := range ID {
		mem_res := getContainerMemStat(each_uuid) // only once
		disk_res := getContainerFsStat(each_uuid)
		info1[each_uuid] = getAllContainerStat(each_uuid)
		info1[each_uuid]["mem"] = mem_res
		info1[each_uuid]["disk"] = disk_res
	}
	//	fmt.Println(info1)
	sleepTime := 3
	time.Sleep(3 * time.Second)
	for _, each_uuid := range ID {
		info2[each_uuid] = getAllContainerStat(each_uuid)
	}

	// CPU
	cpu_user := (info2[each_uuid]["cpu"]["user"] - info1[each_uuid]["cpu"]["user"]) / sleepTime
	cpu_system := (info2[each_uuid]["cpu"]["system"] - info1[each_uuid]["cpu"]["system"]) / sleepTime
	cpu_total := cpu_user + cpu_system
	cpu_num := info1[each_uuid]["cpu"]["cpunum"]
	cpu_quota := cpu_num * 100
	cpu_usage := cpu_total * 100 / cpu_quota
	//	cpu := fmt.Sprintf("|%s|check-vm-cpu|sys_user=%d&user=%d&sys=%d&total_ratio=%d&cpu_n=%d&quota=%d", dname, cpu_total, cpu_user, cpu_system, cpu_usage, cpu_num, cpu_quota)
	cpu := fmt.Sprintf("|%s|cpu_usg=%d&cpu_user=%d&cpu_sys=%d&cpu_ratio=%d&cpu_n=%d&quota=%d", dname, cpu_total, cpu_user, cpu_system, cpu_usage, cpu_num, cpu_quota)

	// MEM
	mem_rss := float64(info1[each_uuid]["mem"]["total_rss"]) / float64(1024)
	mem_limit := float64(info1[each_uuid]["mem"]["mem_limit"]) / float64(1024)
	mem_cache := float64(info1[each_uuid]["mem"]["total_cache"]) / float64(1024)
	mem_mapped_file := float64(info1[each_uuid]["mem"]["total_mapped_file"]) / float64(1024)
	rss_ratio := float64(mem_rss) / float64(mem_limit)
	//	mem := fmt.Sprintf("|%s|check-vm-mem|rss=%.2f&quota=%.2f&cache=%.2f&mapped=%.2f&ratio=%.2f", dname, mem_rss, mem_limit, mem_cache, mem_mapped_file, rss_ratio)
	mem := fmt.Sprintf("&mem_rss=%.2f&mem_quota=%.2f&mem_cache=%.2f&mem_mapped=%.2f&mem_ratio=%.2f", mem_rss, mem_limit, mem_cache, mem_mapped_file, rss_ratio)

	// DISK
	blkio_write := (float64(info2[each_uuid]["io"]["write"]) - float64(info1[each_uuid]["io"]["write"])) / float64(1024) / float64(sleepTime)
	blkio_read := (float64(info2[each_uuid]["io"]["read"]) - float64(info1[each_uuid]["io"]["read"])) / float64(1024) / float64(sleepTime)
	blkio := fmt.Sprintf("&io_write=%.2f&io_read=%.2f", blkio_write, blkio_read)

	// NET
	net_rbyte := (float64(info2[each_uuid]["net"]["rbytes"]) - float64(info1[each_uuid]["net"]["rbytes"])) * float64(8) / float64(1024) / float64(1024) / float64(sleepTime)
	net_tbyte := (float64(info2[each_uuid]["net"]["tbytes"]) - float64(info1[each_uuid]["net"]["tbytes"])) * 8.0 / float64(1024) / float64(1024) / float64(sleepTime)
	//	net := fmt.Sprintf("|%s|check-vm-net|in=%.2f&out=%.2f", dname, net_rbyte, net_tbyte)
	net := fmt.Sprintf("&net_in=%.2f&net_out=%.2f", net_rbyte, net_tbyte)

	// DISK
	capacity := info1[each_uuid]["disk"]["cap_ratio"]
	inode := info1[each_uuid]["disk"]["inode_ratio"]
	disk := fmt.Sprintf("&cap=%d&inode=%d", capacity, inode)
	saveLog("info", cpu+mem+blkio+net+disk)
	return

}

func main() {
	ID := getAllContainerUuid() //docker-xxx.scope
	wg := new(sync.WaitGroup)

	// Begin math
	for _, each_uuid := range ID {

		//		println(each_uuid)
		// GET DOCKER NAME
		regA := regexp.MustCompile(`docker-|.scope`)
		duuid := regA.ReplaceAllString(each_uuid, "")
		cmd := exec.Command("docker", "inspect", "-f", "'{{.Name}}'", duuid)
		out, err := cmd.Output()
		if err != nil {
			println("each_uuid")

			fmt.Println(err)
		}
		regB := regexp.MustCompile(`\n|/|'`)
		dname := regB.ReplaceAllString(string(out), "") // like docker13808,10.10.138.8
		wg.Add(1)
		go calculate(each_uuid, dname, ID, wg)
	}
	wg.Wait()
}
