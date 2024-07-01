import numpy as np
import itertools
import subprocess
import json
import time
import random
import os
from concurrent.futures import ThreadPoolExecutor

iter_class = "memory"

task_def = { # min, max, default, num
    "ap-ml-cpu-hyperparam-d":           [0.0, 1.0, 0.8571428571428571, 101],
    "ap-ml-cpu-hyperparam-wdeltal":     [0.0, 1.0, 0.0, 101],
    "ap-ml-cpu-hyperparam-wdeltam":     [0.0, 1.0, 0.0, 101],
    "ap-ml-cpu-hyperparam-wo":          [0.0, 1.0, 0.14285714285714285,101],
    "ap-ml-cpu-hyperparam-wu":          [0.0, 1.0, 0.0, 101],
    "ap-ml-memory-hyperparam-d":        [0.0, 1.0, 0.7142857142857142, 101],
    "ap-ml-memory-hyperparam-wdeltal":  [0.0, 1.0, 0.14285714285714285, 101],
    "ap-ml-memory-hyperparam-wdeltam":  [0.0, 1.0, 0.7142857142857142, 101],
    "ap-ml-memory-hyperparam-wo":       [0.0, 1.0, 0.8571428571428571, 101],
    "ap-ml-memory-hyperparam-wu":       [0.0, 1.0, 0.0, 101]
}

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
"-exit-memory-large-overrun=10000"
]

max_overrun = {
    "cpu": 20000,
    "memory": 9840,
}

max_adjust = {
    "cpu": 1000,
    "memory": 100,
}

def print_args(args):
    for k, v in args.items():
        if iter_class in k:
            print(k, v)

def process_convex(convex_dimension):

    max_workers = 32

    ranges = []
    names = []
    total_tasks = 0
    for n, d in task_def.items() :
        r = []
        if n == convex_dimension:
            for v in np.linspace(d[0], d[1], d[3]):
                r.append(v)
        else:
            r.append(d[2])
        ranges.append(r)
        names.append(n)
    print("Ranges:", ranges)

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
            copied_args.append("-metrics-summary-file=tmp/{}-{}".format(file_index, convex_dimension))

            threadpool_results.append(executor.submit(subprocess.run, copied_args, check=True))
            file_index += 1
            # break

    print("++++++++++Total tasks++++++++++: ", total_tasks)

    for future in threadpool_results:
        if future.result().returncode != 0:
            print("Error return code in experiment!!!! check and rerun")
            exit(1)

    minoutput = {}
    mingap = 100000000000.0
    for i in range(len(output_args_list)):
        try:
            with open('metrics/tmp/{}-{}_1.04.json'.format(i, convex_dimension), 'r') as f:
                single_run_result = json.load(f)
                if single_run_result['{}-overrun-seconds'.format(iter_class)] <= max_overrun[iter_class]:
                    if single_run_result['{}-request-adjust-times'.format(iter_class)] <= max_adjust[iter_class]:
                        if single_run_result['{}-average-gap'.format(iter_class)] < mingap:
                            mingap = single_run_result['{}-average-gap'.format(iter_class)]
                            output = {}
                            output["args"] = output_args_list[i]
                            output["result"] = single_run_result
                            minoutput = output
            
            os.remove('metrics/tmp/{}-{}_1.04.json'.format(i, convex_dimension))
        except FileNotFoundError:
            continue
    if len(minoutput) > 0:
        for k, v in minoutput['args'].items():
            task_def[k][2] = v

    return mingap, minoutput

select_keys = [key for key in task_def.keys() if iter_class in key]

def do_descent(round):
    print("Cur Randompoint task def:")
    print_args(task_def)
    mingap = 100000000000.0
    minoutput = {}
    prevmingap = 100000000000.0
    i = 0
    while(1):
        random.shuffle(select_keys)
        for k in select_keys:
            print("============Start descent round", i, k)
            mingap, minoutput = process_convex(k)
            print("============Finish descent round, mingap=", mingap)
            print("minoutput=")
            if len(minoutput) > 0:
                print_args(minoutput['args'])
            else:
                print("NO OUTPUT")
            i += 1
        if prevmingap - mingap < 1.0:
            with open('find/cur_best_{}_{}.json'.format(iter_class, round), 'w') as f:
                json.dump(minoutput, f, indent=4)
            return minoutput, mingap
        prevmingap = mingap

r = 0
gmingap = 100000000000.0
gminoutput = {}
while(1):
    for k in select_keys:
        task_def[k][2] = random.uniform(0,1)

    minoutput, mingap = do_descent(r)
    if mingap < gmingap:
        gminoutput = minoutput
        gmingap = mingap

    r += 1
    with open('find/best_{}.json'.format(iter_class), 'w') as f:
        json.dump(gminoutput, f, indent=4)
    