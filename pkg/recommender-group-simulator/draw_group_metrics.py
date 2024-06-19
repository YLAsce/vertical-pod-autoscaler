import os
import matplotlib.pyplot as plt
import pandas as pd

folder_path = 'metrics/group-ml'

nodesize_cpu = nodesize_mem = 0.0
nodenum = 0
firsttime = 0
with open(folder_path + '/nodepool.info', 'r') as file:
    line = file.readline().strip()
    parts = line.split()
    nodesize_cpu = float(parts[0])
    nodesize_mem = float(parts[1])
    nodenum = int(parts[2])

with open(folder_path + '/metricsinfo.info', 'r') as file:
    line = file.readline().strip()
    firsttime = int(line)


fig, axes = plt.subplots(2*nodenum, figsize=(12, nodenum*5))

assigned = [False] * len(axes) * 3
def parse_label(str, i):
    if assigned[i] == False:
        assigned[i] = True
        return str
    
    return ''

base_data = []
for i in range(nodenum):
    base_data.append(pd.read_csv(os.path.join(folder_path, "metricsinfo.csv")))

for d in base_data:
    d['ts'] = d['ts'] - firsttime

for file_name in os.listdir(folder_path):
    if file_name.startswith("daemon"):
        file_path = os.path.join(folder_path, file_name)
        data = pd.read_csv(file_path)
        data['ts'] = data['ts'] - firsttime
        for nodeid in range(nodenum):
            filtered_data = data[data['nodeid'] == nodeid]
            if len(filtered_data) == 0:
                continue
            merged = pd.merge(base_data[nodeid], filtered_data, on='ts', how='outer').fillna(0.0)
            merged['cpu_request'] = merged['cpu_request_x'] + merged['cpu_request_y']   
            merged['mem_request'] = merged['mem_request_x'] + merged['mem_request_y']
            merged['oom'] = merged['oom_y']
            merged['nodeid'] = merged['nodeid_y']
            base_data[nodeid] = merged
            
            axes[nodenum-nodeid-1].plot(base_data[nodeid]['ts'], base_data[nodeid]['cpu_request'], color='red', label=parse_label('Daemon Workload', nodenum-nodeid-1), linewidth=0.3)
            axes[nodenum-nodeid-1].fill_between(base_data[nodeid]['ts'], base_data[nodeid]['cpu_request_x'], base_data[nodeid]['cpu_request'], color='red', alpha=0.3)
            axes[2*nodenum-nodeid-1].plot(base_data[nodeid]['ts'], base_data[nodeid]['mem_request'], color='red', label=parse_label('Daemon Workload', 2*nodenum-nodeid-1), linewidth=0.3)
            axes[2*nodenum-nodeid-1].fill_between(base_data[nodeid]['ts'], base_data[nodeid]['mem_request_x'], base_data[nodeid]['mem_request'], color='red', alpha=0.3)

            marked_points = base_data[nodeid][base_data[nodeid]['oom_y'] == 1]
            axes[2*nodenum-nodeid-1].scatter(marked_points['ts'], marked_points['mem_request'], color='blue', marker='o', s=5, label=parse_label('OOM Event', 6*nodenum-nodeid-1))

            base_data[nodeid].drop(columns=['cpu_request_x', 'cpu_request_y'], inplace=True)
            base_data[nodeid].drop(columns=['mem_request_x', 'mem_request_y'], inplace=True)
            base_data[nodeid].drop(columns=['oom_x', 'oom_y'], inplace=True)
            base_data[nodeid].drop(columns=['nodeid_x', 'nodeid_y'], inplace=True)

for file_name in os.listdir(folder_path):
    if file_name.startswith("serverless"):
        file_path = os.path.join(folder_path, file_name)
        data = pd.read_csv(file_path)
        data['ts'] = data['ts'] - firsttime
        for nodeid in range(nodenum):
            filtered_data = data[data['nodeid'] == nodeid]
            if len(filtered_data) == 0:
                continue
            merged = pd.merge(base_data[nodeid], filtered_data, on='ts', how='outer').fillna(0.0)
            merged['cpu_request'] = merged['cpu_request_x'] + merged['cpu_request_y']   
            merged['mem_request'] = merged['mem_request_x'] + merged['mem_request_y']
            merged['oom'] = merged['oom_y']
            merged['nodeid'] = merged['nodeid_y']

            base_data[nodeid] = merged
            
            axes[nodenum-nodeid-1].plot(merged['ts'], merged['cpu_request'], color='green', label=parse_label('Serverless Workload', 3*nodenum-nodeid-1), linewidth=0.3)
            axes[nodenum-nodeid-1].fill_between(merged['ts'], merged['cpu_request_x'], merged['cpu_request'], color='green', alpha=0.3)
            axes[2*nodenum-nodeid-1].plot(merged['ts'], merged['mem_request'], color='green', label=parse_label('Serverless Workload', 4*nodenum-nodeid-1), linewidth=0.3)
            axes[2*nodenum-nodeid-1].fill_between(merged['ts'], merged['mem_request_x'], merged['mem_request'], color='green', alpha=0.3)

            marked_points = merged[merged['oom_y'] == 1]
            axes[2*nodenum-nodeid-1].scatter(marked_points['ts'], marked_points['mem_request'], color='blue', marker='o', s=5, label=parse_label('OOM Event', 6*nodenum-nodeid-1))

            base_data[nodeid].drop(columns=['cpu_request_x', 'cpu_request_y'], inplace=True)
            base_data[nodeid].drop(columns=['mem_request_x', 'mem_request_y'], inplace=True)
            base_data[nodeid].drop(columns=['oom_x', 'oom_y'], inplace=True)
            base_data[nodeid].drop(columns=['nodeid_x', 'nodeid_y'], inplace=True)

length = len(base_data[0])

y_line1 = pd.Series([nodesize_cpu]*length)
y_line2 = pd.Series([nodesize_mem]*length)
for i in range(nodenum):
    axes[i].plot(base_data[0]['ts'], y_line1, label='Node size')
    axes[nodenum+i].plot(base_data[0]['ts'], y_line2, label='Node size')

for i in range(len(axes)):
    axes[i].set_xlabel('Time(Seconds)')
    if i < len(axes)/2:
        axes[i].set_ylabel('CPU Usage(Cores)')
    else:
        axes[i].set_ylabel('Memory Usage(Bytes)')
    axes[i].legend()
# ax1.legend()
# ax2.legend()

# 调整子图布局
plt.tight_layout()

fig.savefig('{}/graph.png'.format(folder_path))

# 显示图表
plt.show()