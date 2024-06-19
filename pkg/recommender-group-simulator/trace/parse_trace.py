import random
import json

input_file_path = 'trace.data'
init_file_path = 'init.json'

scale_daemons_cpu = 0.4
scale_daemons_mem = 0.4
num_daemons = 5
scale_serverless_cpu = 0.6
scale_serverless_mem = 0.6
num_serverless = 100

init_map = {}
sum_time = []
with open(input_file_path, 'r') as input_file, open(init_file_path, 'w') as init_file:
    lines = input_file.readlines()
    for l in lines:
        if len(sum_time) == 0:
            sum_time.append( int(l.split(' ')[0]) )
        else:
            sum_time.append( sum_time[-1] + int(l.split(' ')[0]) )

    print(sum_time[-1])

    for i in range(num_daemons):
        start = random.randint(0, len(lines)-1)
        print("generate daemon", i, start)
        spl0 = lines[start].split(' ')
        init_map["daemon_" + str(i)] = {
            "cpu": float(spl0[1])*scale_daemons_cpu*1.06,
            "memory": int(float(spl0[2])*scale_daemons_mem*1.06) 
        }
        with open('0.1.'+str(i), 'w') as output_file:
            cur_seconds = 0
            for j in range(start, len(lines)):
                spl = lines[j].split(' ')
                output_file.write('{} {} {} {}\n'.format(cur_seconds, int(spl[0]), float(spl[1])*scale_daemons_cpu, int(float(spl[2])*scale_daemons_mem) ))
                cur_seconds += int(spl[0])

            for j in range(0, start):
                spl = lines[j].split(' ')
                output_file.write('{} {} {} {}\n'.format(cur_seconds, int(spl[0]), float(spl[1])*scale_daemons_cpu, int(float(spl[2])*scale_daemons_mem) ))
                cur_seconds += int(spl[0])

    for i in range(num_serverless):
        start = random.randint(1, len(lines)-100)
        end = random.randint(start+100, min(start + 100 + len(lines)//2, len(lines)))
        print("generate serverless", i, start, end)
        spl0 = lines[start].split(' ')
        init_map["serverless_" + str(i)] = {
            "cpu": float(spl0[1])*scale_serverless_cpu*1.06,
            "memory": int(float(spl0[2])*scale_serverless_mem*1.06) 
        }
        with open('0.0.'+str(i), 'w') as output_file:
            cur_seconds = random.randint(0, sum_time[-1] - (sum_time[end] - sum_time[start-1]))
            for j in range(start, end):
                    spl = lines[j].split(' ')
                    output_file.write('{} {} {} {}\n'.format(cur_seconds, int(spl[0]), float(spl[1])*scale_serverless_cpu, int(float(spl[2])*scale_serverless_mem) ))
                    cur_seconds += int(spl[0])

    json.dump(init_map, init_file, indent=4)
