import json
import pprint
import sys
import matplotlib.pyplot as plt

dimension = "ap-ml-cpu-hyperparam-d"
observe = "cpu"

param = []
average_gap = []
overrun_seconds = []
request_adjust = []

with open('convex/param_sweep_result_{}.json'.format(dimension), 'r') as file:
    data = json.load(file)
    for d in data:
        param.append(d[dimension])
        average_gap.append(d['{}-average-gap'.format(observe)])
        overrun_seconds.append(d['{}-overrun-seconds'.format(observe)])
        request_adjust.append(d['{}-request-adjust-times'.format(observe)])

fig, (ax1, ax2, ax3) = plt.subplots(3)

ax1.plot(param, average_gap)
ax1.set_title('{} Average Gap'.format(observe))
ax1.set_xlabel(dimension + " value")
ax1.set_ylabel('Resource')
ax1.grid(True)

ax2.plot(param, overrun_seconds)
ax2.set_title('{} Overrun Seconds'.format(observe))
ax2.set_xlabel(dimension + " value")
ax2.set_ylabel('Resource')
ax2.grid(True)

ax3.plot(param, request_adjust)
ax3.set_title('{} Request Adjust Times'.format(observe))
ax3.set_xlabel(dimension + " value")
ax3.set_ylabel('Resource')
ax3.grid(True)


# 调整子图布局
plt.tight_layout()

# 显示图表
plt.show()