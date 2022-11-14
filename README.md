# Alertmanger Controller Helm Repository

Add Alertmanager Controller repository to Helm repos:

```bash
helm repo add ekristen-amc https://ekristen.github.io/alertmanager-controller/charts/
```

## Install Alertmanager Controller

```bash
helm upgrade -i amc ekristen-amc/alertmanager-controller \
--namespace ekristen-amc
```
