<?php
date_default_timezone_set('Asia/Shanghai'); 
@$_time=date("Y-m-d");
@$timestamp=date("Y/m/d H:i:s",(strtotime("now")-60));
echo $timestamp."\n";

exec('docker ps | awk "{print \$NF}" | fgrep -v NAMES | wc -l ',$output);
$vm_num=$output[0];

$logfile = "/opt/docker_vm_info_$_time";

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
	echo count($res)."\n";
	echo $vm_num."\n";
	print_r("vm num wrong!");
	exit(1);
}
#//2016/03/07 17:02:52 |desperate_cray|cpu_usg=0&cpu_user=0&cpu_sys=0&cpu_ratio=0&cpu_n=12&quota=1200&mem_rss=4.00&mem_quota=8796093022208.00&mem_cache=6.95&mem_mapped=2.47&mem_ratio=0.00&io_write=0.00&io_read=0.00&net_in=0.00&net_out=0.00&cap=1&inode=1	
foreach ($res as $k => $v){
	if ($v["cpu_ratio"] > 90 ){
		echo "$k cpu used lager than 90%!\n";
	}
	if ($v["mem_ratio"] > 95){
		echo "$k memory used lager than 95%\n";
	}
	if ($v["net_out"]>80){
		echo "$k net_out greater than 80MB/s\n";
	}
	if ($v["net_in"]>80){
		echo "$k net_out greater than 80MB/s\n";
	}
	if ($v["cap"]>97){
		echo "$k / disk capacity used more than 97%\n";
	}
	if ($v["inode"]>97){
		echo "$k / disk inode used more than 97%\n";
	}
}
?>
