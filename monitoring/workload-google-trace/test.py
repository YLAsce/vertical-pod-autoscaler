import queue

class Struct:
    def __init__(self, a, b):
        self.a = a
        self.b = b
    
    def __lt__(self, other):
        if self.a != other.a:
            return self.a < other.a
        else:
            return self.b < other.b


# 创建优先队列
pq = queue.PriorityQueue()

# 定义一些结构体
structs = [
    Struct(1, 2),
    Struct(2, 3),
    Struct(1, 1),
    Struct(2, 1),
]

# 将结构体插入优先队列
for struct in structs:
    pq.put(struct)

# 从优先队列中弹出元素，并按照指定顺序打印
while not pq.empty():
    struct = pq.get()
    print(f"a: {struct.a}, b: {struct.b}")
