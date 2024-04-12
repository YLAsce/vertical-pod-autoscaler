from kubernetes import client, config
from kubernetes.utils import quantity
from azure.identity import DefaultAzureCredential
from azure.mgmt.containerservice import ContainerServiceClient
import requests
import pprint
import json
import time
import copy

subscription_id = "cbd332d2-cbb7-4189-bf84-48155e558134"
prometheus_addr = "108.141.80.121"
deployment_name = "workload"
namespace = "default"
frequency = 1 # 1min
data_points = 40

credentials = DefaultAzureCredential()
aks_client = ContainerServiceClient(credentials, subscription_id)

prometheus_base = 'http://{}:9090/api/v1/query'.format(prometheus_addr)

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


def get_deployment_container_resources(deployment_name, namespace):
    config.load_kube_config()
    apps_api = client.AppsV1Api()
    core_api = client.CoreV1Api()

    # 获取Deployment的所有Pod
    pods = core_api.list_namespaced_pod(namespace=namespace, label_selector=f"app={deployment_name}")

    container_resources = {}
    for pod in pods.items:
        pod_name = pod.metadata.name
        for container in pod.spec.containers:
            container_resources.setdefault(pod_name, {})[container.name] = {
                "requests": container.resources.requests,
                "limits": container.resources.limits
            }
    return container_resources



def get_usage(pod, container, frequency):
    query_str_cpu = 'irate(container_cpu_usage_seconds_total{{pod="{}", container="{}"}}[{}m])'.format(pod, container, frequency)
    result_cpu = query_prometheus(query_str_cpu)

    # container_memory_usage_bytes is larger than this... working set is used by recommender
    query_str_memory = 'avg_over_time(container_memory_working_set_bytes{{pod="{}", container="{}"}}[{}m])'.format(pod, container, frequency)
    result_memory = query_prometheus(query_str_memory)
    if result_cpu['status'] != 'success' or result_memory['status'] != 'success':
        raise Exception("get usage query error!! not success")
    
    if len(result_cpu['data']['result']) != 1 or len(result_memory['data']['result']) != 1:
        raise Exception("get usage query error!! not unique container name result\n" + pprint.pformat(result_cpu) + "\n" + pprint.pformat(result_memory))
    
    return result_cpu['data']['result'][0]['value'][1], result_memory['data']['result'][0]['value'][1]


def collect_request_usage(deployment_name, namespace, frequency):
    resources_request = get_deployment_container_resources(deployment_name, namespace)
    if not resources_request:
        raise Exception("Empty resource request")
    
    request_usage_info = {}

    for pod, podinfo in resources_request.items():
        pod_request_usage_info = {}
        for container, containerinfo in podinfo.items():
            cpu_usage, memory_usage = get_usage(pod, container, frequency)

            container_request_usage_info = {}
            container_request_usage_info['cpu'] = {}
            container_request_usage_info['memory'] = {}
            container_request_usage_info['cpu']['limit'] = float(quantity.parse_quantity(containerinfo['limits']['cpu']))
            container_request_usage_info['cpu']['request'] = float(quantity.parse_quantity(containerinfo['requests']['cpu']))
            container_request_usage_info['cpu']['usage'] = float(cpu_usage)

            container_request_usage_info['memory']['limit'] = float(quantity.parse_quantity(containerinfo['limits']['memory']))
            container_request_usage_info['memory']['request'] = float(quantity.parse_quantity(containerinfo['requests']['memory']))
            container_request_usage_info['memory']['usage'] = float(memory_usage)

            pod_request_usage_info[container] = container_request_usage_info
        request_usage_info[pod] = pod_request_usage_info
    
    return request_usage_info

def make_zero_data(prev_data):
    cur_data = copy.deepcopy(prev_data)

    for pod, podinfo in cur_data.items():
        for container, containerinfo in podinfo.items():

            containerinfo['cpu']['limit'] = 0.0
            containerinfo['cpu']['request'] = 0.0
            containerinfo['cpu']['usage'] = 0.0

            containerinfo['memory']['limit'] = 0.0
            containerinfo['memory']['request'] = 0.0
            containerinfo['memory']['usage'] = 0.0

    return cur_data

# MAIN
arr_data_request_usage = []
for i in range(data_points):
    print("-------start round------", str(i+1))
    try:
        data_request_usage = collect_request_usage(deployment_name, namespace, frequency)
        pprint.pprint(data_request_usage)
        arr_data_request_usage.append(data_request_usage)
    except Exception as e:
        arr_data_request_usage.append(make_zero_data(arr_data_request_usage[-1]))
        print("Exception at round {}, put data 0: {}".format(i+1, e))

    print("-------finish record data------", str(i+1))

    if i != data_points-1:
        time.sleep(60*frequency)


json_data = json.dumps(arr_data_request_usage)

with open('metrics_{}_{}.json'.format(frequency, data_points), 'w') as file:
    file.write(json_data)