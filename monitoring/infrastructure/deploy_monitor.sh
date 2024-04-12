kubectl apply -f namespace.yaml
kubectl apply -f node-exporter.yaml
kubectl apply -f prometheus.yaml 
# kubectl apply -f grafana.yaml
kubectl get pod -n monitor
kubectl get svc -n monitor