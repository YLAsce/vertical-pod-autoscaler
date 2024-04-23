#!/bin/bash
component="workload-cpu-only"
# while true; do
#     sleep 100
# done

data_file="trace.data"

cgroup_prefix="/sys/fs/cgroup"
cgroup_file="/cpu.max"
cgroup_subpath=$(cat /proc/self/cgroup | sed 's/[^/]*//')
cgroup_fullpath="${cgroup_prefix}${cgroup_subpath}${cgroup_file}"
# cgroup_stresspath="${cgroup_prefix}${cgroup_subpath}/stress"

state_file="s3://mlintra/${component}.state"

cur_state=$(aws s3 cp ${state_file} -)

echo "======Read State: ${cur_state}"

data_file_lines=$(wc -l < "${data_file}")
start_line=$((cur_state % data_file_lines))

while true; do
    echo "start from ${start_line}"
    awk "NR > ${start_line}" ${data_file} | while IFS= read -r line; do
        start_time=$(date +%s)

        duration=$(echo "$line" | awk '{print $1}')
        cpu_usage=$(echo "$line" | awk '{print $2}')
        memory_usage=$(echo "$line" | awk '{print $3}')

        # # integer upper bound of cpu usage
        # cpu_upper_bound=$(echo "$cpu_usage" | awk -F'.' '{print $1}')
        # ((cpu_upper_bound++))
        # cpu_percentile=$(echo "scale=0; $cpu_usage * 100 / $cpu_upper_bound" | bc)

        # CPU percentile
        cpu_percentile=$(echo "$cpu_usage * 100000" | bc | cut -d. -f1)

        echo "===Current CPU Usage * 100000: ${cpu_percentile} | Mem Usage: ${memory_usage} | CPU file path: ${cgroup_fullpath}"

        echo ${cpu_percentile} > ${cgroup_fullpath}
        # stress-ng --cpu ${cpu_upper_bound} --cpu-load ${cpu_percentile} --cpu-method loop --vm 1 --vm-bytes ${memory_usage} --vm-keep --timeout ${duration}s
        # wait
        ./stress_worker ${duration} ${memory_usage}
        
        #ps aux | grep 'stress_worker' | grep -v 'grep' | awk '{print $2}' | while read -r pid; do cpulimit -p $pid -l ${cpu_percentile}; done
        wait

        end_time=$(date +%s)
        duration_actual=$((end_time - start_time))

        cur_state=$((cur_state + 1))
        echo ${cur_state} | aws s3 cp - ${state_file}

        echo "===FINISHED! | Time expected: ${duration} used: ${duration_actual} cur_state: ${cur_state}."
        echo
    done
    start_line=0
done