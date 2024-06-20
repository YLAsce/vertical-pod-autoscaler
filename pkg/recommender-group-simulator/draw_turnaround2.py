import pandas as pd
import os

folder_path = 'metrics/group-ml'
target_path = 'metrics/turnaround-ml'

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

print('Loaded CPU size %f, Mem size %f', nodesize_cpu, nodesize_mem)



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


assigned_cpu = []
assigned_mem = []
timestamp = []

cur_state = [0]*len(datas)
cur_timestamp = 0



while(1):
    cur_assigned_cpu = cur_assigned_mem = 0
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

    assigned_cpu.append(cur_assigned_cpu)
    assigned_mem.append(cur_assigned_mem)
    timestamp.append(cur_timestamp)
    cur_timestamp += 1
    if cur_timestamp % 10 == 0:
        print(cur_timestamp)

cpuline = [nodesize_cpu]*len(timestamp)
memline = [nodesize_mem]*len(timestamp)

import matplotlib.pyplot as plt

fig, (ax1, ax2) = plt.subplots(2, figsize=(12, 5))

ax1.plot(timestamp, assigned_cpu, label='CPU Workload', color='gray')
ax2.plot(timestamp, assigned_mem, label='Memory Workload', color='gray')

ax1.plot(timestamp, cpuline, label='Node Size', color='blue')
ax2.plot(timestamp, memline, label='Node Size', color='blue')

ax1.fill_between(timestamp, assigned_cpu, 0, color='gray', alpha=0.5)
ax2.fill_between(timestamp, assigned_mem, 0, color='gray', alpha=0.5)


labeled = False
for k, v in pulse_cpu:
    ax1.plot([k, k], [0, v], color='green', linestyle='-', linewidth=1, label='Start time impulse' if not labeled else '')
    labeled = True

labeled = False
for k, v in pulse_mem:
    ax2.plot([k, k], [0, v], color='green', linestyle='-', linewidth=1, label='Start time impulse' if not labeled else '')
    labeled = True


ax1.legend()
ax1.set_xlabel('CPU Workload(Cores)')
ax1.set_ylabel('Timestamp(Seconds)')
fig.savefig('{}/graph_turnaround.png'.format(folder_path))

plt.show()