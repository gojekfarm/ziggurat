telegraf:
  image: telegraf:latest
  container_name: {{.AppName}}_telegraf
  ports:
    - "8125:8125/udp"
    - "9273:9273"
  volumes:
    - ./telegraf.conf:/etc/telegraf/telegraf.conf:ro

grafana:
  image: grafana/grafana:latest
  container_name: {{.AppName}}_grafana
  ports:
    - "3000:3000"
  user: "0"
  links:
    - prometheus
  volumes:
    - ./data/grafana/data:/var/lib/grafana
    - ./grafana.ini:/etc/grafana/grafana.ini
    - ./datasource.yaml:/etc/grafana/provisioning/datasources/datasource.yaml

prometheus:
  image: prom/prometheus
  container_name: {{.AppName}}_prometheus
  ports:
    - "9090:9090"
  links:
    - telegraf
  volumes:
    - ./prom_conf.yaml:/etc/prometheus/prometheus.yml:ro
    - ./data/prometheus/data:/prometheus

