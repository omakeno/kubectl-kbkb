# kubectl kbkb plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/omakeno/kubectl-kbkb)](https://goreportcard.com/report/github.com/omakeno/kubectl-kbkb)

Display your pods on your nodes as kbkb format.
This only works on bash.

## Install

```bash
wget https://github.com/omakeno/kubectl-kbkb/releases/download/v0.2.1/kubectl-kbkb
chmod +x kubectl-kbkb
sudo cp kubectl-kbkb <your-path>
```

## Run

```bash
kubectl kbkb 

kubectl kbkb --watch

kubectl kbkb --namespace your-namespace

kubectl kbkb --kubeconfig your-kubeconfig
```
