import json
import math

start = 0
stop = 50 * math.pi
num_samples = 120


# x_values = [start + (stop - start) / (num_samples - 1) * i for i in range(num_samples)]
# y_values = [0.4+ 0.3*math.sin(x) for x in x_values]

y_values = []

for i in range(num_samples):
   if i % 60 < 30:
       y_values.append(0.4)
   else:
       y_values.append(0.8)

timestamp = [60 for _ in range(num_samples)]

with open('trace.data', 'w') as f:
    for i in range(num_samples):
        f.write('{} {} {}\n'.format(timestamp[i], y_values[i], int(2000000000 * (y_values[i]-0.35))) ) # 25MB to 400MB


