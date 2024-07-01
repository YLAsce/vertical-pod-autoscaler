import pandas as pd
import os
import sys

# folder_path = 'metrics/group-ml'
# target_path = 'metrics/turnaround-ml'


folder_path = 'metrics/' + sys.argv[1]
target_path = 'metrics/' + sys.argv[2]

nodesize_cpu = nodesize_mem = 0.0
firsttime = 0
with open(folder_path + '/nodepool.info', 'r') as file:
    line = file.readline().strip()
    parts = line.split()
    nodesize_cpu = float(parts[0])
    nodesize_mem = float(parts[1])

with open(folder_path + '/metricsinfo.info', 'r') as file:
    line = file.readline().strip()
    firsttime = int(line)

print('Loaded CPU size {}, Mem size {}'.format(nodesize_cpu, nodesize_mem))



datas = []
def add_data(type):
    for file_name in os.listdir(folder_path):
        if file_name.startswith(type):
            file_path = os.path.join(folder_path, file_name)
            data = pd.read_csv(file_path)
            data['ts'] = data['ts'] - firsttime
            datas.append(data)
add_data('daemon')
print("Loaded Daemon")
add_data('serverless')
print("Loaded Serverless")

pulse_cpu = {}
pulse_mem = {}
for d in datas:
    if d.at[0, 'ts'] not in pulse_cpu:
        pulse_cpu[d.at[0, 'ts']] = 0.0
        pulse_mem[d.at[0, 'ts']] = 0.0
    pulse_cpu[d.at[0, 'ts']] += d.at[0, 'cpu_request']
    pulse_mem[d.at[0, 'ts']] += d.at[0, 'mem_request']

print(pulse_cpu, pulse_mem)

with open(target_path + '/pulse_cpu.info', 'w') as file:
    for k,v in pulse_cpu.items():
        file.write('{} {}\n'.format(k, v))

with open(target_path + '/pulse_mem.info', 'w') as file:
    for k,v in pulse_mem.items():
        file.write('{} {}\n'.format(k, v))

cur_state = [0]*len(datas)
cur_timestamp = 0


with open(target_path + '/workload_cpu.info', 'w') as filecpu, open(target_path + '/workload_mem.info', 'w') as filemem:
    while(1):
        cur_assigned_cpu = cur_assigned_mem = 0.0
        remaining = False
        for i in range(len(datas)):
            if cur_state[i] < len(datas[i]):
                remaining = True
                if cur_timestamp >= datas[i].at[cur_state[i], 'ts']:
                    if cur_assigned_cpu + datas[i].at[cur_state[i], 'cpu_request'] <= nodesize_cpu and cur_assigned_mem + datas[i].at[cur_state[i], 'mem_request'] <= nodesize_mem:
                        cur_assigned_cpu += datas[i].at[cur_state[i], 'cpu_request']
                        cur_assigned_mem += datas[i].at[cur_state[i], 'mem_request']
                        cur_state[i] += 1
        if not remaining:
            break
        filecpu.write('{} {}\n'.format(cur_timestamp, cur_assigned_cpu))
        filemem.write('{} {}\n'.format(cur_timestamp, cur_assigned_mem))

        cur_timestamp += 1
        if cur_timestamp % 1000 == 0:
            print(cur_timestamp)

print("Done")