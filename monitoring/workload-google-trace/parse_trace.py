import json
import queue
import matplotlib.pyplot as plt

pq = queue.PriorityQueue()

class Struct:
    def __init__(self, a, b, cpu, mem):
        self.a = a
        self.b = b
        self.cpu = cpu
        self.mem = mem
    
    def __lt__(self, other):
        if self.a != other.a:
            return self.a < other.a
        else:
            return self.b < other.b
        
with open('131695484619_output.json_part-00000-a982216f-1ccb-4948-b7f7-a47b5f31edff-c000.json', 'r') as f:
    # 逐行读取文件内容
    for line in f:
        # 解析 JSON 字符串
        data = json.loads(line)
        
        s = Struct(data["start_time"], data["end_time"], data["cpus"], data["memory"])
        pq.put(s)

result = []
cur_start = cur_end = 0
cur_cpu = cur_mem = 0.0

while not pq.empty():
    s = pq.get()
    if s.a >= cur_end:
        if cur_end > cur_start:
            result.append([cur_start, cur_end, cur_cpu, cur_mem])
        cur_start = s.a
        cur_end = s.b
        cur_cpu = s.cpu
        cur_mem = s.mem
        continue

    # s.a < cur_end
    if s.a > cur_start:
        result.append([cur_start, s.a, cur_cpu, cur_mem])
        cur_start = s.a

    #s.a == cur_start
    if s.b == cur_end:
        cur_cpu += s.cpu
        cur_mem += s.mem
        continue

    if s.b < cur_end:
        ss = Struct(s.b, cur_end, cur_cpu, cur_mem)
        pq.put(ss)
        cur_end = s.b
        cur_cpu += s.cpu
        cur_mem += s.mem
        continue

    # s.b > cur_end
    ss = Struct(cur_end, s.b, s.cpu, s.mem)
    pq.put(ss)
    cur_cpu += s.cpu
    cur_mem += s.mem

with open('trace.data', 'w') as f:
    for r in result:
        f.write('{} {} {}\n'.format((r[1] - r[0]) // 1000000, r[2], r[3]) )


timestamps = []

cpu = []

memory = []

for r in result:
    timestamps.append( r[0]//1000000 )
    cpu.append( r[2] )
    memory.append( r[3] )

fig, (ax1, ax2) = plt.subplots(2)

# 在第一个子图上绘制 CPU 曲线
ax1.plot(timestamps, cpu, label='CPU Usage')

ax1.set_title('CPU Resource')
ax1.set_xlabel('Time (seconds)')
ax1.set_ylabel('Relative Resource')
ax1.grid(True)
# ax1.hlines(0.7, 0, timestamps[-1], colors='black', linestyles='dashed')
# ax1.hlines(0.1, 0, timestamps[-1], colors='black', linestyles='dashed')


# 在第二个子图上绘制内存曲线
ax2.plot(timestamps, memory, label='Memory Usage')
ax2.set_title('Memory Resource')
ax2.set_xlabel('Time (minutes)')
ax2.set_ylabel('Relative Resource')
ax2.grid(True)
# ax2.hlines(367001600, 0, timestamps[-1], colors='black', linestyles='dashed')
# ax2.hlines(52428800, 0, timestamps[-1], colors='black', linestyles='dashed')

# 调整子图布局
plt.tight_layout()

# 显示图表
plt.show()
