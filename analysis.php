<?php
date_default_timezone_set('Asia/Shanghai'); 
@$cur_day=date("Y-m-d");
@$cur_time=date("H:i:s");
@$timestamp=date("Y/m/d H:i:s",(strtotime("now")-60));

exec('docker ps | awk "{print \$NF}" | fgrep -v NAMES | wc -l ',$output);
$vm_num=$output[0];

$logfile = "/opt/docker_vm_info_$cur_day";

#$reg= "/^\[INFO\] 2016\/03\/07 22\:51/";     FUCK HERE!
$reg = str_replace("/","\/",substr($timestamp,0,16));
$reg = str_replace(":","\:",$reg);
$reg = "/^\[INFO\] ".$reg."/";

$fp=fopen($logfile ,"r");
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
	print_log("Seems vm num mismatch! Exit!");
	exit(1);
}

$standard = array(
	"cpu_ratio"=>90,
	"mem_ratio"=>95,
	"net_out"=>300,
	"net_in"=>300,
	"cap"=>97,
	"inode"=>97
);

foreach ($res as $k => $v){
	foreach ($standard as $key => $val){
		if ($v[$key] > $standard[$key]){
			if ($key == "net_out" || $key =="net_in"){
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

#//2016/03/07 17:02:52 |desperate_cray|cpu_usg=0&cpu_user=0&cpu_sys=0&cpu_ratio=0&cpu_n=12&quota=1200&mem_rss=4.00&mem_quota=8796093022208.00&mem_cache=6.95&mem_mapped=2.47&mem_ratio=0.00&io_write=0.00&io_read=0.00&net_in=0.00&net_out=0.00&cap=1&inode=1	
?>
