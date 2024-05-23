import json
import pprint
import sys
import matplotlib.pyplot as plt

filename = sys.argv[1]

timestamps = []

cpu_request = []
cpu_usage = []

# memory_limit = []
memory_request = []
memory_usage = []

cpu_request2 = []
memory_request2 = []

with open('metrics/{}_1.04.data'.format(filename), 'r') as file:
    i = 0
    for line in file:
        i += 1
        parts = line.split()
        # if (i > 3000):
        #     break
        timestamps.append(int(parts[0]))
        cpu_request.append( float(parts[1]) )
        cpu_usage.append( float(parts[2]) )

        memory_request.append( int(parts[3]) )
        # memory_limit.append( int(float(parts[3])*1.04) )
        memory_usage.append( int(parts[4]) )
        # print(int(parts[4]))

with open('metrics/metrics-rule_1.04.data', 'r') as file:
    i = 0
    for line in file:
        i += 1
        parts = line.split()
        # if (i > 3000):
        #     break
        # timestamps.append(int(parts[0]))
        cpu_request2.append( float(parts[1]) )
        # cpu_usage.append( float(parts[2]) )

        memory_request2.append( int(parts[3]) )
        # memory_limit.append( int(float(parts[3])*1.04) )
        # memory_usage.append( int(parts[4]) )
        # print(int(parts[4]))

basetime = timestamps[0]
for i in range(len(timestamps)):
    timestamps[i] -= basetime

for _ in range(len(timestamps) - len(cpu_request2)):
    cpu_request2.append(0.0)
    memory_request2.append(0.0)

fig, (ax1, ax2) = plt.subplots(2, figsize=(14, 5))

# 在第一个子图上绘制 CPU 曲线
ax1.plot(timestamps, cpu_request, label='CPU Request ML')
ax1.plot(timestamps, cpu_usage, label='CPU Usage')
ax1.plot(timestamps, cpu_request2, label='CPU Request Rule')
ax1.set_title('CPU Resource')
ax1.set_xlabel('Time (minutes)')
ax1.set_ylabel('Resource')
ax1.grid(True)
# ax1.hlines(0.7, 0, timestamps[-1], colors='black', linestyles='dashed')
# ax1.hlines(0.1, 0, timestamps[-1], colors='black', linestyles='dashed')
ax1.legend()


# 在第二个子图上绘制内存曲线
ax2.plot(timestamps, memory_request2, label='Memory Request Rule')
ax2.plot(timestamps, memory_request, label='Memory Request ML')
ax2.plot(timestamps, memory_usage, label='Memory Usage')
ax2.set_title('Memory Resource')
ax2.set_xlabel('Time (minutes)')
ax2.set_ylabel('Resource')
ax2.grid(True)
# ax2.hlines(367001600, 0, timestamps[-1], colors='black', linestyles='dashed')
# ax2.hlines(52428800, 0, timestamps[-1], colors='black', linestyles='dashed')
ax2.legend()

# 调整子图布局
plt.tight_layout()

# 显示图表
plt.show()