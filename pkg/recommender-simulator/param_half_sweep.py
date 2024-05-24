import numpy as np
import itertools
import subprocess
import json
import time
from concurrent.futures import ThreadPoolExecutor

const_cpu = True
const_memory = False

task_def = { # min, max, default, num
    "ap-ml-cpu-hyperparam-d":       [0.0, 1.0, 0.96, 150],
    "ap-ml-cpu-hyperparam-wdeltal": [0.0, 0.0, 0.0, 1],
    "ap-ml-cpu-hyperparam-wdeltam": [0.0, 0.0, 0.0, 1],
    "ap-ml-cpu-hyperparam-wo":      [0.0, 1.0, 0.04, 150],
    "ap-ml-cpu-hyperparam-wu":      [0.0, 0.0, 0.0, 1],
    "ap-ml-memory-hyperparam-d":        [0.0, 1.0, 0.1, 8],
    "ap-ml-memory-hyperparam-wdeltal":  [0.0, 1.0, 0.0, 8],
    "ap-ml-memory-hyperparam-wdeltam":  [0.0, 1.0, 0.1, 8],
    "ap-ml-memory-hyperparam-wo":       [0.0, 1.0, 0.2, 8],
    "ap-ml-memory-hyperparam-wu":       [0.0, 1.0, 0.0, 8]
}

max_workers = 32


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
    "-memory-limit-request-ratio=1.04",
    "-exit-memory-large-overrun=5000"
]

for k, v in task_def.items():
    if 'cpu' in k:
        init_args.append("-{}={}".format(k, v[2]))

file_index = 0
threadpool_results = []
output_args_list = []
with ThreadPoolExecutor(max_workers=max_workers) as executor:
    for md in np.linspace(0.0, 1.0, 20):
        for mwo in np.linspace(0.0, 1.0, 20):
            for mwu in np.linspace(0.01, 0.01, 1):
                for mwdeltaldiff in range(3, 10):
                    mwdeltal = max(0, mwu - 10**-mwdeltaldiff)

                    for mwdeltamdiff in range(mwdeltaldiff, 10):
                        mwdeltam = max(0, 10**-mwdeltaldiff - 10**-mwdeltamdiff)
                        print(md, mwo, mwu, mwdeltal, mwdeltam)
                        copied_args = init_args[:]
                        copied_args.append("-{}={}".format("ap-ml-memory-hyperparam-d", md))
                        copied_args.append("-{}={}".format("ap-ml-memory-hyperparam-wdeltal", mwdeltal))
                        copied_args.append("-{}={}".format("ap-ml-memory-hyperparam-wdeltam", mwdeltam))
                        copied_args.append("-{}={}".format("ap-ml-memory-hyperparam-wo", mwo))
                        copied_args.append("-{}={}".format("ap-ml-memory-hyperparam-wu", mwu))
                        output_args = {}
                        output_args["ap-ml-memory-hyperparam-d"] = md
                        output_args["ap-ml-memory-hyperparam-wdeltal"] = mwdeltal
                        output_args["ap-ml-memory-hyperparam-wdeltam"] = mwdeltam
                        output_args["ap-ml-memory-hyperparam-wo"] = mwo
                        output_args["ap-ml-memory-hyperparam-wu"] = mwu
                        
                        output_args_list.append(output_args)
                        copied_args.append("-metrics-summary-file=tmp/{}".format(file_index))

                        threadpool_results.append(executor.submit(subprocess.run, copied_args, check=True))
                        file_index += 1
                        # break

print("++++++++++Total tasks++++++++++: ", file_index)

for future in threadpool_results:
    if future.result().returncode != 0:
        print("Error return code in experiment!!!! check and rerun")
        exit(1)

minoutput = {}
mingap = 100000000000.0

iter_class = "memory"
max_overrun = {
    "cpu": 20000,
    "memory": 5000,
}

max_adjust = {
    "cpu": 1000,
    "memory": 100,
}

for i in range(len(output_args_list)):
    try:
        with open('metrics/tmp/{}_1.04.json'.format(i), 'r') as f:
            single_run_result = json.load(f)
            if single_run_result['{}-overrun-seconds'.format(iter_class)] <= max_overrun[iter_class]:
                    if single_run_result['{}-request-adjust-times'.format(iter_class)] <= max_adjust[iter_class]:
                        if single_run_result['{}-average-gap'.format(iter_class)] < mingap:
                            mingap = single_run_result['{}-average-gap'.format(iter_class)]
                            output = {}
                            output["args"] = output_args_list[i]
                            output["result"] = single_run_result
                            minoutput = output

    except FileNotFoundError:
        continue

with open('halfsweep/half_sweep_result_2.json', 'w') as f:
    json.dump(minoutput, f)
    