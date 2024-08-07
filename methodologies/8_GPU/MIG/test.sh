#!/bin/bash
kubectl patch clusterpolicies.nvidia.com/cluster-policy -n kube-system --type='json' -p='[{"op": "replace", "path": "/spec/mig/strategy", "value": "single"}]'
kubectl label nodes scw-k8s-mlintra-pool-gpu-4ea7b7901a8c44d0b1646 nvidia.com/mig.config=all-1g.10gb --overwrite
# sleep 2
kubectl apply -f example.yaml

#37s from 1g.10GB to 1g.20GB
#38s from 1g.20GB to 2g.20GB
#39s from 2g.20GB to 3g.40GB
#38s from 3g.40GB to 4g.40GB
#38s from 4g.40GB to 7g.80GB. The status may not immediately goes to Pending at first, still success of last state, be careful!!
#39s from 7g.80GB to 1g.10GB
#39s from 1g.20GB to 1g.10GB
#36s from 1g.20GB to all balanced
#33s from all balanced to 1g.10GB
#34s from 7g.80GB to all balanced
#35s

#pending time: 44s + 38s=82s 在update之后立即启动
#08.15-09.50=1min35s=95s, 在update同时启动
#15:00-16:50=1min50s=110s,在update同时2s后启动
#23:33-24:49=1min16s=76s,在update之后15s启动
#29:30-30:51=1min21s=81s,在update之后25s启动


#33:39-34:49=1min10s-70s,mixed mode
#36:26-37:49=1min23s=83s, mixed mode to single mode
#21-51, 30s