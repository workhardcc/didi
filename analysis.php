#!/usr/bin/env php
<?php
date_default_timezone_set('Asia/Shanghai'); 
@$cur_day=date("Y-m-d");
@$cur_day_hour=date("Y-m-d_H");
@$cur_time=date("H:i:s");
@$timestamp=date("Y/m/d H:i:s",(strtotime("now")-60));
$alarm_url="http://www.baidu.com/notify";

exec('docker ps | awk "{print \$NF}" | fgrep -v NAMES | wc -l ',$output);
$vm_num=$output[0];

$logfile = "/home/monitor/logs/docker_vm_info_$cur_day_hour";

#$reg= "/^\[INFO\] 2016\/03\/07 22\:51/";     FUCK HERE!
$reg = str_replace("/","\/",substr($timestamp,0,16));
$reg = str_replace(":","\:",$reg);
$reg = "/^\[INFO\] ".$reg."/";

$fp=fopen($logfile ,"r");
if ($fp === false) {
	  print_log("Fail to open $logfile for r\n");
	  exit(1);
}

$res = array();
while(!feof($fp)){
	$line=trim(fgets($fp));
	if(empty($line)) continue;
	if(preg_match($reg,$line)){
		$content = explode("|",$line);
		$dname = $content[1];
		parse_str($content[2],$detail);
		$res[$dname] = $detail; 
	}
}
fclose($fp);

if (count($res) != $vm_num){
	  echo count($res)."\n";
	  echo $vm_num."\n";
	print_log("Seems vm num mismatch! Exit!");
	exit(1);
}

$standard = array(
	"cpu_ratio"=>90,
	"mem_ratio"=>95,
	"net_out"=>300,
	"net_in"=>300,
//	"cap"=>97,		// Didn't collect disk info, due to metrics spend a lot of time
//	"inode"=>97
);

foreach ($res as $k => $v){
	foreach ($standard as $key => $val){
		if ($v[$key] > $standard[$key]){
			if ($key == "net_out" || $key =="net_in"){
				$string = '{"groups":["cc"],"subject": "notify", "content":"M:$k $key usage $v[$key]MB/s more than ".$val."MB/s"}';
				$post_arr = array("content" => $string);
				http_request($alarm_url,5,1,$string);
				print_log("VM:$k $key usage $v[$key]MB/s more than ".$val."MB/s\n");
			}else {
				print_log("VM:$k $key usage $v[$key]% more than $val%\n");
			}
		}
	}
}

function print_log($content) {
	global $cur_time;
	global $cur_day;
	$content="[".$cur_time."] ".$content;
	print_r($content);
	file_put_contents("/home/monitor/logs/analysis_message.".$cur_day, $content, FILE_APPEND | LOCK_EX);
}

function http_request($url, $timeout = 5, $post_flag = 0, $post_data = array()) {
	  $ch = curl_init();
	  if (!$ch) return false;
	  curl_setopt($ch, CURLOPT_URL, $url);
	  curl_setopt($ch, CURLOPT_TIMEOUT, $timeout);
	  curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
	  if ($post_flag) {
			curl_setopt($ch, CURLOPT_POST, 1);
			curl_setopt($ch, CURLOPT_POSTFIELDS, $post_data);
	  }      
	  $data = @curl_exec($ch);
	  curl_close($ch);
	  return $data;  
}  

?>
