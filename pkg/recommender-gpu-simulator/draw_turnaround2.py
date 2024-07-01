import pandas as pd
import os

folder_path = 'metrics/group'
target_path = 'metrics/turnaround'
target_path_ml = 'metrics/turnaround-ml'
target_path_const = 'metrics/turnaround-const'

nodesize_cpu = nodesize_mem = 0.0

with open(folder_path + '/nodepool.info', 'r') as file:
    line = file.readline().strip()
    parts = line.split()
    nodesize_cpu = float(parts[0])
    nodesize_mem = float(parts[1])

print('Loaded CPU size %f, Mem size %f', nodesize_cpu, nodesize_mem)


pulse_cpu = {}
pulse_mem = {}
with open(target_path + '/pulse_cpu.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        pulse_cpu[int(spl[0])] = float(spl[1])

with open(target_path + '/pulse_mem.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        pulse_mem[int(spl[0])] = float(spl[1])

print(pulse_cpu, pulse_mem)


assigned_cpu = []
assigned_mem = []

ml_assigned_cpu = []
ml_assigned_mem = []

const_assigned_cpu = []
const_assigned_mem = []

with open(target_path + '/workload_cpu.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        assigned_cpu.append(float(spl[1]))

with open(target_path + '/workload_mem.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        assigned_mem.append(float(spl[1]))

with open(target_path_ml + '/workload_cpu.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        ml_assigned_cpu.append(float(spl[1]))

with open(target_path_ml + '/workload_mem.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        ml_assigned_mem.append(float(spl[1]))

with open(target_path_const + '/workload_cpu.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        const_assigned_cpu.append(float(spl[1]))

with open(target_path_const + '/workload_mem.info', 'r') as file:
    lines = file.readlines()
    for l in lines:
        spl = l.strip().split()
        const_assigned_mem.append(float(spl[1]))

print(len(assigned_cpu), len(assigned_mem), len(ml_assigned_cpu), len(ml_assigned_mem), len(const_assigned_cpu), len(const_assigned_mem))

ml_assigned_cpu += [0.0]* (len(const_assigned_cpu) - len(ml_assigned_cpu))
ml_assigned_mem += [0.0]* (len(const_assigned_mem) - len(ml_assigned_mem))

assigned_cpu += [0.0]* (len(const_assigned_cpu) - len(assigned_cpu))
assigned_mem += [0.0]* (len(const_assigned_mem) - len(assigned_mem))

ts = [i for i in range(len(ml_assigned_cpu))]
print(len(assigned_cpu), len(assigned_mem), len(ml_assigned_cpu), len(ml_assigned_mem), len(const_assigned_cpu), len(const_assigned_mem), len(ts))

cpusize = [nodesize_cpu]*len(ml_assigned_cpu)
memsize = [nodesize_mem]*len(ml_assigned_mem)

import matplotlib.pyplot as plt

fig, (ax1, ax2, ax3, ax4, ax5, ax6) = plt.subplots(6, figsize=(12, 18))

ax1.plot(ts, const_assigned_cpu, label='CPU Workload Fixed-size', color='gray', linewidth=1)
ax2.plot(ts, const_assigned_mem, label='Memory Workload Fixed-size', color='gray', linewidth=1)

ax1.plot(ts, cpusize, label='Node Size', color='blue')
ax2.plot(ts, memsize, label='Node Size', color='blue')

ax1.fill_between(ts, const_assigned_cpu, 0, color='gray', alpha=0.5)
ax2.fill_between(ts, const_assigned_mem, 0, color='gray', alpha=0.5)

ax3.plot(ts, assigned_cpu, label='CPU Workload Rule-Based', color='gray', linewidth=1)
ax4.plot(ts, assigned_mem, label='Memory Workload Rule-Based', color='gray', linewidth=1)

ax3.plot(ts, cpusize, label='Node Size', color='blue')
ax4.plot(ts, memsize, label='Node Size', color='blue')

ax3.fill_between(ts, assigned_cpu, 0, color='gray', alpha=0.5)
ax4.fill_between(ts, assigned_mem, 0, color='gray', alpha=0.5)

ax5.plot(ts, ml_assigned_cpu, label='CPU Workload Autopilot', color='gray', linewidth=1)
ax6.plot(ts, ml_assigned_mem, label='Memory Workload Autopilot', color='gray', linewidth=1)

ax5.plot(ts, cpusize, label='Node Size', color='blue')
ax6.plot(ts, memsize, label='Node Size', color='blue')

ax5.fill_between(ts, ml_assigned_cpu, 0, color='gray', alpha=0.5)
ax6.fill_between(ts, ml_assigned_mem, 0, color='gray', alpha=0.5)


labeled = False
for k, v in pulse_cpu.items():
    ax3.plot([k, k], [0, v], color='green', linestyle='-', linewidth=1, label='Start time impulse' if not labeled else '')
    labeled = True

labeled = False
for k, v in pulse_mem.items():
    ax4.plot([k, k], [0, v], color='green', linestyle='-', linewidth=1, label='Start time impulse' if not labeled else '')
    labeled = True

labeled = False
for k, v in pulse_cpu.items():
    ax5.plot([k, k], [0, v], color='green', linestyle='-', linewidth=1, label='Start time impulse' if not labeled else '')
    labeled = True

labeled = False
for k, v in pulse_mem.items():
    ax6.plot([k, k], [0, v], color='green', linestyle='-', linewidth=1, label='Start time impulse' if not labeled else '')
    labeled = True

ax1.set_ylabel('CPU Workload(Cores)')

ax2.set_ylabel('Mem Workload(Bytes)')

ax3.set_ylabel('CPU Workload(Cores)')

ax4.set_ylabel('Mem Workload(Bytes)')

ax5.set_ylabel('CPU Workload(Cores)')

ax6.set_ylabel('Mem Workload(Bytes)')
ax6.set_xlabel('Timestamp(Seconds)')


ax1.legend()
ax2.legend()
ax3.legend()
ax4.legend()
ax5.legend()
ax6.legend()

fig.savefig('{}/graph_turnaround_full.png'.format(target_path))


plt.show()