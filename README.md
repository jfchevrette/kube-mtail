# kube-mtail

*Work-in-progress*

kube-mtail is a tool used to retrieve logs from kubernetes pods and process them with [mtail](https://github.com/google/mtail). mtail exposes metrics based on the mtail programs

# Instructions

## Installing

user    go get github.com/jfchevrette/kube-mtail

    kube-mtail -h
      -kubeconfig string
            (optional) absolute path to the kubeconfig file (default "/home/jfchevrette/.kube/config")
      -namespace string
            (required) namespace to target
      -selector string
            (optional) Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. key1=value1,key2=value2

## Example usage

    # Run mtail and watch
    mtail -progs mtail-programs -logs logs/*.log -logtostderr

    # Run kube-mtail and output logs to the logs directory
    kube-mtail -namespace monitoring -selector app=prometheus | tee -a logs/prometheus.log


## Configuration

`kube-mtail` will attempt to find a kubeconfig file at `~/.kube/config` and use it's current context. A different kubeconfig file can be passed with the `-kubeconfig` argument. If no kubeconfig can be found, `kube-mtail` will attempt to connect using the in-cluster strategy, expecting a serviceaccount token and certificate to be available under `/var/run/secrets/kubernetes.io/serviceaccount/` as well as using the `KUBERNETES_SERVICE_HOST` and `KUBERNETES_SERVICE_PORT` environment variables which are available by default in every kubernetes pod.

# TODO
- Have kube-mtail launch and manage the mtail process
- kube-mtail should append metadata to log lines so that mtail can pick them up and use them (namespace, pod name, etc..)
- Docs on how to integrate with prometheus

