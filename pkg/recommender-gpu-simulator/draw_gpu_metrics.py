import os
import matplotlib.pyplot as plt
import pandas as pd

folder_path = 'metrics/gpu-ml'

nodenum = 0
firsttime = 0
with open(folder_path + '/nodepool.info', 'r') as file:
    line = file.readline().strip()
    nodenum = int(line)

with open(folder_path + '/metricsinfo.info', 'r') as file:
    line = file.readline().strip()
    firsttime = int(line)


fig, axes = plt.subplots(nodenum, figsize=(12, nodenum*3))

assigned = [False] * len(axes) * 4
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
            merged['requestid'] = merged['requestid_x'] + merged['requestid_y']   
            merged['sm_overrun'] = merged['sm_overrun_y']
            merged['mem_overrun'] = merged['mem_overrun_y']
            merged['nodeid'] = merged['nodeid_y']
            base_data[nodeid] = merged
            
            axes[nodenum-nodeid-1].plot(base_data[nodeid]['ts'], base_data[nodeid]['requestid'], color='red', label=parse_label('Daemon Workload', nodenum-nodeid-1), linewidth=0.3)
            axes[nodenum-nodeid-1].fill_between(base_data[nodeid]['ts'], base_data[nodeid]['requestid_x'], base_data[nodeid]['requestid'], color='red', alpha=0.3)

            marked_points = base_data[nodeid][base_data[nodeid]['sm_overrun_y'] == 1]
            axes[nodenum-nodeid-1].scatter(marked_points['ts'], marked_points['requestid'], color='blue', marker='o', s=10, label=parse_label('SM Overrun', 2*nodenum-nodeid-1))

            marked_points2 = base_data[nodeid][base_data[nodeid]['mem_overrun_y'] == 1]
            axes[nodenum-nodeid-1].scatter(marked_points2['ts'], marked_points2['requestid'], color='orange', marker='o', s=10, label=parse_label('Memory Overrun', 3*nodenum-nodeid-1))

            base_data[nodeid].drop(columns=['requestid_x', 'requestid_y'], inplace=True)
            base_data[nodeid].drop(columns=['sm_overrun_x', 'sm_overrun_y'], inplace=True)
            base_data[nodeid].drop(columns=['mem_overrun_x', 'mem_overrun_y'], inplace=True)
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
            merged['requestid'] = merged['requestid_x'] + merged['requestid_y']   
            merged['sm_overrun'] = merged['sm_overrun_y']
            merged['mem_overrun'] = merged['mem_overrun_y']
            merged['nodeid'] = merged['nodeid_y']
            base_data[nodeid] = merged
            
            axes[nodenum-nodeid-1].plot(base_data[nodeid]['ts'], base_data[nodeid]['requestid'], color='green', label=parse_label('Serverless Workload', 4*nodenum-nodeid-1), linewidth=0.3)
            axes[nodenum-nodeid-1].fill_between(base_data[nodeid]['ts'], base_data[nodeid]['requestid_x'], base_data[nodeid]['requestid'], color='green', alpha=0.3)

            marked_points = base_data[nodeid][base_data[nodeid]['sm_overrun_y'] == 1]
            axes[nodenum-nodeid-1].scatter(marked_points['ts'], marked_points['requestid'], color='blue', marker='o', s=10, label=parse_label('SM Overrun', 2*nodenum-nodeid-1))

            marked_points2 = base_data[nodeid][base_data[nodeid]['mem_overrun_y'] == 1]
            axes[nodenum-nodeid-1].scatter(marked_points2['ts'], marked_points2['requestid'], color='orange', marker='o', s=10, label=parse_label('Memory Overrun', 3*nodenum-nodeid-1))

            base_data[nodeid].drop(columns=['requestid_x', 'requestid_y'], inplace=True)
            base_data[nodeid].drop(columns=['sm_overrun_x', 'sm_overrun_y'], inplace=True)
            base_data[nodeid].drop(columns=['mem_overrun_x', 'mem_overrun_y'], inplace=True)
            base_data[nodeid].drop(columns=['nodeid_x', 'nodeid_y'], inplace=True)

length = len(base_data[0])

y_line1 = pd.Series([1]*length)
for i in range(nodenum):
    axes[i].plot(base_data[0]['ts'], y_line1, label='Node size')

for i in range(len(axes)):
    axes[i].set_xlabel('Time(Seconds)')
    axes[i].set_ylabel('GPU Usage(GIs)')

    axes[i].legend()
# ax1.legend()
# ax2.legend()

# 调整子图布局
plt.tight_layout()

fig.savefig('{}/graph.png'.format(folder_path))

# 显示图表
plt.show()