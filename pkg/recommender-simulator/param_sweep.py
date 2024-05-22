import numpy as np
import itertools
import subprocess
import json
import time
from concurrent.futures import ThreadPoolExecutor

const_cpu = False
const_memory = True

task_def = { # min, max, default, num
    "ap-ml-cpu-hyperparam-d":       [0.0, 1.0, 0.96, 8],
    "ap-ml-cpu-hyperparam-wdeltal": [0.0, 1.0, 0.0, 8],
    "ap-ml-cpu-hyperparam-wdeltam": [0.0, 1.0, 0.0, 8],
    "ap-ml-cpu-hyperparam-wo":      [0.0, 1.0, 0.04, 8],
    "ap-ml-cpu-hyperparam-wu":      [0.0, 1.0, 0.0, 8],
    "ap-ml-memory-hyperparam-d":        [0.0, 1.0, 0.02, 11],
    "ap-ml-memory-hyperparam-wdeltal":  [0.0, 1.0, 0.0000000000099999, 11],
    "ap-ml-memory-hyperparam-wdeltam":  [0.0, 1.0, 0.00000000000000005, 11],
    "ap-ml-memory-hyperparam-wo":       [0.0, 1.0, 0.5, 11],
    "ap-ml-memory-hyperparam-wu":       [0.0, 0.0, 0.00000000001, 1]
}

max_workers = 32

ranges = []
names = []
total_tasks = 0
for n, d in task_def.items() :
    r = []
    if const_cpu and n.startswith("ap-ml-cpu"):
        r.append(d[2])
    elif const_memory and n.startswith("ap-ml-memory"):
        r.append(d[2])
    else:
        print(d[0], d[1], d[3])
        for v in np.linspace(d[0], d[1], d[3]):
            r.append(v)
    print(r)
    ranges.append(r)
    names.append(n)


init_args = [
    "bin/recommender-simulator",
    "-ap-algorithm-ml=true",
    "-ap-cpu-histogram-bucket-num=400",
    "-ap-cpu-histogram-decay-half-life=12h",
    "-ap-cpu-histogram-max=1.9",
    "-ap-cpu-histogram-n=5",
    "-ap-cpu-recommend-policy=sp_95",
    "-ap-fluctuation-reducer-duration=1h",
    "-ap-memory-histogram-bucket-num=400",
    "-ap-memory-histogram-decay-half-life=48h",
    "-ap-memory-histogram-max=3483278000",
    "-ap-memory-histogram-n=5",
    "-ap-memory-recommend-policy=sp_98",
    "-ap-ml-cpu-size-buckets-mm=1",
    "-ap-ml-cpu-num-dm=50",
    "-ap-ml-cpu-num-mm=400",
    "-ap-ml-memory-size-buckets-mm=1",
    "-ap-ml-memory-num-dm=50",
    "-ap-ml-memory-num-mm=500",
    "-initial-cpu=0.6",
    "-initial-memory=600000000",
    "-metrics-file=",
    "-oom-bump-up-ratio=1.2",
    "-oom-min-bump-up-bytes=100000000",
    "-recommender-interval=5m",
    "-trace-file=trace",
    "-metrics-summary-ignore-head=1800",
    "-memory-limit-request-ratio=1.04"
]

file_index = 0
threadpool_results = []
output_args_list = []
with ThreadPoolExecutor(max_workers=max_workers) as executor:
    for c in itertools.product(*ranges) :
        total_tasks += 1
        copied_args = init_args[:]

        output_args = {}
        for i in range(len(c)):
            copied_args.append("-{}={}".format(names[i], c[i]))
            output_args[names[i]] = c[i]

        output_args_list.append(output_args)
        copied_args.append("-metrics-summary-file=tmp/{}".format(file_index))

        print("Add Task: ARG:", c)
        threadpool_results.append(executor.submit(subprocess.run, copied_args, check=True))
        file_index += 1
        # break

print("++++++++++Total tasks++++++++++: ", total_tasks)

for future in threadpool_results:
    if future.result().returncode != 0:
        print("Error return code in experiment!!!! check and rerun")
        exit(1)

outputs = []

for i in range(len(output_args_list)):
    try:
        with open('metrics/tmp/{}_1.04.json'.format(i), 'r') as f:
            single_run_result = json.load(f)
            output = {}
            output["args"] = output_args_list[i]
            output["result"] = single_run_result
            outputs.append(output)
    except FileNotFoundError:
        continue

with open('param_sweep_result.json', 'w') as f:
    json.dump(outputs, f)
    