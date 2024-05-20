import json

with open('param_sweep_result.json', 'r') as f:
    inputs = json.load(f)

filtered_inputs = []

for input in inputs:
    result = input['result']
    # if result['cpu-overrun-seconds'] > 20000:
    #     continue
    # if result['cpu-request-adjust-times'] > 600:
    #     continue

    if result['memory-overrun-seconds'] > 400:
        continue
    if result['memory-request-adjust-times'] > 100:
        continue
    if result['oom-seconds'] > 5:
        continue

    filtered_inputs.append(input)

sorted_inputs = sorted(filtered_inputs, key=lambda x: x['result']['memory-average-gap'])
i = 0
for s in sorted_inputs:
    print(s['result'])
    print('- ', s['args'])
    i += 1
    if i > 20:
        break