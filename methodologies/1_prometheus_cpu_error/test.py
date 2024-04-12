import pprint
import subprocess
import time
import json
import requests

prometheus_base = 'http://{}:9090/api/v1/query'.format("108.141.80.121")

def query_prometheus(query):
    """
    Execute promQL query for metrics on the cluster
    Return: JSON object
    """
    url = prometheus_base
    params = {
        'query': query
    }
    response = requests.get(url, params=params)
    data = response.json()
    return data

# tm = []
# rs = []
# tm1 = []
# rs1 = []

# tm2 = []
# rs2 = []
# for i in range(1000):
#     result = query_prometheus('rate(container_cpu_usage_seconds_total{container="workload",pod="workload-6f6646c47b-fswfr"}[60s])')
#     tm.append(int(result['data']['result'][0]['value'][0]) % 1000)
#     rs.append(float(result['data']['result'][0]['value'][1]))

#     result1 = query_prometheus('container_cpu_usage_seconds_total{container="workload",pod="workload-6f6646c47b-fswfr"}')
#     tm1.append(int(result1['data']['result'][0]['value'][0]) % 1000)
#     rs1.append(float(result1['data']['result'][0]['value'][1]))

#     command = 'kubectl get --raw /api/v1/nodes/aks-mlintra-35076703-vmss000000/proxy/metrics/cadvisor | grep "container_cpu_usage_seconds_total{container=\\"workload\\""'
#     process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

#     output, error = process.communicate()
#     o = output.decode("utf-8").strip()
#     arr = o.split(' ')
#     tm2.append((int(arr[-1])// 1000) % 1000)
#     rs2.append(float(arr[-2]))
#     time.sleep(0.2)
#     print(i)


# data = {
#     'tm': tm,
#     'rs': rs,
#     'tm1': tm1,
#     'rs1': rs1,
#     'tm2': tm2,
#     'rs2': rs2
# }

# with open('data.json', 'w') as f:
#     json.dump(data, f)

# tm2 = []
# rs2 = []
# for i in range(1000):
#     command = 'kubectl get --raw /api/v1/nodes/aks-mlintra-35076703-vmss000000/proxy/metrics/cadvisor | grep "container_cpu_usage_seconds_total{container=\\"workload\\""'
#     process = subprocess.Popen(command, shell=True, stdout=subprocess.PIPE, stderr=subprocess.PIPE)

#     output, error = process.communicate()
#     o = output.decode("utf-8").strip()
#     arr = o.split(' ')
#     tm2.append((int(arr[-1])// 1000))
#     rs2.append(float(arr[-2]))
#     print(i)

# result = query_prometheus('container_cpu_usage_seconds_total{container="workload",pod="workload-6f6646c47b-fswfr"}[1000s]')

# tm = []
# rs = []

# for v in result['data']['result'][0]['values']:
#     tm.append(int(v[0]))
#     rs.append(float(v[1]))

# data = {
#     'tm': tm, # prometheus
#     'rs': rs,
#     'tm2': tm2, # cadvisor
#     'rs2': rs2
# }

# with open('data_new.json', 'w') as f:
#     json.dump(data, f)


import matplotlib.pyplot as plt
import json
with open('data_new.json', 'r') as f:
    data = json.load(f)

tm = data['tm']
rs = data['rs']
# tm1 = data['tm1']
# rs1 = data['rs1']
tm2 = data['tm2']
rs2 = data['rs2']

s = [i for i in range(1000)]

fig, (ax1, ax2) = plt.subplots(2)
# ax1.plot(s, rs[:150])
# ax2.plot(s, rs1[:150])
# ax3.plot(s, rs2[:150])

ax1.plot(tm[-5:], rs[-5:], marker='o')
# ax2.plot(tm1[950:], rs1[950:])
# ax3.plot(tm2[950:], rs2[950:])

ax2.plot(tm2[-20:], rs2[-20:], marker='o')
# ax2.plot(s[900:], [rs1[i]-rs2[i] for i in range(900, 1000)])
# ax3.plot(s[900:], [tm2[i] for i in range(900, 1000)])
# ax4.plot(s[900:], [rs2[i] for i in range(900, 1000)])

plt.tight_layout()
plt.show()