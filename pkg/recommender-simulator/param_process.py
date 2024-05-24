import json

with open('sweep/param_sweep_result_v2_memory_8.json', 'r') as f:
    inputs = json.load(f)

filtered_inputs = []

for input in inputs:
    result = input['result']
    # if result['cpu-overrun-seconds'] > 20000:
    #     continue
    # if result['cpu-request-adjust-times'] > 600:
    #     continue

    # if result['memory-overrun-seconds'] > 20000:
    #     continue
    # if result['memory-request-adjust-times'] > 2000:
    #     continue
    # if result['oom-seconds'] > 5:
    #     continue

    filtered_inputs.append(input)

sorted_inputs = sorted(filtered_inputs, key=lambda x: x['result']['memory-average-gap'])
i = 0
with open('sweep/param_sweep_first.data', 'w') as file:
    for s in sorted_inputs:
        file.write(str(s['result']) + '\n')
        file.write('-  ' + str(s['args']) + '\n')
        i += 1
        if i > 20:
            break