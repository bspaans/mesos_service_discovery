package main

import (
  "log"
  "os"
  "io/ioutil"
  "os/exec"
  "strings"
  "text/template"
)

var haproxyTemplate = `
global
  daemon
  log 127.0.0.1 local0
  log 127.0.0.1 local1 notice
  maxconn 4096

defaults
  log         global
  retries     3
  maxconn     2000
  contimeout  5000
  clitimeout  50000
  srvtimeout  50000

listen stats
  bind 127.0.0.1:9090
  balance
  mode http
  stats enable
  stats auth admin:admin

{{ range $appId, $app := . }}
{{ if appExposesPorts $app }}
listen {{ sanitizeApplicationId $appId }}
  bind 0.0.0.0:{{ port $app }}
  mode tcp
  option tcplog
  balance leastconn
  {{ range $taskId, $task := $app.ApplicationInstances }}
  server {{$taskId}} {{$task.Host}}:{{port $app}} check
  {{ end }}
{{ end }}
{{ end }}
`

func appExposesPorts (app Application) bool {
  return len(app.Ports) != 0;
}

func sanitizeApplicationId(appId string) string {
  return strings.Replace(appId, "/", "_", -1)
}

func getApplicationPort(app Application) int {
  return app.Ports[0]
}

func updateHAProxyConfig(applicationMap map[string]Application) {
  tmp, err := ioutil.TempFile("", "haproxy.cfg")
  if err != nil {
    return
  }
  generateHAProxyConfig(tmp, applicationMap)
  replaceHAProxyConfiguration(tmp.Name())
  reloadHAProxy()
}

func generateHAProxyConfig(tmp *os.File, applicationMap map[string]Application) {
  funcMap := template.FuncMap {
    "appExposesPorts": appExposesPorts,
    "sanitizeApplicationId": sanitizeApplicationId,
    "port": getApplicationPort,
  }
  tpl, err := template.New("haproxy").Funcs(funcMap).Parse(haproxyTemplate);
  if err != nil { panic(err); }
  err = tpl.Execute(tmp, applicationMap);
  if err != nil { panic(err); }
}

func replaceHAProxyConfiguration(tmpFile string) {
  err := os.Rename(tmpFile, "/etc/haproxy/haproxy.cfg")
  if err != nil {
    log.Printf("ERR Couldn't write /etc/haproxy/haproxy.cfg: %s", err)
    return
  }
  log.Println("INFO Written new /etc/haproxy/haproxy.cfg")
}

func reloadHAProxy() {
  cmd := exec.Command("service", "haproxy", "reload")
  err := cmd.Start()
  if err != nil {
    log.Println("ERR failed to reload HAProxy")
    return
  }
  err = cmd.Wait()
  if err != nil {
    log.Println("ERR failed to reload HAProxy")
    return
  }
}
