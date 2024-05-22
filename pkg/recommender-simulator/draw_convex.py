import json
import pprint
import sys
import matplotlib.pyplot as plt

dimension = "ap-ml-memory-hyperparam-d"
small = 0.05
mid = 0.01
large = 0.01
observe = "memory"

param = []
average_gap = []
overrun_seconds = []
request_adjust = []

with open('convex/param_sweep_result_{}.json'.format(dimension), 'r') as file:
    data = json.load(file)
    for d in data:
        if d['args'][dimension] > small:
            continue
        param.append(d['args'][dimension])
        average_gap.append(d['result']['{}-average-gap'.format(observe)])
        overrun_seconds.append(d['result']['{}-overrun-seconds'.format(observe)])
        request_adjust.append(d['result']['{}-request-adjust-times'.format(observe)])

fig, (ax1, ax2, ax3) = plt.subplots(3, figsize=(10, 7))

ax1.plot(param, average_gap, marker='o', markersize=3)
ax1.set_title('{} Average Gap'.format(observe))
ax1.set_xlabel(dimension + " value")
ax1.set_ylabel('mcores' if observe == 'cpu' else 'bytes')
ax1.grid(True)

ax2.plot(param, overrun_seconds, marker='o', markersize=3)
ax2.set_title('{} Overrun Seconds'.format(observe))
ax2.set_xlabel(dimension + " value")
ax2.set_ylabel('seconds')
ax2.grid(True)

ax3.plot(param, request_adjust, marker='o', markersize=3)
ax3.set_title('{} Request Adjust Times'.format(observe))
ax3.set_xlabel(dimension + " value")
ax3.set_ylabel('adjust times')
ax3.grid(True)


# 调整子图布局
plt.tight_layout()

dname = dimension.split('-')[-1]
fig.savefig('convex/convex_memory_{}_small{}.png'.format(dname, small))
# 显示图表
plt.show()