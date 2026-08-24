# MedOps 医疗设备物联网远程运维平台

MedOps 是医疗设备物联网远程运维平台：监护仪、呼吸机、输液泵等设备注册后按周期
上报心跳，平台根据心跳窗口判定在线状态；质控计划按周期推进，质控测试结果落盘并
生成报告；性能参数趋势基于质控结果更新，保养计划按趋势与质控结果生成；设备状态
在在用、维护、停用间流转，全链路写入审计日志。

## 构建与运行

```bash
go build -mod=vendor ./...
go test -mod=vendor -count=1 ./...
go vet -mod=vendor ./...
```

启动服务：

```bash
go run ./cmd/medops -addr :8080 -data ./data
```

健康检查：`curl http://localhost:8080/healthz`

## Docker

```bash
bash build_benzhi_docker.sh medops linux/amd64
docker run --rm -p 8080:8080 medops bash -c 'go run ./cmd/medops -addr :8080 -data /tmp/data'
```
