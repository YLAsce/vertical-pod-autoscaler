import json
import pprint
import matplotlib.pyplot as plt

frequency = 1 # 1min
data_points = 240

with open('data/metrics_{}_{}.json'.format(frequency, data_points), 'r') as file:
    data = json.load(file)


# pprint.pprint(data)

timestamps = [i*frequency for i in range(data_points)]
pod_name = ''

cpu_limit = []
cpu_request = []
cpu_usage = []

memory_limit = []
memory_request = []
memory_usage = []

for d in data:
    if pod_name not in d:
        pod_name = list(d.keys())[0]
    print(pod_name)
    cpu_limit.append( list( d[pod_name] .values())[0]     ['cpu']['limit'] )
    cpu_request.append( list( d[pod_name] .values())[0]     ['cpu']['request'] )
    cpu_usage.append( list( d[pod_name] .values())[0]     ['cpu']['usage'] )

    memory_limit.append( list( d[pod_name] .values())[0]     ['memory']['limit'] )
    memory_request.append( list( d[pod_name] .values())[0]     ['memory']['request'] )
    memory_usage.append( list( d[pod_name] .values())[0]     ['memory']['usage'] )

fig, (ax1, ax2) = plt.subplots(2)

# 在第一个子图上绘制 CPU 曲线
ax1.plot(timestamps, cpu_limit, label='CPU Limit')
ax1.plot(timestamps, cpu_request, label='CPU Request')
ax1.plot(timestamps, cpu_usage, label='CPU Usage')
ax1.set_title('CPU Resource')
ax1.set_xlabel('Time (minutes)')
ax1.set_ylabel('Resource')
ax1.grid(True)
# ax1.hlines(0.7, 0, timestamps[-1], colors='black', linestyles='dashed')
# ax1.hlines(0.1, 0, timestamps[-1], colors='black', linestyles='dashed')
ax1.legend()


# 在第二个子图上绘制内存曲线
ax2.plot(timestamps, memory_limit, label='Memory Limit')
ax2.plot(timestamps, memory_request, label='Memory Request')
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